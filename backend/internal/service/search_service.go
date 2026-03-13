package service

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"

	"go.uber.org/zap"
)

// SearchService 搜索服务
type SearchService struct {
	repo         *repository.Repository
	comicService *ComicService
}

// NewSearchService 创建搜索服务
func NewSearchService(repo *repository.Repository, comicService *ComicService) *SearchService {
	return &SearchService{
		repo:         repo,
		comicService: comicService,
	}
}

// SearchWithHistory 搜索并记录历史
func (s *SearchService) SearchWithHistory(userID int, keyword string, page, pageSize int) (*model.ListResponse, error) {
	resp, err := s.comicService.SearchLocal(keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 异步保存搜索历史
	go func() {
		if err := s.repo.AddSearchHistory(userID, keyword, int(resp.Total)); err != nil {
			logger.Warn("保存搜索历史失败", zap.Error(err))
		}
	}()

	return resp, nil
}

// GetSearchHistory 获取搜索历史
func (s *SearchService) GetSearchHistory(userID, limit, offset int) ([]model.SearchHistory, int64, error) {
	return s.repo.GetSearchHistory(userID, limit, offset)
}

// ClearSearchHistory 清空搜索历史
func (s *SearchService) ClearSearchHistory(userID int) error {
	return s.repo.ClearSearchHistory(userID)
}
