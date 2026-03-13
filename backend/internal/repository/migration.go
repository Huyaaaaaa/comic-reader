package repository

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunMigration 执行幂等数据迁移
func RunMigration(db *gorm.DB) {
	logger.Info("开始执行数据迁移...")

	migrateCacheStates(db)
	migrateReadingProgress(db)

	logger.Info("数据迁移完成")
}

// migrateCacheStates 从 comic_metadata_state 生成 cache_states 初始状态
func migrateCacheStates(db *gorm.DB) {
	var count int64
	db.Model(&model.CacheState{}).Count(&count)
	if count > 0 {
		logger.Info("cache_states 已有数据，跳过迁移", zap.Int64("count", count))
		return
	}

	var states []model.ComicMetadataState
	if err := db.Find(&states).Error; err != nil {
		logger.Error("读取 comic_metadata_state 失败", zap.Error(err))
		return
	}

	if len(states) == 0 {
		return
	}

	now := time.Now()
	for _, s := range states {
		cs := model.CacheState{
			ComicID:   s.ComicID,
			L1Cached:  s.ListCached,
			L2Cached:  s.DetailCached,
			L3Cached:  false,
			UpdatedAt: now,
		}
		if err := db.Create(&cs).Error; err != nil {
			logger.Warn("迁移 cache_state 失败", zap.Int("comic_id", s.ComicID), zap.Error(err))
		}
	}

	logger.Info("cache_states 迁移完成", zap.Int("count", len(states)))
}

// migrateReadingProgress 从 comic_favorites 生成 reading_progress 初始记录
func migrateReadingProgress(db *gorm.DB) {
	var count int64
	db.Model(&model.ReadingProgress{}).Count(&count)
	if count > 0 {
		logger.Info("reading_progress 已有数据，跳过迁移", zap.Int64("count", count))
		return
	}

	var favorites []model.ComicFavorite
	if err := db.Find(&favorites).Error; err != nil {
		logger.Error("读取 comic_favorites 失败", zap.Error(err))
		return
	}

	if len(favorites) == 0 {
		return
	}

	now := time.Now()
	for _, f := range favorites {
		rp := model.ReadingProgress{
			ComicID:   f.ComicID,
			UserID:    1,
			UpdatedAt: now,
		}
		if err := db.Create(&rp).Error; err != nil {
			logger.Warn("迁移 reading_progress 失败", zap.Int("comic_id", f.ComicID), zap.Error(err))
		}
	}

	logger.Info("reading_progress 迁移完成", zap.Int("count", len(favorites)))
}
