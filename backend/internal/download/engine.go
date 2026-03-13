package download

import (
	"comic-viewer-claude/internal/crawler"
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/config"
	"comic-viewer-claude/pkg/logger"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// Engine 下载引擎
type Engine struct {
	repo       *repository.Repository
	crawler    *crawler.Client
	cfg        *config.DownloadConfig
	antiBanCfg *config.CrawlerConfig

	downloadDir string
	sfGroup     singleflight.Group
	workerWg    sync.WaitGroup
	stopCh      chan struct{}
	taskCh      chan *model.DownloadTask
	mu          sync.Mutex
	running     bool
}

// NewEngine 创建下载引擎
func NewEngine(repo *repository.Repository, crawler *crawler.Client, dlCfg *config.DownloadConfig, crawlerCfg *config.CrawlerConfig) *Engine {
	dir := dlCfg.Dir
	if dir == "" {
		dir = "./downloads"
	}

	return &Engine{
		repo:        repo,
		crawler:     crawler,
		cfg:         dlCfg,
		antiBanCfg:  crawlerCfg,
		downloadDir: dir,
		stopCh:      make(chan struct{}),
		taskCh:      make(chan *model.DownloadTask, 100),
	}
}

// Start 启动下载引擎
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return
	}
	e.running = true

	// 确保下载目录存在
	if err := os.MkdirAll(e.downloadDir, 0755); err != nil {
		logger.Error("创建下载目录失败", zap.Error(err))
	}

	// 启动时恢复停滞任务
	if recovered, err := e.repo.RecoverStaleTasks(); err == nil && recovered > 0 {
		logger.Info("恢复停滞任务", zap.Int64("count", recovered))
	}

	// 启动 worker
	concurrency := e.cfg.Concurrency
	if concurrency < 1 {
		concurrency = 2
	}
	for i := 0; i < concurrency; i++ {
		e.workerWg.Add(1)
		go e.worker(i)
	}

	// 启动调度器
	go e.scheduler()

	logger.Info("下载引擎启动", zap.Int("workers", concurrency))
}

// Stop 停止下载引擎
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}
	e.running = false

	close(e.stopCh)
	e.workerWg.Wait()
	logger.Info("下载引擎已停止")
}

// scheduler 调度器：定期从数据库拉取排队任务
func (e *Engine) scheduler() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.dispatchTasks()
		}
	}
}

// dispatchTasks 从数据库拉取排队任务分发给 worker
func (e *Engine) dispatchTasks() {
	tasks, err := e.repo.GetQueuedTasks(e.cfg.Concurrency)
	if err != nil {
		logger.Error("获取排队任务失败", zap.Error(err))
		return
	}

	for i := range tasks {
		select {
		case e.taskCh <- &tasks[i]:
		default:
			// channel 满了，下次再来
			return
		}
	}
}

// worker 下载工作协程
func (e *Engine) worker(id int) {
	defer e.workerWg.Done()

	for {
		select {
		case <-e.stopCh:
			return
		case task := <-e.taskCh:
			e.processTask(id, task)
		}
	}
}

