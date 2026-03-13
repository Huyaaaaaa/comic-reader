package service

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
)

// SettingsService 设置服务
type SettingsService struct {
	repo *repository.Repository
}

// NewSettingsService 创建设置服务
func NewSettingsService(repo *repository.Repository) *SettingsService {
	return &SettingsService{repo: repo}
}

// GetSetting 获取单个设置
func (s *SettingsService) GetSetting(userID int, key string) (string, error) {
	return s.repo.GetUserSetting(userID, key)
}

// SetSetting 设置单个设置
func (s *SettingsService) SetSetting(userID int, key, value string) error {
	return s.repo.SetUserSetting(userID, key, value)
}

// GetAllSettings 获取所有设置
func (s *SettingsService) GetAllSettings(userID int) ([]model.UserSetting, error) {
	return s.repo.GetAllUserSettings(userID)
}

// SetBatchSettings 批量更新设置
func (s *SettingsService) SetBatchSettings(userID int, settings map[string]string) error {
	for key, value := range settings {
		if err := s.repo.SetUserSetting(userID, key, value); err != nil {
			return err
		}
	}
	return nil
}
