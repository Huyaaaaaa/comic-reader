package repository

import (
	"comic-viewer-claude/internal/model"
	"time"
)

// CreateImportExportJob 创建导入导出任务
func (r *Repository) CreateImportExportJob(job *model.ImportExportJob) error {
	return r.db.Create(job).Error
}

// GetImportExportJob 获取任务
func (r *Repository) GetImportExportJob(id int) (*model.ImportExportJob, error) {
	var job model.ImportExportJob
	err := r.db.First(&job, id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateImportExportJob 更新任务
func (r *Repository) UpdateImportExportJob(job *model.ImportExportJob) error {
	job.UpdatedAt = time.Now()
	return r.db.Save(job).Error
}

// UpdateImportExportJobStatus 更新任务状态
func (r *Repository) UpdateImportExportJobStatus(id int, status string, updates map[string]interface{}) error {
	if updates == nil {
		updates = map[string]interface{}{}
	}
	updates["status"] = status
	updates["updated_at"] = time.Now()
	return r.db.Model(&model.ImportExportJob{}).Where("id = ?", id).Updates(updates).Error
}

// GetRunningExportCount 获取正在运行的导出任务数
func (r *Repository) GetRunningExportCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.ImportExportJob{}).
		Where("direction = ? AND status IN ?", "export", []string{"queued", "running"}).
		Count(&count).Error
	return count, err
}

// GetAllCachedComicIDs 获取所有已缓存漫画ID（有detail的）
func (r *Repository) GetAllCachedComicIDs() ([]int, error) {
	var ids []int
	err := r.db.Model(&model.Comic{}).Where("title != ''").Pluck("id", &ids).Error
	return ids, err
}

// GetOfflineReadyComicIDs 获取所有离线就绪漫画ID（L3Cached=true）
func (r *Repository) GetOfflineReadyComicIDs() ([]int, error) {
	var ids []int
	err := r.db.Model(&model.CacheState{}).Where("l3_cached = ?", true).Pluck("comic_id", &ids).Error
	return ids, err
}
