package api

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DownloadHandler 下载管理处理器
type DownloadHandler struct {
	downloadService *service.DownloadService
}

// NewDownloadHandler 创建下载处理器
func NewDownloadHandler(downloadService *service.DownloadService) *DownloadHandler {
	return &DownloadHandler{downloadService: downloadService}
}

// CreateTask 创建下载任务
// POST /api/v2/downloads
func (h *DownloadHandler) CreateTask(c *gin.Context) {
	var req model.CreateDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	task, err := h.downloadService.CreateTask(req.ComicID, req.TaskType)
	if err != nil {
		Error(c, http.StatusConflict, CodeBadRequest, err.Error())
		return
	}

	Success(c, task)
}

// Precheck 下载预检查
// POST /api/v2/downloads/precheck
func (h *DownloadHandler) Precheck(c *gin.Context) {
	var req model.DownloadPrecheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	result, err := h.downloadService.Precheck(req.ComicIDs)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	Success(c, result)
}

// GetTasks 获取下载任务列表
// GET /api/v2/downloads?group=active
func (h *DownloadHandler) GetTasks(c *gin.Context) {
	group := c.DefaultQuery("group", "")

	tasks, err := h.downloadService.GetTasks(group)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	Success(c, tasks)
}

// PauseTask 暂停下载任务
// POST /api/v2/downloads/:id/pause
func (h *DownloadHandler) PauseTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的任务ID")
		return
	}

	if err := h.downloadService.PauseTask(id); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, nil)
}

// ResumeTask 恢复下载任务
// POST /api/v2/downloads/:id/resume
func (h *DownloadHandler) ResumeTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的任务ID")
		return
	}

	if err := h.downloadService.ResumeTask(id); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, nil)
}

// CancelTask 取消下载任务
// POST /api/v2/downloads/:id/cancel
func (h *DownloadHandler) CancelTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的任务ID")
		return
	}

	if err := h.downloadService.CancelTask(id); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, nil)
}

// RetryTask 重试下载任务
// POST /api/v2/downloads/:id/retry
func (h *DownloadHandler) RetryTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的任务ID")
		return
	}

	if err := h.downloadService.RetryTask(id); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, nil)
}
