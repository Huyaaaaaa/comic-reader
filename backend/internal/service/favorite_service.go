package service

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// FavoriteService 收藏服务（包装现有功能，同步写入 reading_progress 和 cache_state）
type FavoriteService struct {
	repo         *repository.Repository
	comicService *ComicService
}

// NewFavoriteService 创建收藏服务
func NewFavoriteService(repo *repository.Repository, comicService *ComicService) *FavoriteService {
	return &FavoriteService{
		repo:         repo,
		comicService: comicService,
	}
}

// ToggleFavorite 切换收藏，同步更新 reading_progress 和 cache_state
func (s *FavoriteService) ToggleFavorite(comicID int, title, coverURL string) (bool, error) {
	isFavorited, err := s.comicService.ToggleFavorite(comicID, title, coverURL)
	if err != nil {
		return false, err
	}

	// 收藏时同步写入 reading_progress
	if isFavorited {
		go func() {
			rp := &model.ReadingProgress{
				ComicID:   comicID,
				UserID:    1,
				UpdatedAt: time.Now(),
			}
			if err := s.repo.UpsertReadingProgress(rp); err != nil {
				logger.Warn("同步 reading_progress 失败", zap.Error(err))
			}

			// 确保 cache_state 记录存在（不覆盖已有层级状态）
			if err := s.repo.UpsertCacheStateL1(comicID); err != nil {
				logger.Warn("同步 cache_state 失败", zap.Error(err))
			}
		}()
	}

	return isFavorited, nil
}
