package service

import (
	"comic-viewer-claude/internal/download"
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/config"
	"fmt"
	"time"
)

// DownloadService 下载服务
type DownloadService struct {
	repo   *repository.Repository
	engine *download.Engine
	dlCfg  *config.DownloadConfig
}

// NewDownloadService 创建下载服务
func NewDownloadService(repo *repository.Repository, engine *download.Engine, dlCfg *config.DownloadConfig) *DownloadService {
	return &DownloadService{
		repo:   repo,
		engine: engine,
		dlCfg:  dlCfg,
	}
}

// 允许的状态转换
var allowedTransitions = map[string][]string{
	"queued":      {"downloading", "paused", "canceled"},
	"downloading": {"verifying", "paused", "failed", "canceled"},
	"verifying":   {"completed", "failed", "canceled"},
	"paused":      {"queued", "canceled"},
	"failed":      {"queued", "canceled"},
}

// CreateTask 创建下载任务
func (s *DownloadService) CreateTask(comicID int, taskType string) (*model.DownloadTask, error) {
	if taskType == "" {
		taskType = "images"
	}

	// 去重检查
	existing, err := s.repo.GetDownloadTaskByComic(comicID, 1)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("该漫画已有活跃下载任务 (id=%d, status=%s)", existing.ID, existing.Status)
	}

	// 获取图片列表确定 total
	images, err := s.repo.GetComicImages(comicID)
	total := len(images)

	// 获取漫画标题
	title := ""
	if comic, err := s.repo.GetComic(comicID); err == nil {
		title = comic.Title
	}

	now := time.Now()
	task := &model.DownloadTask{
		ComicID:       comicID,
		UserID:        1,
		Title:         title,
		TaskType:      taskType,
		TriggerSource: "manual",
		Status:        "queued",
		Priority:      0,
		Total:         total,
		Current:       0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateDownloadTask(task); err != nil {
		return nil, fmt.Errorf("创建下载任务失败: %w", err)
	}

	return task, nil
}

// Precheck 下载预检查
func (s *DownloadService) Precheck(comicIDs []int) (*model.DownloadPrecheckResponse, error) {
	dir := s.dlCfg.Dir
	if dir == "" {
		dir = "./downloads"
	}

	// 估算大小：每张图片约 500KB
	estimatedImages := 0
	for _, comicID := range comicIDs {
		images, _ := s.repo.GetComicImages(comicID)
		if len(images) > 0 {
			estimatedImages += len(images)
		} else {
			estimatedImages += 30 // 默认估算
		}
	}
	estimatedSize := int64(estimatedImages) * 500 * 1024 // 500KB per image

	availableSpace, err := download.GetAvailableSpace(dir)
	if err != nil {
		// 目录不存在，尝试获取父目录的空间
		availableSpace, err = download.GetAvailableSpace(".")
		if err != nil {
			return nil, fmt.Errorf("无法获取磁盘空间: %w", err)
		}
	}

	canDownload := availableSpace > estimatedSize+download.ReservedSpace
	var warnings []string
	if !canDownload {
		warnings = append(warnings, fmt.Sprintf(
			"磁盘空间不足: 需要 %dMB, 可用 %dMB",
			estimatedSize/1024/1024,
			availableSpace/1024/1024,
		))
	} else if availableSpace < estimatedSize*2+download.ReservedSpace {
		warnings = append(warnings, "磁盘空间较紧张，建议清理后再下载")
	}

	return &model.DownloadPrecheckResponse{
		TargetPath:     dir,
		AvailableSpace: availableSpace,
		ReservedSpace:  download.ReservedSpace,
		EstimatedSize:  estimatedSize,
		CanDownload:    canDownload,
		Warnings:       warnings,
	}, nil
}

// GetTasks 获取任务列表
func (s *DownloadService) GetTasks(group string) ([]model.DownloadTaskResponse, error) {
	tasks, err := s.repo.GetDownloadTasks(group)
	if err != nil {
		return nil, err
	}

	result := make([]model.DownloadTaskResponse, len(tasks))
	for i, task := range tasks {
		resp := model.DownloadTaskResponse{
			DownloadTask: task,
		}
		// 附带漫画信息
		if comic, err := s.repo.GetComic(task.ComicID); err == nil {
			resp.ComicTitle = comic.Title
			resp.ComicCoverURL = comic.CoverURL
		}
		result[i] = resp
	}

	return result, nil
}

// PauseTask 暂停任务
func (s *DownloadService) PauseTask(taskID int) error {
	task, err := s.repo.GetDownloadTask(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	if !isTransitionAllowed(task.Status, "paused") {
		return fmt.Errorf("当前状态 %s 不允许暂停", task.Status)
	}

	return s.repo.UpdateDownloadTaskStatus(taskID, "paused", map[string]interface{}{
		"lock_token": "",
	})
}

// ResumeTask 恢复任务
func (s *DownloadService) ResumeTask(taskID int) error {
	task, err := s.repo.GetDownloadTask(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	if !isTransitionAllowed(task.Status, "queued") {
		return fmt.Errorf("当前状态 %s 不允许恢复", task.Status)
	}

	return s.repo.UpdateDownloadTaskStatus(taskID, "queued", nil)
}

// CancelTask 取消任务
func (s *DownloadService) CancelTask(taskID int) error {
	task, err := s.repo.GetDownloadTask(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	if task.Status == "completed" {
		return fmt.Errorf("已完成的任务不能取消")
	}

	if !isTransitionAllowed(task.Status, "canceled") {
		return fmt.Errorf("当前状态 %s 不允许取消", task.Status)
	}

	return s.repo.UpdateDownloadTaskStatus(taskID, "canceled", map[string]interface{}{
		"lock_token": "",
	})
}

// RetryTask 重试任务（仅限 failed 状态）
func (s *DownloadService) RetryTask(taskID int) error {
	task, err := s.repo.GetDownloadTask(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	if task.Status != "failed" {
		return fmt.Errorf("只有失败的任务才能重试 (当前: %s)", task.Status)
	}

	return s.repo.UpdateDownloadTaskStatus(taskID, "queued", map[string]interface{}{
		"retry_count": task.RetryCount + 1,
		"last_error":  "",
		"lock_token":  "",
	})
}

// isTransitionAllowed 检查状态转换是否合法
func isTransitionAllowed(from, to string) bool {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
