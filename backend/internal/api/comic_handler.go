package api

import (
	"comic-viewer-claude/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ComicHandler 漫画处理器
type ComicHandler struct {
	comicService *service.ComicService
}

// NewComicHandler 创建漫画处理器
func NewComicHandler(comicService *service.ComicService) *ComicHandler {
	return &ComicHandler{
		comicService: comicService,
	}
}

// GetList 获取漫画列表
// GET /api/comics?page=1&page_size=20&use_cache=true
func (h *ComicHandler) GetList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	useCache := c.DefaultQuery("use_cache", "true") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	resp, err := h.comicService.GetList(page, pageSize, useCache)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDetail 获取漫画详情
// GET /api/comics/:id
func (h *ComicHandler) GetDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的漫画ID"})
		return
	}

	detail, err := h.comicService.GetDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// GetReaderImages 获取阅读器图片列表
// GET /api/comics/:id/images
func (h *ComicHandler) GetReaderImages(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的漫画ID"})
		return
	}

	images, err := h.comicService.GetReaderImages(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"images": images})
}

// Search 搜索漫画
// GET /api/comics/search?keyword=xxx&page=1&mode=local&page_size=100
func (h *ComicHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	mode := c.DefaultQuery("mode", "local")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 200 {
		pageSize = 100
	}

	if mode == "local" {
		resp, err := h.comicService.SearchLocal(keyword, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	} else {
		// TODO: 在线搜索
		c.JSON(http.StatusNotImplemented, gin.H{"error": "在线搜索暂未实现"})
	}
}

// AddReadingHistory 添加阅读历史
// POST /api/comics/:id/history
func (h *ComicHandler) AddReadingHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的漫画ID"})
		return
	}

	var req struct {
		Title    string `json:"title"`
		CoverURL string `json:"cover_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.comicService.AddReadingHistory(id, req.Title, req.CoverURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "添加成功"})
}

// ToggleFavorite 切换收藏状态
// POST /api/comics/:id/favorite
func (h *ComicHandler) ToggleFavorite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的漫画ID"})
		return
	}

	var req struct {
		Title    string `json:"title"`
		CoverURL string `json:"cover_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isFavorited, err := h.comicService.ToggleFavorite(id, req.Title, req.CoverURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_favorited": isFavorited,
		"message":      map[bool]string{true: "已收藏", false: "已取消收藏"}[isFavorited],
	})
}

// GetDashboardStats 获取仪表盘统计
// GET /api/dashboard/stats
func (h *ComicHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.comicService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Filter 筛选漫画（按作者/标签从网站获取）
// GET /api/comics/filter?tag_id=1&author_id=1&page=1
func (h *ComicHandler) Filter(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	var tagID, authorID *int
	if v := c.Query("tag_id"); v != "" {
		id, _ := strconv.Atoi(v)
		tagID = &id
	}
	if v := c.Query("author_id"); v != "" {
		id, _ := strconv.Atoi(v)
		authorID = &id
	}

	resp, err := h.comicService.GetFilteredList(page, tagID, authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetFavorites 获取收藏列表
// GET /api/favorites?page=1&page_size=20
func (h *ComicHandler) GetFavorites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	favorites, total, err := h.comicService.GetFavorites(pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":        favorites,
		"total":        total,
		"current_page": page,
		"total_pages":  (int(total) + pageSize - 1) / pageSize,
	})
}

// GetHistory 获取历史记录
// GET /api/history?page=1&page_size=20
func (h *ComicHandler) GetHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	history, total, err := h.comicService.GetHistory(pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":        history,
		"total":        total,
		"current_page": page,
		"total_pages":  (int(total) + pageSize - 1) / pageSize,
	})
}
