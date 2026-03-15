package scheduler

import (
	"comic-viewer-claude/internal/event"
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"context"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ComicLister L1 缓存所需接口
type ComicLister interface {
	GetList(page, pageSize int, useCache bool) (*model.ListResponse, error)
}

// ComicDetailer L2 缓存所需接口
type ComicDetailer interface {
	GetDetail(comicID int) (*model.ComicDetail, error)
}

// DownloadCreator L3 缓存所需接口
type DownloadCreator interface {
	CreateTask(comicID int, taskType string) (*model.DownloadTask, error)
}

// StrategyConfig 缓存策略配置
type StrategyConfig struct {
	L1Strategy string
	L1Count    int
	L2Strategy string
	L2Count    int
	L3Strategy string
	L3Count    int
}

// StrategyScheduler 缓存策略调度器
type StrategyScheduler struct {
	repo            *repository.Repository
	comicLister     ComicLister
	comicDetailer   ComicDetailer
	downloadCreator DownloadCreator
	eventService    *event.Service

	config    StrategyConfig
	mu        sync.RWMutex
	runCancel context.CancelFunc
	stopCh    chan struct{}
}

// NewStrategyScheduler 创建缓存策略调度器
func NewStrategyScheduler(
	repo *repository.Repository,
	comicLister ComicLister,
	comicDetailer ComicDetailer,
	downloadCreator DownloadCreator,
	eventService *event.Service,
) *StrategyScheduler {
	return &StrategyScheduler{
		repo:            repo,
		comicLister:     comicLister,
		comicDetailer:   comicDetailer,
		downloadCreator: downloadCreator,
		eventService:    eventService,
		stopCh:          make(chan struct{}),
	}
}

// Start 启动调度器
func (s *StrategyScheduler) Start() {
	s.loadConfig()
	logger.Info("缓存策略调度器启动",
		zap.String("L1", s.config.L1Strategy),
		zap.String("L2", s.config.L2Strategy),
		zap.String("L3", s.config.L3Strategy))
	s.startRun()
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.loadConfig()
				s.startRun()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop 停止调度器
func (s *StrategyScheduler) Stop() {
	s.mu.Lock()
	if s.runCancel != nil {
		s.runCancel()
		s.runCancel = nil
	}
	s.mu.Unlock()
	close(s.stopCh)
}

// ReloadConfig 外部触发重新加载配置
func (s *StrategyScheduler) ReloadConfig() {
	s.loadConfig()
	s.startRun()
}

func (s *StrategyScheduler) loadConfig() {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.repo.GetAllUserSettings(1)
	if err != nil {
		logger.Warn("读取缓存策略配置失败", zap.Error(err))
		return
	}
	m := make(map[string]string)
	for _, st := range settings {
		m[st.Key] = st.Value
	}
	s.config = StrategyConfig{
		L1Strategy: getOrDefault(m, "cache_l1_strategy", "passive"),
		L1Count:    getIntOrDefault(m, "cache_l1_count", 500),
		L2Strategy: getOrDefault(m, "cache_l2_strategy", "passive"),
		L2Count:    getIntOrDefault(m, "cache_l2_count", 500),
		L3Strategy: getOrDefault(m, "cache_l3_strategy", "passive"),
		L3Count:    getIntOrDefault(m, "cache_l3_count", 100),
	}
}

func (s *StrategyScheduler) startRun() {
	s.mu.Lock()
	if s.runCancel != nil {
		s.runCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.runCancel = cancel
	cfg := s.config
	s.mu.Unlock()

	go s.runOnce(ctx, cfg)
}

func (s *StrategyScheduler) runOnce(ctx context.Context, cfg StrategyConfig) {
	if s.shouldStop(ctx) {
		return
	}
	if cfg.L1Strategy == "active" || cfg.L1Strategy == "all" {
		s.runL1Active(ctx, cfg)
	}
	if s.shouldStop(ctx) {
		return
	}
	if cfg.L2Strategy == "active" || cfg.L2Strategy == "all" {
		s.runL2Active(ctx, cfg)
	}
	if s.shouldStop(ctx) {
		return
	}
	if cfg.L3Strategy == "active" || cfg.L3Strategy == "all" {
		s.runL3Active(ctx, cfg)
	}
}

// PLACEHOLDER_REST

func (s *StrategyScheduler) runL1Active(ctx context.Context, cfg StrategyConfig) {
	count := cfg.L1Count
	if cfg.L1Strategy == "all" {
		count = 999999
	}
	s.broadcast("cache:l1_start", map[string]interface{}{"target_count": count})
	pages := (count + 99) / 100
	if pages > 50 {
		pages = 50
	}
	cached := 0
	for page := 1; page <= pages; page++ {
		if s.shouldStop(ctx) {
			logger.Info("L1 主动缓存已取消", zap.Int("cached", cached))
			return
		}
		_, err := s.comicLister.GetList(page, 100, false)
		if err != nil {
			logger.Warn("L1 主动缓存拉取列表失败", zap.Int("page", page), zap.Error(err))
			break
		}
		cached += 100
		if cached >= count {
			break
		}
	}
	if s.shouldStop(ctx) {
		logger.Info("L1 主动缓存已取消", zap.Int("cached", cached))
		return
	}
	s.broadcast("cache:l1_done", map[string]interface{}{"cached": cached})
	logger.Info("L1 主动缓存完成", zap.Int("cached", cached))
}

func (s *StrategyScheduler) runL2Active(ctx context.Context, cfg StrategyConfig) {
	count := cfg.L2Count
	if cfg.L2Strategy == "all" {
		count = 999999
	}
	comicIDs, err := s.repo.GetL1OnlyComicIDs(count)
	if err != nil || len(comicIDs) == 0 {
		return
	}
	s.broadcast("cache:l2_start", map[string]interface{}{"target_count": len(comicIDs)})
	cached := 0
	for _, id := range comicIDs {
		if s.shouldStop(ctx) {
			logger.Info("L2 主动缓存已取消", zap.Int("cached", cached))
			return
		}
		if _, err := s.comicDetailer.GetDetail(id); err == nil {
			cached++
		}
		select {
		case <-ctx.Done():
			logger.Info("L2 主动缓存已取消", zap.Int("cached", cached))
			return
		case <-s.stopCh:
			logger.Info("L2 主动缓存已取消", zap.Int("cached", cached))
			return
		case <-time.After(2 * time.Second):
		}
	}
	if s.shouldStop(ctx) {
		logger.Info("L2 主动缓存已取消", zap.Int("cached", cached))
		return
	}
	s.broadcast("cache:l2_done", map[string]interface{}{"cached": cached})
	logger.Info("L2 主动缓存完成", zap.Int("cached", cached))
}

// PLACEHOLDER_L3

func (s *StrategyScheduler) runL3Active(ctx context.Context, cfg StrategyConfig) {
	count := cfg.L3Count
	if cfg.L3Strategy == "all" {
		count = 999999
	}
	comicIDs, err := s.repo.GetL2OnlyComicIDs(count)
	if err != nil || len(comicIDs) == 0 {
		return
	}
	s.broadcast("cache:l3_start", map[string]interface{}{"target_count": len(comicIDs)})
	created := 0
	for _, id := range comicIDs {
		if s.shouldStop(ctx) {
			logger.Info("L3 主动缓存已取消", zap.Int("created", created))
			return
		}
		if _, err := s.downloadCreator.CreateTask(id, "images"); err == nil {
			created++
		}
	}
	if s.shouldStop(ctx) {
		logger.Info("L3 主动缓存已取消", zap.Int("created", created))
		return
	}
	s.broadcast("cache:l3_done", map[string]interface{}{"created": created})
	logger.Info("L3 主动缓存任务创建完成", zap.Int("created", created))
}

func (s *StrategyScheduler) broadcast(eventType string, data map[string]interface{}) {
	if s.eventService != nil {
		s.eventService.Broadcast(event.Event{Type: eventType, Data: data})
	}
}

func (s *StrategyScheduler) shouldStop(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-s.stopCh:
		return true
	default:
		return false
	}
}

func getOrDefault(m map[string]string, key, def string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return def
}

func getIntOrDefault(m map[string]string, key string, def int) int {
	if v, ok := m[key]; ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
