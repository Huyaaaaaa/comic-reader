package repository

import (
	"comic-viewer-claude/internal/model"
)

// CreateUpdateRecord 创建更新记录
func (r *Repository) CreateUpdateRecord(record *model.UpdateRecord) error {
	return r.db.Create(record).Error
}

// GetUpdateRecord 获取更新记录
func (r *Repository) GetUpdateRecord(id int) (*model.UpdateRecord, error) {
	var record model.UpdateRecord
	err := r.db.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetLatestUpdateRecord 获取最新更新记录
func (r *Repository) GetLatestUpdateRecord(updateType string) (*model.UpdateRecord, error) {
	var record model.UpdateRecord
	err := r.db.Where("update_type = ?", updateType).
		Order("checked_at DESC").First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetKnownComicIDs 获取所有已知漫画ID（用于检测新漫画）
func (r *Repository) GetKnownComicIDs() (map[int]bool, error) {
	var ids []int
	err := r.db.Model(&model.Comic{}).Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

// GetKnownListCacheIDs 获取列表缓存中的漫画ID
func (r *Repository) GetKnownListCacheIDs() (map[int]bool, error) {
	var ids []int
	err := r.db.Model(&model.ComicListCache{}).Distinct("comic_id").Pluck("comic_id", &ids).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}
