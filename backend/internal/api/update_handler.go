package api

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateHandler 更新处理器
type UpdateHandler struct {
	updateService *service.UpdateService
}

// NewUpdateHandler 创建更新处理器
func NewUpdateHandler(updateService *service.UpdateService) *UpdateHandler {
	return &UpdateHandler{updateService: updateService}
}

// CheckContentUpdate 检查内容更新
func (h *UpdateHandler) CheckContentUpdate(c *gin.Context) {
	var req model.ContentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body，使用默认值
		req = model.ContentUpdateRequest{}
	}

	resp, err := h.updateService.CheckContentUpdate(req.Pages, req.Mode)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "检查内容更新失败: "+err.Error())
		return
	}

	Success(c, resp)
}

// CheckAppUpdate 检查应用更新
func (h *UpdateHandler) CheckAppUpdate(c *gin.Context) {
	resp, err := h.updateService.CheckAppUpdate()
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "检查应用更新失败: "+err.Error())
		return
	}

	Success(c, resp)
}
