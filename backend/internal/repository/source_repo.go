package repository

import (
	"comic-viewer-claude/internal/model"
	"time"
)

// GetAllSources 获取所有源站
func (r *Repository) GetAllSources() ([]model.SourceSite, error) {
	var sources []model.SourceSite
	err := r.db.Order("priority ASC, id ASC").Find(&sources).Error
	return sources, err
}

// GetActiveSource 获取当前活跃源站（优先级最高的 active 源站）
func (r *Repository) GetActiveSource() (*model.SourceSite, error) {
	var source model.SourceSite
	err := r.db.Where("status = ?", "active").Order("priority ASC, id ASC").First(&source).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// GetSourceByID 根据 ID 获取源站
func (r *Repository) GetSourceByID(id int) (*model.SourceSite, error) {
	var source model.SourceSite
	err := r.db.First(&source, id).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// GetSourceByURL 根据 URL 获取源站
func (r *Repository) GetSourceByURL(url string) (*model.SourceSite, error) {
	var source model.SourceSite
	err := r.db.Where("url = ?", url).First(&source).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// AddSource 添加源站
func (r *Repository) AddSource(source *model.SourceSite) error {
	return r.db.Create(source).Error
}

// DeleteSource 删除源站
func (r *Repository) DeleteSource(id int) error {
	return r.db.Delete(&model.SourceSite{}, id).Error
}

// UpdateSourceStatus 更新源站总体状态
func (r *Repository) UpdateSourceStatus(id int, status string, latency int) error {
	now := time.Now()
	return r.db.Model(&model.SourceSite{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"latency":    latency,
		"last_check": &now,
	}).Error
}

// UpdateSourceFields 批量更新源站字段
func (r *Repository) UpdateSourceFields(id int, fields map[string]interface{}) error {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if _, ok := fields["last_check"]; !ok {
		now := time.Now()
		fields["last_check"] = &now
	}
	return r.db.Model(&model.SourceSite{}).Where("id = ?", id).Updates(fields).Error
}

// IncrementSourceFailCount 增加源站失败计数
func (r *Repository) IncrementSourceFailCount(id int) error {
	return r.db.Model(&model.SourceSite{}).Where("id = ?", id).
		UpdateColumn("fail_count", r.db.Raw("fail_count + 1")).Error
}

// ResetSourceFailCount 重置源站失败计数
func (r *Repository) ResetSourceFailCount(id int) error {
	return r.db.Model(&model.SourceSite{}).Where("id = ?", id).
		Update("fail_count", 0).Error
}
