package repository

import (
	"comic-viewer-claude/internal/model"
	"time"

	"gorm.io/gorm/clause"
)

// GetUserSetting 获取用户设置
func (r *Repository) GetUserSetting(userID int, key string) (string, error) {
	var setting model.UserSetting
	err := r.db.Where("user_id = ? AND `key` = ?", userID, key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

// SetUserSetting 设置用户设置（upsert）
func (r *Repository) SetUserSetting(userID int, key, value string) error {
	setting := model.UserSetting{
		UserID:    userID,
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&setting).Error
}

// GetAllUserSettings 获取用户所有设置
func (r *Repository) GetAllUserSettings(userID int) ([]model.UserSetting, error) {
	var settings []model.UserSetting
	err := r.db.Where("user_id = ?", userID).Find(&settings).Error
	return settings, err
}
