package service

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

const appVersion = "1.0.0"

// UpdateService 更新服务
type UpdateService struct {
	repo         *repository.Repository
	comicService *ComicService
}

// NewUpdateService 创建更新服务
func NewUpdateService(repo *repository.Repository, comicService *ComicService) *UpdateService {
	return &UpdateService{
		repo:         repo,
		comicService: comicService,
	}
}

// CheckContentUpdate 检查内容更新（头部扫描前 N 页，发现新漫画）
func (s *UpdateService) CheckContentUpdate(pages int, mode string) (*model.ContentUpdateResponse, error) {
	if pages < 1 {
		pages = 2
	}
	if pages > 10 {
		pages = 10
	}
	if mode == "" {
		mode = "manual"
	}

	// 获取已知漫画ID
	knownIDs, err := s.repo.GetKnownListCacheIDs()
	if err != nil {
		knownIDs = make(map[int]bool)
	}

	// 扫描前 N 页
	var newComics []model.NewComicBrief
	for page := 1; page <= pages; page++ {
		resp, err := s.comicService.GetList(page, 100, false)
		if err != nil {
			logger.Warn("扫描列表页失败", zap.Int("page", page), zap.Error(err))
			continue
		}

		for _, item := range resp.Items {
			if !knownIDs[item.ID] {
				newComics = append(newComics, model.NewComicBrief{
					ID:       item.ID,
					Title:    item.Title,
					CoverURL: item.CoverURL,
				})
			}
		}
	}

	hasUpdate := len(newComics) > 0

	// 保存更新记录
	resultBytes, _ := json.Marshal(map[string]interface{}{
		"new_comics":    len(newComics),
		"scanned_pages": pages,
	})

	record := &model.UpdateRecord{
		UpdateType:     "content",
		HasUpdate:      hasUpdate,
		CheckMode:      mode,
		CurrentVersion: appVersion,
		Result:         string(resultBytes),
		CheckedAt:      time.Now(),
	}
	s.repo.CreateUpdateRecord(record)

	return &model.ContentUpdateResponse{
		HasUpdate:    hasUpdate,
		NewComics:    len(newComics),
		ScannedPages: pages,
		Details:      newComics,
		RecordID:     record.ID,
	}, nil
}

// CheckAppUpdate 检查应用更新
func (s *UpdateService) CheckAppUpdate() (*model.AppUpdateResponse, error) {
	// 当前版本
	currentVersion := appVersion

	// 目前无远端更新源，直接返回无更新
	// 未来可对接 GitHub Releases API 或自建更新服务器
	resp := &model.AppUpdateResponse{
		HasUpdate:      false,
		CurrentVersion: currentVersion,
	}

	// 保存更新记录
	resultBytes, _ := json.Marshal(map[string]interface{}{
		"current_version": currentVersion,
		"has_update":      false,
	})

	record := &model.UpdateRecord{
		UpdateType:     "app",
		HasUpdate:      false,
		CheckMode:      "manual",
		CurrentVersion: currentVersion,
		Result:         string(resultBytes),
		CheckedAt:      time.Now(),
	}
	s.repo.CreateUpdateRecord(record)

	resp.RecordID = record.ID
	return resp, nil
}
