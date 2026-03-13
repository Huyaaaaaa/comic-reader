package repository

import (
	"comic-viewer-claude/internal/model"
	"time"

	"gorm.io/gorm/clause"
)

// GetReadingProgress 获取阅读进度
func (r *Repository) GetReadingProgress(comicID, userID int) (*model.ReadingProgress, error) {
	var progress model.ReadingProgress
	err := r.db.Where("comic_id = ? AND user_id = ?", comicID, userID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// UpsertReadingProgress 更新或插入阅读进度
func (r *Repository) UpsertReadingProgress(progress *model.ReadingProgress) error {
	progress.UpdatedAt = time.Now()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "comic_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_page", "total_pages", "progress", "chapter_id", "updated_at"}),
	}).Create(progress).Error
}

// GetAllReadingProgress 获取所有阅读进度
func (r *Repository) GetAllReadingProgress(userID, limit, offset int) ([]model.ReadingProgress, int64, error) {
	var progresses []model.ReadingProgress
	var total int64

	query := r.db.Model(&model.ReadingProgress{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&progresses).Error
	return progresses, total, err
}
