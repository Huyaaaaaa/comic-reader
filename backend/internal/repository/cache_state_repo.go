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

// UpsertCacheState 更新或插入缓存状态
func (r *Repository) UpsertCacheState(state *model.CacheState) error {
	state.UpdatedAt = time.Now()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "comic_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"l1_cached", "l2_cached", "l3_cached", "l3_progress", "l3_total", "l3_current", "updated_at"}),
	}).Create(state).Error
}

// GetOfflineComics 获取离线可用漫画（L3Cached=true）
func (r *Repository) GetOfflineComics(limit, offset int) ([]model.CacheState, int64, error) {
	var states []model.CacheState
	var total int64

	query := r.db.Model(&model.CacheState{}).Where("l3_cached = ?", true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&states).Error
	return states, total, err
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
