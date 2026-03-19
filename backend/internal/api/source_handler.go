package api

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/source"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SourceHandler 源站处理器
type SourceHandler struct {
	sourceManager *source.Manager
}

// NewSourceHandler 创建源站处理器
func NewSourceHandler(sourceManager *source.Manager) *SourceHandler {
	return &SourceHandler{sourceManager: sourceManager}
}

// GetSources GET /api/v2/sources — 获取所有源站
func (h *SourceHandler) GetSources(c *gin.Context) {
	sources := h.sourceManager.GetAll()
	Success(c, sources)
}

// AddSource POST /api/v2/sources — 添加源站
func (h *SourceHandler) AddSource(c *gin.Context) {
	var req model.AddSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	s, err := h.sourceManager.Add(req.URL, req.Name, req.ImageCDN)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "添加源站失败: "+err.Error())
		return
	}

	Success(c, s)
}

// ImportReleasePageSources POST /api/v2/sources/import-release — 从发布页导入源站
func (h *SourceHandler) ImportReleasePageSources(c *gin.Context) {
	var req model.ImportReleasePageSourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	result, err := h.sourceManager.ImportFromReleasePage(req.ReleasePageURL)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "导入发布页源站失败: "+err.Error())
		return
	}

	Success(c, result)
}

// DeleteSource DELETE /api/v2/sources/:id — 删除源站
func (h *SourceHandler) DeleteSource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的源站 ID")
		return
	}

	if err := h.sourceManager.Remove(id); err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "删除源站失败: "+err.Error())
		return
	}

	Success(c, nil)
}

// CheckSourceHealth POST /api/v2/sources/:id/check — 手动健康检查
func (h *SourceHandler) CheckSourceHealth(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的源站 ID")
		return
	}

	sources := h.sourceManager.GetAll()
	var target *model.SourceSite
	for _, s := range sources {
		if s.ID == id {
			target = s
			break
		}
	}
	if target == nil {
		Error(c, http.StatusNotFound, CodeNotFound, "源站不存在")
		return
	}

	resp, checkErr := h.sourceManager.CheckHealth(target)
	if resp == nil {
		resp = &model.SourceHealthResponse{
			ID:     target.ID,
			URL:    target.URL,
			Status: "inactive",
		}
	}
	if checkErr != nil {
		resp.Error = checkErr.Error()
	}

	Success(c, resp)
}

// CheckProxyHealth GET /api/v2/sources/proxy/check — 检测代理可用性
func (h *SourceHandler) CheckProxyHealth(c *gin.Context) {
	force := c.DefaultQuery("force", "false") == "true"
	Success(c, h.sourceManager.GetProxyHealth(force))
}