// processTask 处理单个下载任务
func (e *Engine) processTask(workerID int, task *model.DownloadTask) {
	token := fmt.Sprintf("worker-%d-%d", workerID, time.Now().UnixNano())

	// 获取锁
	if err := e.repo.AcquireTaskLock(task.ID, token); err != nil {
		logger.Debug("获取任务锁失败，跳过", zap.Int("task_id", task.ID), zap.Error(err))
		return
	}

	logger.Info("开始下载任务",
		zap.Int("worker", workerID),
		zap.Int("task_id", task.ID),
		zap.Int("comic_id", task.ComicID))

	// 获取图片列表
	images, err := e.repo.GetComicImages(task.ComicID)
	if err != nil || len(images) == 0 {
		e.failTask(task.ID, "获取图片列表失败")
		return
	}

	// 更新 total
	total := len(images)
	e.repo.UpdateDownloadTaskStatus(task.ID, "downloading", map[string]interface{}{
		"total": total,
	})

	// 创建漫画下载目录
	comicDir := filepath.Join(e.downloadDir, fmt.Sprintf("%d", task.ComicID))
	if err := os.MkdirAll(comicDir, 0755); err != nil {
		e.failTask(task.ID, fmt.Sprintf("创建目录失败: %v", err))
		return
	}

	// 逐张下载（从 current 位置断点续传）
	current := task.Current
	for i := current; i < total; i++ {
		// 检查是否被停止/暂停
		select {
		case <-e.stopCh:
			e.repo.ReleaseTaskLock(task.ID)
			return
		default:
		}

		// 重新读取任务状态，检查是否被暂停/取消
		freshTask, err := e.repo.GetDownloadTask(task.ID)
		if err != nil || freshTask.Status == "paused" || freshTask.Status == "canceled" {
			e.repo.ReleaseTaskLock(task.ID)
			return
		}

		img := images[i]
		destPath := filepath.Join(comicDir, fmt.Sprintf("%03d%s", img.Sort, img.Extension))

		// 文件已存在则跳过
		if _, err := os.Stat(destPath); err == nil {
			current = i + 1
			e.updateProgress(task.ID, current, total)
			continue
		}

		// 通过 singleflight 下载
		if err := e.downloadImage(img.URL, destPath); err != nil {
			logger.Warn("下载图片失败",
				zap.Int("comic_id", task.ComicID),
				zap.Int("sort", img.Sort),
				zap.Error(err))
			e.failTask(task.ID, fmt.Sprintf("下载图片 %d 失败: %v", img.Sort, err))
			return
		}

		// 更新本地路径
		e.repo.GetDB().Model(&model.ComicImage{}).
			Where("comic_id = ? AND sort = ?", task.ComicID, img.Sort).
			Update("local_path", destPath)

		current = i + 1
		e.updateProgress(task.ID, current, total)

		// 防封延迟
		e.antiBanDelay()
	}

	// 验证阶段
	e.repo.UpdateDownloadTaskStatus(task.ID, "verifying", nil)

	// 快速验证：检查文件数量
	missingCount := 0
	for _, img := range images {
		destPath := filepath.Join(comicDir, fmt.Sprintf("%03d%s", img.Sort, img.Extension))
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			missingCount++
		}
	}

	if missingCount > 0 {
		e.failTask(task.ID, fmt.Sprintf("验证失败: %d 个文件缺失", missingCount))
		return
	}

	// 完成
	now := time.Now()
	e.repo.UpdateDownloadTaskStatus(task.ID, "completed", map[string]interface{}{
		"current":     total,
		"finished_at": &now,
		"lock_token":  "",
	})

	// 更新 cache_state L3
	e.repo.UpsertCacheState(&model.CacheState{
		ComicID:    task.ComicID,
		L3Cached:   true,
		L3Total:    total,
		L3Current:  total,
		L3Progress: 100.0,
	})

	logger.Info("下载任务完成",
		zap.Int("task_id", task.ID),
		zap.Int("comic_id", task.ComicID),
		zap.Int("total", total))
}

// updateProgress 更新下载进度
func (e *Engine) updateProgress(taskID, current, total int) {
	e.repo.UpdateDownloadTaskStatus(taskID, "downloading", map[string]interface{}{
		"current": current,
	})
}

// failTask 标记任务失败
func (e *Engine) failTask(taskID int, errMsg string) {
	e.repo.UpdateDownloadTaskStatus(taskID, "failed", map[string]interface{}{
		"last_error": errMsg,
		"lock_token": "",
	})
}

// downloadImage 通过 singleflight 下载图片
func (e *Engine) downloadImage(url, destPath string) error {
	_, err, _ := e.sfGroup.Do(url, func() (interface{}, error) {
		data, err := e.crawler.DownloadImage(url)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return nil, fmt.Errorf("写入文件失败: %w", err)
		}
		return nil, nil
	})
	return err
}

// antiBanDelay 防封延迟
func (e *Engine) antiBanDelay() {
	minDelay := e.cfg.DelayMin
	maxDelay := e.cfg.DelayMax
	if minDelay <= 0 {
		minDelay = 0.5
	}
	if maxDelay <= minDelay {
		maxDelay = minDelay + 0.5
	}

	delay := minDelay + rand.Float64()*(maxDelay-minDelay)
	time.Sleep(time.Duration(delay * float64(time.Second)))
}
