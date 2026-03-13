package cache

import (
	"bytes"
	"comic-viewer-claude/internal/crawler"
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

// CoverCacheManager 封面缓存管理器
type CoverCacheManager struct {
	repo           *repository.Repository
	crawler        *crawler.Client
	maxCovers      int
	targetSizeKB   int
	isRunning      bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewCoverCacheManager 创建封面缓存管理器
func NewCoverCacheManager(repo *repository.Repository, crawler *crawler.Client, maxCovers, targetSizeKB int) *CoverCacheManager {
	return &CoverCacheManager{
		repo:         repo,
		crawler:      crawler,
		maxCovers:    maxCovers,
		targetSizeKB: targetSizeKB,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动后台工作线程
func (m *CoverCacheManager) Start() {
	if m.isRunning {
		return
	}

	m.isRunning = true
	m.wg.Add(1)
	go m.workerLoop()
	logger.Info("封面缓存管理器已启动")
}

// Stop 停止后台工作线程
func (m *CoverCacheManager) Stop() {
	if !m.isRunning {
		return
	}

	m.isRunning = false
	close(m.stopChan)
	m.wg.Wait()
	logger.Info("封面缓存管理器已停止")
}

// workerLoop 后台工作循环
func (m *CoverCacheManager) workerLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.processBatch()
		}
	}
}

// processBatch 处理一批任务
func (m *CoverCacheManager) processBatch() {
	// 获取待处理的任务（一次20个）
	tasks, err := m.repo.GetPendingCoverTasks(20)
	if err != nil {
		logger.Error("获取封面缓存任务失败", zap.Error(err))
		return
	}

	if len(tasks) == 0 {
		return
	}

	logger.Info("开始处理封面缓存任务", zap.Int("count", len(tasks)))

	for _, task := range tasks {
		if !m.isRunning {
		break
		}

		if err := m.processTask(task); err != nil {
			logger.Error("处理封面任务失败",
				zap.Int("comic_id", task.ComicID),
				zap.String("title", task.Title),
				zap.Error(err))

			// 更新重试次数
			retryCount := task.RetryCount + 1
			if retryCount >= 3 {
		// 重试3次后删除任务
				m.repo.DeleteCoverTask(task.ComicID)
			} else {
				m.repo.UpdateCoverTaskStatus(task.ComicID, "pending", retryCount)
			}
		}

		// 控制请求频率
		time.Sleep(300 * time.Millisecond)
	}
}

// processTask 处理单个封面缓存任务
func (m *CoverCacheManager) processTask(task model.CoverCacheQueue) error {
	// 检查是否已缓存
	comic, err := m.repo.GetComic(task.ComicID)
	if err == nil && comic.HasCoverCached {
		// 已缓存，删除任务
		return m.repo.DeleteCoverTask(task.ComicID)
	}

	// 下载封面
	logger.Debug("下载封面", zap.String("title", task.Title))
	imageData, err := m.crawler.DownloadImage(task.CoverURL)
	if err != nil {
		return fmt.Errorf("下载封面失败: %w", err)
	}

	// 压缩封面
	compressedData, err := m.compressImage(imageData)
	if err != nil {
		return fmt.Errorf("压缩封面失败: %w", err)
	}

	// Base64编码
	coverBase64 := base64.StdEncoding.EncodeToString(compressedData)

	// 更新数据库
	now := time.Now()
	if comic == nil {
		comic = &model.Comic{ID: task.ComicID}
	}
	comic.CoverBase64 = coverBase64
	comic.HasCoverCached = true
	comic.CoverCachedAt = &now
	comic.CoverSize = len(compressedData)

	if err := m.repo.SaveComic(comic); err != nil {
		return fmt.Errorf("保存封面失败: %w", err)
	}

	// 删除任务
	if err := m.repo.DeleteCoverTask(task.ComicID); err != nil {
		logger.Warn("删除封面任务失败", zap.Error(err))
	}

	logger.Debug("封面缓存成功",
		zap.Int("comic_id", task.ComicID),
		zap.Int("size_kb", len(compressedData)/1024))

	return nil
}

// compressImage 压缩图片到目标大小
func (m *CoverCacheManager) compressImage(data []byte) ([]byte, error) {
	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	// 调整大小（高度300px）
	resized := imaging.Resize(img, 0, 300, imaging.Lanczos)

	// 尝试不同的质量级别，直到达到目标大小
	targetSize := m.targetSizeKB * 1024
	quality := 85
	for quality >= 20 {
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, resized, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("编码图片失败: %w", err)
		}

		if buf.Len() <= targetSize || quality <= 20 {
			return buf.Bytes(), nil
		}

		quality -= 5
	}

	// 如果还是太大，进一步缩小尺寸
	resized = imaging.Resize(img, 0, 200, imaging.Lanczos)
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, resized, &jpeg.Options{Quality: 75}); err != nil {
		return nil, fmt.Errorf("编码图片失败: %w", err)
	}

	return buf.Bytes(), nil
}

// AddTask 添加封面缓存任务
func (m *CoverCacheManager) AddTask(comicID int, title, coverURL string, priority int) error {
	return m.repo.AddCoverCacheTask(comicID, title, coverURL, priority)
}
