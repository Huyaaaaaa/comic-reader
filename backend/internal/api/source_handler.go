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

	latency, checkErr := h.sourceManager.CheckHealth(target)
	resp := model.SourceHealthResponse{
		ID:      target.ID,
		URL:     target.URL,
		Latency: latency,
	}
	if checkErr != nil {
		resp.Status = "failed"
		resp.Error = checkErr.Error()
	} else {
		resp.Status = "active"
	}

	Success(c, resp)
}
