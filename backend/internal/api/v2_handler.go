package api

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StrategyReloader 缓存策略重载接口
type StrategyReloader interface {
	ReloadConfig()
}

// SettingsInvalidator 设置缓存失效接口
type SettingsInvalidator interface {
	InvalidateProxySettingsCache()
}

// V2Handler v2 API 处理器
type V2Handler struct {
	searchService    *service.SearchService
	settingsService  *service.SettingsService
	comicService     *service.ComicService
	repo             *repository.Repository
	strategyReloader StrategyReloader
	settingsHooks    []SettingsInvalidator
}

// NewV2Handler 创建 V2Handler
func NewV2Handler(
	searchService *service.SearchService,
	settingsService *service.SettingsService,
	comicService *service.ComicService,
	repo *repository.Repository,
) *V2Handler {
	return &V2Handler{
		searchService:   searchService,
		settingsService: settingsService,
		comicService:    comicService,
		repo:            repo,
	}
}

// SetStrategyReloader 注入缓存策略重载器
func (h *V2Handler) SetStrategyReloader(r StrategyReloader) {
	h.strategyReloader = r
}

// AddSettingsInvalidator 注入设置缓存失效器
func (h *V2Handler) AddSettingsInvalidator(invalidator SettingsInvalidator) {
	if invalidator == nil {
		return
	}
	h.settingsHooks = append(h.settingsHooks, invalidator)
}

// Search 搜索+记录历史
// GET /api/v2/search?keyword=xxx&page=1&page_size=20
func (h *V2Handler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		Error(c, http.StatusBadRequest, CodeBadRequest, "keyword 不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	resp, err := h.searchService.SearchWithHistory(1, keyword, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	PagedSuccess(c, resp.Items, resp.CurrentPage, resp.TotalPages, resp.Total)
}

// GetSearchHistory 获取搜索历史
// GET /api/v2/search/history?page=1&page_size=20
func (h *V2Handler) GetSearchHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	histories, total, err := h.searchService.GetSearchHistory(1, pageSize, offset)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	PagedSuccess(c, histories, page, totalPages, total)
}

// ClearSearchHistory 清空搜索历史
// DELETE /api/v2/search/history
func (h *V2Handler) ClearSearchHistory(c *gin.Context) {
	if err := h.searchService.ClearSearchHistory(1); err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}
	Success(c, nil)
}

// GetSettings 获取用户设置
// GET /api/v2/settings
func (h *V2Handler) GetSettings(c *gin.Context) {
	settings, err := h.settingsService.GetAllSettings(1)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	// 转换为 map
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	Success(c, result)
}

// UpdateSettings 更新设置
// PUT /api/v2/settings
func (h *V2Handler) UpdateSettings(c *gin.Context) {
	var req model.SettingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	if err := h.settingsService.SetBatchSettings(1, req.Settings); err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	// 通知缓存策略调度器重新加载配置并立即执行
	if h.strategyReloader != nil {
		go h.strategyReloader.ReloadConfig()
	}
	for _, hook := range h.settingsHooks {
		hook.InvalidateProxySettingsCache()
	}

	Success(c, nil)
}

// GetCacheStatus 缓存概览统计
// GET /api/v2/cache/status
func (h *V2Handler) GetCacheStatus(c *gin.Context) {
	total, l1, l2, l3, err := h.repo.GetCacheOverview()
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	Success(c, model.CacheOverview{
		TotalComics: total,
		L1Cached:    l1,
		L2Cached:    l2,
		L3Cached:    l3,
	})
}

// GetHistoryWithProgress 带阅读进度的历史
// GET /api/v2/history?page=1&page_size=20
func (h *V2Handler) GetHistoryWithProgress(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	histories, total, err := h.comicService.GetHistory(pageSize, offset)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	items := make([]model.HistoryWithProgress, 0, len(histories))
	for _, hist := range histories {
		item := model.HistoryWithProgress{
			ComicID:    hist.ComicID,
			Title:      hist.Title,
			CoverURL:   hist.CoverURL,
			LastReadAt: hist.LastReadAt.Format("2006-01-02 15:04:05"),
			PageNumber: hist.PageNumber,
			Progress:   hist.Progress,
		}

		// 附加 reading_progress
		if rp, err := h.repo.GetReadingProgress(hist.ComicID, 1); err == nil {
			item.LastPage = rp.LastPage
			item.TotalPages = rp.TotalPages
		}

		items = append(items, item)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	PagedSuccess(c, items, page, totalPages, total)
}

// UpdateProgress 更新阅读进度
// PUT /api/v2/history/:id/progress
func (h *V2Handler) UpdateProgress(c *gin.Context) {
	comicID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的漫画ID")
		return
	}

	var req model.ProgressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	progress := &model.ReadingProgress{
		ComicID:    comicID,
		UserID:     1,
		LastPage:   req.LastPage,
		TotalPages: req.TotalPages,
		Progress:   req.Progress,
		ChapterID:  req.ChapterID,
	}

	if err := h.repo.UpsertReadingProgress(progress); err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	Success(c, nil)
}

// GetOfflineComics 离线可用漫画列表
// GET /api/v2/offline/comics?page=1&page_size=20
func (h *V2Handler) GetOfflineComics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	states, total, err := h.repo.GetOfflineComics(pageSize, offset)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	PagedSuccess(c, states, page, totalPages, total)
}

// GetComicCacheState 单漫画缓存状态
// GET /api/v2/comics/:id/cache
func (h *V2Handler) GetComicCacheState(c *gin.Context) {
	comicID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的漫画ID")
		return
	}

	state, err := h.repo.GetCacheState(comicID)
	if err != nil {
		Error(c, http.StatusNotFound, CodeNotFound, "缓存状态不存在")
		return
	}

	Success(c, state)
}
