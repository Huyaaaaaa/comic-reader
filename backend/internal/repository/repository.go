package repository

import (
	"comic-viewer-claude/internal/model"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Repository 数据仓库
type Repository struct {
	db *gorm.DB
}

// New 创建新的数据仓库
func New(dbPath string) (*Repository, error) {
	// 创建数据库目录
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 打开数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Repository{db: db}, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Comic{},
		&model.ComicImage{},
		&model.Tag{},
		&model.ComicTag{},
		&model.ComicAuthor{},
		&model.ComicMetadataState{},
		&model.ReadingHistory{},
		&model.ComicFavorite{},
		&model.ComicListCache{},
		&model.DownloadTask{},
		&model.CoverCacheQueue{},
		&model.SystemMetadata{},
		&model.SearchHistory{},
		&model.UserSetting{},
		&model.CacheState{},
		&model.ReadingProgress{},
		&model.SourceSite{},
	)
}

// Close 关闭数据库连接
func (r *Repository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取数据库实例
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// === Comic 相关 ===

// GetComic 获取漫画详情
func (r *Repository) GetComic(id int) (*model.Comic, error) {
	var comic model.Comic
	err := r.db.First(&comic, id).Error
	if err != nil {
		return nil, err
	}
	return &comic, nil
}

// SaveComic 保存漫画
func (r *Repository) SaveComic(comic *model.Comic) error {
	return r.db.Save(comic).Error
}

// GetComics 批量获取漫画
func (r *Repository) GetComics(ids []int) ([]model.Comic, error) {
	var comics []model.Comic
	err := r.db.Where("id IN ?", ids).Find(&comics).Error
	return comics, err
}

// SearchComicsLocal 本地搜索漫画
func (r *Repository) SearchComicsLocal(keyword string, limit, offset int) ([]model.Comic, int64, error) {
	var comics []model.Comic
	var total int64

	query := r.db.Model(&model.Comic{}).Where("title LIKE ? OR author LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Find(&comics).Error
	return comics, total, err
}

// === ComicImage 相关 ===

// GetComicImages 获取漫画图片列表
func (r *Repository) GetComicImages(comicID int) ([]model.ComicImage, error) {
	var images []model.ComicImage
	err := r.db.Where("comic_id = ?", comicID).Order("sort ASC").Find(&images).Error
	return images, err
}

// SaveComicImages 保存漫画图片
func (r *Repository) SaveComicImages(images []model.ComicImage) error {
	return r.db.Save(&images).Error
}

// === Tag 相关 ===

// GetAllTags 获取所有标签
func (r *Repository) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Order("id ASC").Find(&tags).Error
	return tags, err
}

// SaveTag 保存标签
func (r *Repository) SaveTag(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

// GetComicTags 获取漫画的标签
func (r *Repository) GetComicTags(comicID int) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Table("tags").
		Joins("JOIN comic_tags ON tags.id = comic_tags.tag_id").
		Where("comic_tags.comic_id = ?", comicID).
		Find(&tags).Error
	return tags, err
}

// SaveComicTags 保存漫画标签关联
func (r *Repository) SaveComicTags(comicID int, tagIDs []int) error {
	// 先删除旧的关联
	if err := r.db.Where("comic_id = ?", comicID).Delete(&model.ComicTag{}).Error; err != nil {
		return err
	}

	// 插入新的关联
	for _, tagID := range tagIDs {
		if err := r.db.Create(&model.ComicTag{ComicID: comicID, TagID: tagID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// === ComicAuthor 相关 ===

// GetComicAuthors 获取漫画的作者
func (r *Repository) GetComicAuthors(comicID int) ([]model.ComicAuthor, error) {
	var authors []model.ComicAuthor
	err := r.db.Where("comic_id = ?", comicID).Order("position ASC").Find(&authors).Error
	return authors, err
}

// SaveComicAuthors 保存漫画作者关联
func (r *Repository) SaveComicAuthors(authors []model.ComicAuthor) error {
	if len(authors) == 0 {
		return nil
	}

	// 先删除旧的关联
	comicID := authors[0].ComicID
	if err := r.db.Where("comic_id = ?", comicID).Delete(&model.ComicAuthor{}).Error; err != nil {
		return err
	}

	// 插入新的关联
	return r.db.Create(&authors).Error
}

// === ComicListCache 相关 ===

// GetCachedList 获取缓存的列表
func (r *Repository) GetCachedList(page int) ([]model.ComicListCache, error) {
	var items []model.ComicListCache
	err := r.db.Where("page = ?", page).Order("position ASC").Find(&items).Error
	return items, err
}

// SaveListCache 保存列表缓存
func (r *Repository) SaveListCache(items []model.ComicListCache) error {
	if len(items) == 0 {
		return nil
	}

	// 先删除该页的旧缓存
	page := items[0].Page
	if err := r.db.Where("page = ?", page).Delete(&model.ComicListCache{}).Error; err != nil {
		return err
	}

	// 插入新缓存
	return r.db.Create(&items).Error
}

// GetMaxCachedPage 获取最大缓存页码
func (r *Repository) GetMaxCachedPage() (int, error) {
	var maxPage int
	err := r.db.Model(&model.ComicListCache{}).Select("COALESCE(MAX(page), 0)").Scan(&maxPage).Error
	return maxPage, err
}

// GetSystemMetadata 获取系统元数据
func (r *Repository) GetSystemMetadata(key string) (string, error) {
	var meta model.SystemMetadata
	err := r.db.Where("`key` = ?", key).First(&meta).Error
	if err != nil {
		return "", err
	}
	return meta.Value, nil
}

// SetSystemMetadata 设置系统元数据
func (r *Repository) SetSystemMetadata(key, value string) error {
	meta := model.SystemMetadata{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
	return r.db.Save(&meta).Error
}

// === ReadingHistory 相关 ===

// AddReadingHistory 添加阅读历史
func (r *Repository) AddReadingHistory(comicID int, title, coverURL string) error {
	history := model.ReadingHistory{
		ComicID:    comicID,
		Title:      title,
		CoverURL:   coverURL,
		LastReadAt: time.Now(),
	}
	return r.db.Save(&history).Error
}

// GetReadingHistory 获取阅读历史
func (r *Repository) GetReadingHistory(limit, offset int) ([]model.ReadingHistory, int64, error) {
	var histories []model.ReadingHistory
	var total int64

	if err := r.db.Model(&model.ReadingHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("last_read_at DESC").Limit(limit).Offset(offset).Find(&histories).Error
	return histories, total, err
}

// === ComicFavorite 相关 ===

// AddFavorite 添加收藏
func (r *Repository) AddFavorite(comicID int, title, coverURL string) error {
	favorite := model.ComicFavorite{
		ComicID:     comicID,
		Title:       title,
		CoverURL:    coverURL,
		FavoritedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}
	return r.db.Save(&favorite).Error
}

// RemoveFavorite 移除收藏
func (r *Repository) RemoveFavorite(comicID int) error {
	return r.db.Delete(&model.ComicFavorite{}, comicID).Error
}

// IsFavorited 检查是否已收藏
func (r *Repository) IsFavorited(comicID int) (bool, error) {
	var count int64
	err := r.db.Model(&model.ComicFavorite{}).Where("comic_id = ?", comicID).Count(&count).Error
	return count > 0, err
}

// GetFavorites 获取收藏列表
func (r *Repository) GetFavorites(limit, offset int) ([]model.ComicFavorite, int64, error) {
	var favorites []model.ComicFavorite
	var total int64

	if err := r.db.Model(&model.ComicFavorite{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("favorited_at DESC").Limit(limit).Offset(offset).Find(&favorites).Error
	return favorites, total, err
}

// === DownloadTask 相关 ===

// GetDownloadTask 获取下载任务
func (r *Repository) GetDownloadTask(comicID int) (*model.DownloadTask, error) {
	var task model.DownloadTask
	err := r.db.First(&task, comicID).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// SaveDownloadTask 保存下载任务
func (r *Repository) SaveDownloadTask(task *model.DownloadTask) error {
	return r.db.Save(task).Error
}

// GetPendingDownloadTasks 获取待处理的下载任务
func (r *Repository) GetPendingDownloadTasks() ([]model.DownloadTask, error) {
	var tasks []model.DownloadTask
	err := r.db.Where("status IN ?", []string{"pending", "downloading"}).Find(&tasks).Error
	return tasks, err
}

// === CoverCacheQueue 相关 ===

// AddCoverCacheTask 添加封面缓存任务
func (r *Repository) AddCoverCacheTask(comicID int, title, coverURL string, priority int) error {
	task := model.CoverCacheQueue{
		ComicID:    comicID,
		Title:      title,
		CoverURL:   coverURL,
		Priority:   priority,
		Status:     "pending",
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	return r.db.Save(&task).Error
}

// GetPendingCoverTasks 获取待处理的封面缓存任务
func (r *Repository) GetPendingCoverTasks(limit int) ([]model.CoverCacheQueue, error) {
	var tasks []model.CoverCacheQueue
	err := r.db.Where("status = ?", "pending").
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// UpdateCoverTaskStatus 更新封面缓存任务状态
func (r *Repository) UpdateCoverTaskStatus(comicID int, status string, retryCount int) error {
	return r.db.Model(&model.CoverCacheQueue{}).
	Where("comic_id = ?", comicID).
		Updates(map[string]interface{}{
			"status":      status,
			"retry_count": retryCount,
			"updated_at":  time.Now(),
		}).Error
}

// DeleteCoverTask 删除封面缓存任务
func (r *Repository) DeleteCoverTask(comicID int) error {
	return r.db.Delete(&model.CoverCacheQueue{}, comicID).Error
}

// ClearPendingCoverQueue 清空待处理的封面缓存队列
func (r *Repository) ClearPendingCoverQueue() error {
	return r.db.Where("status = ?", "pending").Delete(&model.CoverCacheQueue{}).Error
}

// === 统计相关 ===

// GetDashboardStats 获取仪表盘统计数据
func (r *Repository) GetDashboardStats() (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	var totalComics int64
	var cachedComics int64
	var totalTags int64
	var favoritesCount int64
	var historyCount int64
	var downloadingCount int64
	var pendingDownloads int64

	// 总漫画数
	r.db.Model(&model.Comic{}).Count(&totalComics)
	stats.TotalComics = int(totalComics)

	// 已缓存漫画数（有封面缓存的）
	r.db.Model(&model.Comic{}).Where("has_cover_cached = ?", true).Count(&cachedComics)
	stats.CachedComics = int(cachedComics)

	// 封面缓存数
	stats.CoverCached = stats.CachedComics

	// 总标签数
	r.db.Model(&model.Tag{}).Count(&totalTags)
	stats.TotalTags = int(totalTags)

	// 收藏数
	r.db.Model(&model.ComicFavorite{}).Count(&favoritesCount)
	stats.FavoritesCount = int(favoritesCount)

	// 历史记录数
	r.db.Model(&model.ReadingHistory{}).Count(&historyCount)
	stats.HistoryCount = int(historyCount)

	// 下载中的任务数
	r.db.Model(&model.DownloadTask{}).Where("status = ?", "downloading").Count(&downloadingCount)
	stats.DownloadingCount = int(downloadingCount)

	// 待下载任务数
	r.db.Model(&model.DownloadTask{}).Where("status = ?", "pending").Count(&pendingDownloads)
	stats.PendingDownloads = int(pendingDownloads)

	return stats, nil
}
