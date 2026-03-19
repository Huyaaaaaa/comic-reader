package service

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"fmt"

	"go.uber.org/zap"
)

// TagInitService 标签初始化服务
type TagInitService struct {
	repo         *repository.Repository
	comicService *ComicService
}

// NewTagInitService 创建标签初始化服务
func NewTagInitService(repo *repository.Repository, comicService *ComicService) *TagInitService {
	return &TagInitService{
		repo:         repo,
		comicService: comicService,
	}
}

// InitializeTags 初始化标签（从源站抓取多页漫画列表来收集标签）
func (s *TagInitService) InitializeTags() error {
	// 检查标签表是否为空
	tags, err := s.repo.GetAllTags()
	if err != nil {
		return fmt.Errorf("检查标签表失败: %w", err)
	}

	if len(tags) > 0 {
		logger.Info("标签表已有数据，跳过初始化", zap.Int("count", len(tags)))
		return nil
	}

	logger.Info("开始初始化标签，从源站抓取漫画列表...")

	// 抓取前 10 页漫画列表来收集标签
	tagMap := make(map[int]string) // tag_id -> tag_name
	maxPages := 10
	pageSize := 40

	for page := 1; page <= maxPages; page++ {
		logger.Info("抓取漫画列表", zap.Int("page", page))

		// 使用 ComicService 的 GetList 方法，useCache=false 强制从源站获取
		listResp, err := s.comicService.GetList(page, pageSize, false)
		if err != nil {
			logger.Warn("抓取漫画列表失败", zap.Int("page", page), zap.Error(err))
			continue
		}

		// 遍历每个漫画，获取详情来收集标签
		for _, item := range listResp.Items {
			detail, err := s.comicService.GetDetail(item.ID)
			if err != nil {
				logger.Warn("获取漫画详情失败", zap.Int("comic_id", item.ID), zap.Error(err))
				continue
			}

			// 收集标签
			for _, tag := range detail.Tags {
				tagMap[tag.TagID] = tag.TagName
			}
		}

		logger.Info("已收集标签", zap.Int("page", page), zap.Int("total_tags", len(tagMap)))
	}

	// 保存标签到数据库
	logger.Info("保存标签到数据库", zap.Int("count", len(tagMap)))
	savedCount := 0
	for tagID, tagName := range tagMap {
		tag := &model.Tag{
			ID:   tagID,
			Name: tagName,
		}
		if err := s.repo.SaveTag(tag); err != nil {
			logger.Warn("保存标签失败", zap.Int("tag_id", tagID), zap.String("tag_name", tagName), zap.Error(err))
		} else {
			savedCount++
		}
	}

	logger.Info("标签初始化完成", zap.Int("saved", savedCount), zap.Int("total", len(tagMap)))
	return nil
}
