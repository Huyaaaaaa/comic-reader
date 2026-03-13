package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码常量
const (
	CodeOK         = 0
	CodeBadRequest = 400
	CodeNotFound   = 404
	CodeInternal   = 500
)

// APIResponse 标准 JSON 信封
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PagedData 分页数据
type PagedData struct {
	Items      any   `json:"items"`
	Page       int   `json:"page"`
	TotalPages int   `json:"total_pages"`
	Total      int64 `json:"total"`
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeOK,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus, code int, msg string) {
	c.JSON(httpStatus, APIResponse{
		Code:    code,
		Message: msg,
	})
}

// PagedSuccess 分页成功响应
func PagedSuccess(c *gin.Context, items any, page, totalPages int, total int64) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeOK,
		Message: "success",
		Data: PagedData{
			Items:      items,
			Page:       page,
			TotalPages: totalPages,
			Total:      total,
		},
	})
}
