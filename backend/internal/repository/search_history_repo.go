package repository

import (
	"comic-viewer-claude/internal/model"
	"time"
)

// AddSearchHistory 添加搜索历史
func (r *Repository) AddSearchHistory(userID int, keyword string, resultCount int) error {
	history := model.SearchHistory{
		UserID:      userID,
		Keyword:     keyword,
		ResultCount: resultCount,
		SearchedAt:  time.Now(),
	}
	return r.db.Create(&history).Error
}

// GetSearchHistory 获取搜索历史
func (r *Repository) GetSearchHistory(userID, limit, offset int) ([]model.SearchHistory, int64, error) {
	var histories []model.SearchHistory
	var total int64

	query := r.db.Model(&model.SearchHistory{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("searched_at DESC").Limit(limit).Offset(offset).Find(&histories).Error
	return histories, total, err
}

// ClearSearchHistory 清空搜索历史
func (r *Repository) ClearSearchHistory(userID int) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error
}
