package service

import (
	"comic-viewer-claude/internal/cache"
	"comic-viewer-claude/internal/crawler"
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/parser"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// ComicService 漫画服务
type ComicService struct {
	repo       *repository.Repository
	crawler    *crawler.Client
	coverCache *cache.CoverCacheManager
}

// NewComicService 创建漫画服务
func NewComicService(repo *repository.Repository, crawler *crawler.Client, coverCache *cache.CoverCacheManager) *ComicService {
	return &ComicService{
		repo:       repo,
		crawler:    crawler,
		coverCache: coverCache,
	}
}

const sitePageSize = 100

// GetList 获取漫画列表（支持自定义pageSize，内部跨网站页拼接）
func (s *ComicService) GetList(page, pageSize int, useCache bool) (*model.ListResponse, error) {
	// 读取精确总漫画数
	totalComics := int64(0)
	if val, err := s.repo.GetSystemMetadata("total_comics_count"); err == nil {
		if tc, err := strconv.ParseInt(val, 10, 64); err == nil && tc > 0 {
			totalComics = tc
		}
	}

	// 计算前端请求的数据范围（0-based）
	startItem := (page - 1) * pageSize
	endItem := startItem + pageSize // exclusive

	// 计算需要哪些网站页
	startSitePage := startItem/sitePageSize + 1
	endSitePage := (endItem-1)/sitePageSize + 1

	// 获取所有需要的网站页数据
	var allItems []model.ComicListItem
	fromCache := true
	for sp := startSitePage; sp <= endSitePage; sp++ {
		items, cached, err := s.fetchSitePage(sp, useCache)
		if err != nil {
			return nil, err
		}
		if !cached {
			fromCache = false
		}
		allItems = append(allItems, items...)
	}

	// 在拼接的数据中切片
	baseOffset := (startSitePage - 1) * sitePageSize
	sliceStart := startItem - baseOffset
	sliceEnd := endItem - baseOffset
	if sliceEnd > len(allItems) {
		sliceEnd = len(allItems)
	}
	if sliceStart > len(allItems) {
		sliceStart = len(allItems)
	}
	displayItems := allItems[sliceStart:sliceEnd]

	// 计算总页数
	totalPages := 0
	if totalComics > 0 {
		totalPages = int((totalComics + int64(pageSize) - 1) / int64(pageSize))
	}

	return &model.ListResponse{
		Items:       displayItems,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		Total:       totalComics,
		FromCache:   fromCache,
	}, nil
}

// fetchSitePage 获取网站某一页的数据（优先缓存）
func (s *ComicService) fetchSitePage(sitePage int, useCache bool) ([]model.ComicListItem, bool, error) {
	if useCache {
		cachedItems, err := s.repo.GetCachedList(sitePage)
		if err == nil && len(cachedItems) > 0 {
			items := make([]model.ComicListItem, len(cachedItems))
			for i, cached := range cachedItems {
				items[i] = model.ComicListItem{
					ID:          cached.ComicID,
					Title:       cached.Title,
					Author:      cached.Author,
					AuthorID:    cached.AuthorID,
					CoverURL:    cached.CoverURL,
					Rating:      cached.Rating,
					RatingCount: cached.RatingCount,
					Favorites:   cached.Favorites,
				}
				if comic, err := s.repo.GetComic(cached.ComicID); err == nil && comic.HasCoverCached {
					items[i].CoverBase64 = comic.CoverBase64
				}
			}
			return items, true, nil
		}
	}

	// 从网络加载
	logger.Info("从网络加载列表", zap.Int("sitePage", sitePage))
	html, err := s.crawler.GetListPage(sitePage, nil, nil, nil, "")
	if err != nil {
		return nil, false, fmt.Errorf("获取列表页失败: %w", err)
	}

	items, _, err := parser.ParseListPage(html)
	if err != nil {
		return nil, false, fmt.Errorf("解析列表页失败: %w", err)
	}

	// 保存到缓存
	go s.saveListCache(sitePage, items)

	// 加载封面缓存
	for i := range items {
		if comic, err := s.repo.GetComic(items[i].ID); err == nil && comic.HasCoverCached {
			items[i].CoverBase64 = comic.CoverBase64
		}
	}

	return items, false, nil
}

// GetFilteredList 按标签/分类/作者筛选漫画（从网站获取）
func (s *ComicService) GetFilteredList(page int, tagID, categoryID, authorID *int, author string) (*model.ListResponse, error) {
	logger.Info("从网站筛选漫画", zap.Int("page", page))
	html, err := s.crawler.GetListPage(page, tagID, categoryID, authorID, author)
	if err != nil {
		return nil, fmt.Errorf("获取筛选页失败: %w", err)
	}

	items, totalPages, err := parser.ParseListPage(html)
	if err != nil {
		return nil, fmt.Errorf("解析筛选页失败: %w", err)
	}

	// 筛选结果同样属于已获取的列表元数据，应记作 L1。
	go s.markListItemsL1(items)

	// 加载封面缓存
	for i := range items {
		if comic, err := s.repo.GetComic(items[i].ID); err == nil && comic.HasCoverCached {
			items[i].CoverBase64 = comic.CoverBase64
		}
	}

	return &model.ListResponse{
		Items:       items,
		CurrentPage: page,
		TotalPages:  totalPages,
		Total:       int64(totalPages * len(items)),
		FromCache:   false,
	}, nil
}

func (s *ComicService) markListItemsL1(items []model.ComicListItem) {
	for _, item := range items {
		if err := s.repo.UpsertCacheStateL1(item.ID); err != nil {
			logger.Warn("更新筛选结果 cache_state L1 失败", zap.Int("comic_id", item.ID), zap.Error(err))
		}
	}
}

// saveListCache 保存列表缓存
func (s *ComicService) saveListCache(page int, items []model.ComicListItem) {
	cacheItems := make([]model.ComicListCache, len(items))
	now := time.Now()

	for i, item := range items {
		cacheItems[i] = model.ComicListCache{
			ComicID:     item.ID,
			Page:        page,
			Position:    i,
			Title:       item.Title,
			Author:      item.Author,
			AuthorID:    item.AuthorID,
			CoverURL:    item.CoverURL,
			Rating:      item.Rating,
			RatingCount: item.RatingCount,
			Favorites:   item.Favorites,
			CachedAt:    now,
		}
	}

	if err := s.repo.SaveListCache(cacheItems); err != nil {
		logger.Error("保存列表缓存失败", zap.Error(err))
	}

	s.markListItemsL1(items)
}

// GetDetail 获取漫画详情
func (s *ComicService) GetDetail(comicID int) (*model.ComicDetail, error) {
	// 先尝试从数据库加载
	comic, err := s.repo.GetComic(comicID)
	if err == nil && comic.Title != "" {
		// 从数据库构建详情
		detail := &model.ComicDetail{
			ID:           comic.ID,
			Title:        comic.Title,
			Subtitle:     comic.Subtitle,
			Author:       comic.Author,
			AuthorID:     comic.AuthorID,
			CoverURL:     comic.CoverURL,
			Rating:       comic.Rating,
			RatingCount:  comic.RatingCount,
			Favorites:    comic.Favorites,
			CategoryID:   comic.CategoryID,
			CategoryName: comic.CategoryName,
			CreatedAt:    comic.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    comic.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// 加载作者
		authors, _ := s.repo.GetComicAuthors(comicID)
		for _, author := range authors {
			detail.Authors = append(detail.Authors, model.AuthorInfo{
				AuthorID:   author.AuthorID,
				AuthorName: author.AuthorName,
			})
		}

		// 加载标签
		tags, _ := s.repo.GetComicTags(comicID)
		for _, tag := range tags {
			detail.Tags = append(detail.Tags, model.TagInfo{
				TagID:   tag.ID,
				TagName: tag.Name,
			})
		}

		// 检查是否收藏
		detail.IsFavorited, _ = s.repo.IsFavorited(comicID)

		logger.Info("从数据库加载详情", zap.Int("comic_id", comicID))
		return detail, nil
	}

	// 从网络加载
	logger.Info("从网络加载详情", zap.Int("comic_id", comicID))
	html, err := s.crawler.GetDetailPage(comicID)
	if err != nil {
		return nil, fmt.Errorf("获取详情页失败: %w", err)
	}

	detail, err := parser.ParseDetailPage(html, comicID)
	if err != nil {
		return nil, fmt.Errorf("解析详情页失败: %w", err)
	}

	// 保存到数据库
	go s.saveComicDetail(detail)

	// 检查是否收藏
	detail.IsFavorited, _ = s.repo.IsFavorited(comicID)

	return detail, nil
}

// saveComicDetail 保存漫画详情
func (s *ComicService) saveComicDetail(detail *model.ComicDetail) {
	now := time.Now()
	comic := &model.Comic{
		ID:           detail.ID,
		Title:        detail.Title,
		Subtitle:     detail.Subtitle,
		Author:       detail.Author,
		AuthorID:     detail.AuthorID,
		CoverURL:     detail.CoverURL,
		Rating:       detail.Rating,
		RatingCount:  detail.RatingCount,
		Favorites:    detail.Favorites,
		CategoryID:   detail.CategoryID,
		CategoryName: detail.CategoryName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.SaveComic(comic); err != nil {
		logger.Error("保存漫画失败", zap.Error(err))
		return
	}

	// 保存作者
	if len(detail.Authors) > 0 {
		authors := make([]model.ComicAuthor, len(detail.Authors))
		for i, author := range detail.Authors {
			authors[i] = model.ComicAuthor{
				ComicID:    detail.ID,
				AuthorID:   author.AuthorID,
				AuthorName: author.AuthorName,
				Position:   i,
			}
		}
		if err := s.repo.SaveComicAuthors(authors); err != nil {
			logger.Error("保存作者失败", zap.Error(err))
		}
	}

	// 更新 cache_states L2
	if err := s.repo.UpsertCacheStateL2(detail.ID); err != nil {
		logger.Warn("更新 cache_state L2 失败", zap.Int("comic_id", detail.ID), zap.Error(err))
	}

	// 保存标签
	if len(detail.Tags) > 0 {
		tagIDs := make([]int, 0, len(detail.Tags))
		for _, tag := range detail.Tags {
			// 先保存标签
			if err := s.repo.SaveTag(&model.Tag{ID: tag.TagID, Name: tag.TagName}); err != nil {
				logger.Warn("保存标签失败", zap.Error(err))
			}
			tagIDs = append(tagIDs, tag.TagID)
		}
		if err := s.repo.SaveComicTags(detail.ID, tagIDs); err != nil {
			logger.Error("保存漫画标签关联失败", zap.Error(err))
		}
	}
}

// GetReaderImages 获取阅读器图片列表
func (s *ComicService) GetReaderImages(comicID int) ([]model.ImageInfo, error) {
	// 先尝试从数据库加载
	images, err := s.repo.GetComicImages(comicID)
	if err == nil && len(images) > 0 {
		logger.Info("从数据库加载图片列表", zap.Int("comic_id", comicID), zap.Int("count", len(images)))
		result := make([]model.ImageInfo, len(images))
		for i, img := range images {
			result[i] = model.ImageInfo{
				Sort:      img.Sort,
				ComicID:   img.ComicID,
				Filename:  img.Filename,
				Extension: img.Extension,
				URL:       img.URL,
				LocalPath: img.LocalPath,
			}
		}
		return result, nil
	}

	// 从网络加载
	logger.Info("从网络加载图片列表", zap.Int("comic_id", comicID))
	html, err := s.crawler.GetReaderPage(comicID)
	if err != nil {
		return nil, fmt.Errorf("获取阅读页失败: %w", err)
	}

	_, images2, err := parser.ParseReaderPage(html, comicID)
	if err != nil {
		return nil, fmt.Errorf("解析阅读页失败: %w", err)
	}

	// 修正 sort=0 的情况，用索引代替
	for i := range images2 {
		if images2[i].Sort == 0 {
			images2[i].Sort = i + 1
		}
	}

	// 保存到数据库。这里需要同步落库，避免阅读页刚返回后立即创建下载任务时读不到图片列表。
	s.saveComicImages(comicID, images2)

	return images2, nil
}

// saveComicImages 保存漫画图片
func (s *ComicService) saveComicImages(comicID int, images []model.ImageInfo) {
	dbImages := make([]model.ComicImage, len(images))
	for i, img := range images {
		sort := img.Sort
		if sort == 0 {
			sort = i + 1 // 原始 sort 为 0 时用索引代替，避免主键冲突
		}
		dbImages[i] = model.ComicImage{
			ComicID:   img.ComicID,
			Sort:      sort,
			Filename:  img.Filename,
			Extension: img.Extension,
			URL:       img.URL,
		}
	}

	if err := s.repo.SaveComicImages(dbImages); err != nil {
		logger.Error("保存图片列表失败", zap.Error(err))
		return
	}

	// 更新 cache_states L2（图片URL列表已缓存，非文件下载）
	if err := s.repo.UpsertCacheStateL2(comicID); err != nil {
		logger.Warn("更新 cache_state L2 失败", zap.Int("comic_id", comicID), zap.Error(err))
	}
}

// SearchLocal 本地搜索
func (s *ComicService) SearchLocal(keyword string, page int, pageSize int) (*model.ListResponse, error) {
	offset := (page - 1) * pageSize
	comics, total, err := s.repo.SearchComicsLocal(keyword, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("本地搜索失败: %w", err)
	}

	items := make([]model.ComicListItem, len(comics))
	for i, comic := range comics {
		items[i] = model.ComicListItem{
			ID:          comic.ID,
			Title:       comic.Title,
			Author:      comic.Author,
			AuthorID:    comic.AuthorID,
			CoverURL:    comic.CoverURL,
			CoverBase64: comic.CoverBase64,
			Rating:      comic.Rating,
			RatingCount: comic.RatingCount,
			Favorites:   comic.Favorites,
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &model.ListResponse{
		Items:       items,
		CurrentPage: page,
		TotalPages:  totalPages,
		Total:       total,
		FromCache:   true,
	}, nil
}

// AddReadingHistory 添加阅读历史
func (s *ComicService) AddReadingHistory(comicID int, title, coverURL string) error {
	if err := s.repo.AddReadingHistory(comicID, title, coverURL); err != nil {
		return err
	}

	// 同步更新 reading_progress
	go func() {
		rp := &model.ReadingProgress{
			ComicID: comicID,
			UserID:  1,
		}
		if err := s.repo.UpsertReadingProgress(rp); err != nil {
			logger.Warn("同步 reading_progress 失败", zap.Int("comic_id", comicID), zap.Error(err))
		}
	}()

	return nil
}

// ToggleFavorite 切换收藏状态
func (s *ComicService) ToggleFavorite(comicID int, title, coverURL string) (bool, error) {
	isFavorited, err := s.repo.IsFavorited(comicID)
	if err != nil {
		return false, err
	}

	if isFavorited {
		if err := s.repo.RemoveFavorite(comicID); err != nil {
			return false, err
		}
		return false, nil
	} else {
		if err := s.repo.AddFavorite(comicID, title, coverURL); err != nil {
			return false, err
		}
		return true, nil
	}
}

// GetDashboardStats 获取仪表盘统计
func (s *ComicService) GetDashboardStats() (*model.DashboardStats, error) {
	return s.repo.GetDashboardStats()
}

// GetFavorites 获取收藏列表
func (s *ComicService) GetFavorites(limit, offset int) ([]model.ComicFavorite, int64, error) {
	return s.repo.GetFavorites(limit, offset)
}

// GetHistory 获取历史记录
func (s *ComicService) GetHistory(limit, offset int) ([]model.ReadingHistory, int64, error) {
	return s.repo.GetReadingHistory(limit, offset)
}

// UpdateSiteStats 从网站获取并更新总漫画数和总页数
func (s *ComicService) UpdateSiteStats() error {
	// 获取第1页，拿到网站总页数
	html, err := s.crawler.GetListPage(1, nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("获取第1页失败: %w", err)
	}

	items, siteTotalPages, err := parser.ParseListPage(html)
	if err != nil {
		return fmt.Errorf("解析第1页失败: %w", err)
	}

	// 缓存第1页数据
	go s.saveListCache(1, items)

	if siteTotalPages <= 0 {
		return fmt.Errorf("网站总页数异常: %d", siteTotalPages)
	}

	// 检查本地存储的总页数是否一致
	oldSitePages := 0
	if val, err := s.repo.GetSystemMetadata("site_total_pages"); err == nil {
		oldSitePages, _ = strconv.Atoi(val)
	}

	var totalComics int64
	if siteTotalPages == oldSitePages && oldSitePages > 0 {
		// 总页数没变，不需要重新计算
		if val, err := s.repo.GetSystemMetadata("total_comics_count"); err == nil {
			tc, _ := strconv.ParseInt(val, 10, 64)
			if tc > 0 {
				logger.Info("网站总数未变化", zap.Int("site_total_pages", siteTotalPages), zap.Int64("total_comics", tc))
				return nil
			}
		}
	}

	// 总页数变了或首次获取，需要访问最后一页算精确总数
	logger.Info("网站总页数变化，更新总数", zap.Int("old", oldSitePages), zap.Int("new", siteTotalPages))

	if siteTotalPages == 1 {
		totalComics = int64(len(items))
	} else {
		lastHTML, err := s.crawler.GetListPage(siteTotalPages, nil, nil, nil, "")
		if err != nil {
			return fmt.Errorf("获取最后一页失败: %w", err)
		}
		lastItems, _, err := parser.ParseListPage(lastHTML)
		if err != nil {
			return fmt.Errorf("解析最后一页失败: %w", err)
		}
		totalComics = int64(siteTotalPages-1)*sitePageSize + int64(len(lastItems))

		// 缓存最后一页数据
		go s.saveListCache(siteTotalPages, lastItems)
	}

	// 保存到SystemMetadata
	if err := s.repo.SetSystemMetadata("site_total_pages", strconv.Itoa(siteTotalPages)); err != nil {
		return fmt.Errorf("保存site_total_pages失败: %w", err)
	}
	if err := s.repo.SetSystemMetadata("total_comics_count", strconv.FormatInt(totalComics, 10)); err != nil {
		return fmt.Errorf("保存total_comics_count失败: %w", err)
	}

	logger.Info("更新网站总数完成", zap.Int("site_total_pages", siteTotalPages), zap.Int64("total_comics", totalComics))
	return nil
}

// StartSiteStatsUpdater 启动定时更新网站总数的任务
func (s *ComicService) StartSiteStatsUpdater(stopCh <-chan struct{}) {
	// 启动时立即执行一次
	if err := s.UpdateSiteStats(); err != nil {
		logger.Error("首次更新网站总数失败", zap.Error(err))
	}

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.UpdateSiteStats(); err != nil {
				logger.Error("定时更新网站总数失败", zap.Error(err))
			}
		case <-stopCh:
			logger.Info("网站总数更新任务已停止")
			return
		}
	}
}
