package api

import (
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/config"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// StorageHandler 存储管理处理器
type StorageHandler struct {
	repo  *repository.Repository
	dlCfg *config.DownloadConfig
}

// NewStorageHandler 创建存储管理处理器
func NewStorageHandler(repo *repository.Repository, dlCfg *config.DownloadConfig) *StorageHandler {
	return &StorageHandler{repo: repo, dlCfg: dlCfg}
}

// GetStats 获取存储统计
// GET /api/v2/storage/stats
func (h *StorageHandler) GetStats(c *gin.Context) {
	_, l1, l2, l3, err := h.repo.GetCacheOverview()
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	dbPath := h.repo.GetDBPath()
	dbSize := getFileSizeMB(dbPath)

	dlDir := h.dlCfg.Dir
	if dlDir == "" {
		dlDir = "./downloads"
	}
	dlSize := getDirSizeMB(dlDir)

	Success(c, gin.H{
		"db_size_mb":       dbSize,
		"download_size_mb": dlSize,
		"l1_count":         l1,
		"l2_count":         l2,
		"l3_count":         l3,
		"l3_size_mb":       dlSize,
		"total_size_mb":    dbSize + dlSize,
	})
}

// ClearStorage 清除缓存
// POST /api/v2/storage/clear
func (h *StorageHandler) ClearStorage(c *gin.Context) {
	var req struct {
		Level string `json:"level"` // l1, l2, l3, 或空（全部）
	}
	c.ShouldBindJSON(&req)

	switch req.Level {
	case "l3":
		// 清除下载文件
		dlDir := h.dlCfg.Dir
		if dlDir == "" {
			dlDir = "./downloads"
		}
		os.RemoveAll(dlDir)
		os.MkdirAll(dlDir, 0755)
		// 重置 L3 缓存状态
		h.repo.ResetCacheLevel("l3")
	case "l2":
		h.repo.ResetCacheLevel("l2")
	case "l1":
		h.repo.ResetCacheLevel("l1")
	default:
		// 清除全部
		dlDir := h.dlCfg.Dir
		if dlDir == "" {
			dlDir = "./downloads"
		}
		os.RemoveAll(dlDir)
		os.MkdirAll(dlDir, 0755)
		h.repo.ResetAllCacheLevels()
	}

	Success(c, nil)
}

func getFileSizeMB(path string) int {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return int(info.Size() / 1024 / 1024)
}

func getDirSizeMB(path string) int {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})
	return int(size / 1024 / 1024)
}
