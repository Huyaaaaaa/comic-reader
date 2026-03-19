package repository

import (
	"comic-viewer-claude/internal/model"
	"time"

	"gorm.io/gorm/clause"
)

// GetCacheState 获取缓存状态
func (r *Repository) GetCacheState(comicID int) (*model.CacheState, error) {
	var state model.CacheState
	err := r.db.First(&state, comicID).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// UpsertCacheState 更新或插入缓存状态（仅用于完整覆盖场景，一般请用分层方法）
func (r *Repository) UpsertCacheState(state *model.CacheState) error {
	state.UpdatedAt = time.Now()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "comic_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"l1_cached", "l2_cached", "l3_cached", "l3_progress", "l3_total", "l3_current", "updated_at"}),
	}).Create(state).Error
}

// UpsertCacheStateL1 仅更新 L1 缓存状态，不影响其他层级
func (r *Repository) UpsertCacheStateL1(comicID int) error {
	now := time.Now()
	result := r.db.Model(&model.CacheState{}).Where("comic_id = ?", comicID).Update("l1_cached", true).Update("updated_at", now)
	if result.RowsAffected == 0 {
		return r.db.Create(&model.CacheState{ComicID: comicID, L1Cached: true, UpdatedAt: now}).Error
	}
	return result.Error
}

// UpsertCacheStateL2 仅更新 L2 缓存状态，不影响其他层级
func (r *Repository) UpsertCacheStateL2(comicID int) error {
	now := time.Now()
	result := r.db.Model(&model.CacheState{}).Where("comic_id = ?", comicID).Update("l2_cached", true).Update("updated_at", now)
	if result.RowsAffected == 0 {
		return r.db.Create(&model.CacheState{ComicID: comicID, L2Cached: true, UpdatedAt: now}).Error
	}
	return result.Error
}

// UpsertCacheStateL3 仅更新 L3 缓存状态，不影响其他层级
func (r *Repository) UpsertCacheStateL3(comicID int, total, current int, progress float64) error {
	now := time.Now()
	result := r.db.Model(&model.CacheState{}).Where("comic_id = ?", comicID).Updates(map[string]interface{}{
		"l3_cached":   true,
		"l3_total":    total,
		"l3_current":  current,
		"l3_progress": progress,
		"updated_at":  now,
	})
	if result.RowsAffected == 0 {
		return r.db.Create(&model.CacheState{ComicID: comicID, L3Cached: true, L3Total: total, L3Current: current, L3Progress: progress, UpdatedAt: now}).Error
	}
	return result.Error
}

// OfflineComicItem 离线漫画（缓存状态+漫画基本信息）
type OfflineComicItem struct {
	ComicID    int       `json:"comic_id"`
	Title      string    `json:"title"`
	CoverURL   string    `json:"cover_url"`
	Author     string    `json:"author"`
	Rating     float64   `json:"rating"`
	RatingCount int      `json:"rating_count"`
	Favorites  int       `json:"favorites"`
	L3Cached   bool      `json:"l3_cached"`
	L3Progress float64   `json:"l3_progress"`
	L3Total    int       `json:"l3_total"`
	L3Current  int       `json:"l3_current"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetOfflineComics 获取离线可用漫画（L3Cached=true），JOIN comics 表获取基本信息
func (r *Repository) GetOfflineComics(limit, offset int) ([]OfflineComicItem, int64, error) {
	var total int64
	r.db.Model(&model.CacheState{}).Where("l3_cached = ?", true).Count(&total)

	var items []OfflineComicItem
	err := r.db.Table("cache_states").
		Select("cache_states.comic_id, comics.title, comics.cover_url, comics.author, comics.rating, comics.rating_count, comics.favorites, cache_states.l3_cached, cache_states.l3_progress, cache_states.l3_total, cache_states.l3_current, cache_states.updated_at").
		Joins("LEFT JOIN comics ON cache_states.comic_id = comics.id").
		Where("cache_states.l3_cached = ?", true).
		Order("cache_states.updated_at DESC").
		Limit(limit).Offset(offset).
		Scan(&items).Error

	return items, total, err
}

// GetCacheOverview 获取缓存概览统计
func (r *Repository) GetCacheOverview() (total, l1, l2, l3 int, err error) {
	var totalCount, l1Count, l2Count, l3Count int64

	r.db.Model(&model.CacheState{}).Count(&totalCount)
	r.db.Model(&model.CacheState{}).Where("l1_cached = ?", true).Count(&l1Count)
	r.db.Model(&model.CacheState{}).Where("l2_cached = ?", true).Count(&l2Count)
	r.db.Model(&model.CacheState{}).Where("l3_cached = ?", true).Count(&l3Count)

	return int(totalCount), int(l1Count), int(l2Count), int(l3Count), nil
}

// GetL1OnlyComicIDs 获取已有 L1 缓存但未有 L2 缓存的漫画ID列表
func (r *Repository) GetL1OnlyComicIDs(limit int) ([]int, error) {
	var ids []int
	err := r.db.Model(&model.CacheState{}).
		Where("l1_cached = ? AND l2_cached = ?", true, false).
		Order("updated_at DESC").
		Limit(limit).
		Pluck("comic_id", &ids).Error
	return ids, err
}

// GetL2OnlyComicIDs 获取已有 L2 缓存但未有 L3 缓存的漫画ID列表
func (r *Repository) GetL2OnlyComicIDs(limit int) ([]int, error) {
	var ids []int
	err := r.db.Model(&model.CacheState{}).
		Where("l2_cached = ? AND l3_cached = ?", true, false).
		Order("updated_at DESC").
		Limit(limit).
		Pluck("comic_id", &ids).Error
	return ids, err
}
