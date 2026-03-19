package api

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ImportExportHandler 导入导出处理器
type ImportExportHandler struct {
	ieService *service.ImportExportService
}

// NewImportExportHandler 创建导入导出处理器
func NewImportExportHandler(ieService *service.ImportExportService) *ImportExportHandler {
	return &ImportExportHandler{ieService: ieService}
}

// CreateExport 创建导出任务
// POST /api/v2/export
func (h *ImportExportHandler) CreateExport(c *gin.Context) {
	var req model.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	job, err := h.ieService.CreateExport(req.Scope, req.ComicIDs)
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, gin.H{
		"job_id": job.ID,
		"status": job.Status,
	})
}

// UploadImport 上传导入包到服务端临时目录
// POST /api/v2/import/upload
func (h *ImportExportHandler) UploadImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "缺少导入文件")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "打开上传文件失败")
		return
	}
	defer src.Close()

	filePath, err := h.ieService.SaveImportUpload(src, fileHeader.Filename)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, err.Error())
		return
	}

	Success(c, gin.H{
		"file_path": filePath,
		"filename":  fileHeader.Filename,
	})
}

// GetExportStatus 查询导出任务状态
// GET /api/v2/export/:id
func (h *ImportExportHandler) GetExportStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的任务ID")
		return
	}

	job, err := h.ieService.GetJob(id)
	if err != nil {
		Error(c, http.StatusNotFound, CodeNotFound, "任务不存在")
		return
	}

	resp := gin.H{
		"id":        job.ID,
		"direction": job.Direction,
		"scope":     job.Scope,
		"status":    job.Status,
	}

	if job.FilePath != "" {
		resp["file_path"] = job.FilePath
		if info, err := os.Stat(job.FilePath); err == nil {
			resp["file_size"] = info.Size()
		}
	}
	if job.SummaryJSON != "{}" && job.SummaryJSON != "" {
		resp["summary"] = job.SummaryJSON
	}
	if job.LastError != "" {
		resp["last_error"] = job.LastError
	}

	Success(c, resp)
}

// DownloadExport 下载导出文件
// GET /api/v2/export/:id/download
func (h *ImportExportHandler) DownloadExport(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的任务ID")
		return
	}

	filePath, err := h.ieService.GetExportFilePath(id)
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	filename := filepath.Base(filePath)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "application/zip")
	c.File(filePath)
}

// ScanImport 预扫描导入包
// POST /api/v2/import/scan
func (h *ImportExportHandler) ScanImport(c *gin.Context) {
	var req model.ImportScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	result, err := h.ieService.ScanImport(req.FilePath)
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, result)
}

// ExecuteImport 执行导入
// POST /api/v2/import/execute
func (h *ImportExportHandler) ExecuteImport(c *gin.Context) {
	var req model.ImportExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	job, err := h.ieService.ExecuteImport(req.FilePath, req.Strategy)
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, gin.H{
		"job_id": job.ID,
		"status": job.Status,
	})
}

// GetImportStatus 查询导入任务状态
// GET /api/v2/import/:id
func (h *ImportExportHandler) GetImportStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的任务ID")
		return
	}

	job, err := h.ieService.GetJob(id)
	if err != nil {
		Error(c, http.StatusNotFound, CodeNotFound, "任务不存在")
		return
	}

	resp := gin.H{
		"id":             job.ID,
		"direction":      job.Direction,
		"status":         job.Status,
		"conflict_count": job.ConflictCount,
	}
	if job.SummaryJSON != "{}" && job.SummaryJSON != "" {
		resp["summary"] = job.SummaryJSON
	}
	if job.LastError != "" {
		resp["last_error"] = job.LastError
	}

	Success(c, resp)
}
