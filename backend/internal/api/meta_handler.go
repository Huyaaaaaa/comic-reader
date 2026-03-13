package api

import (
	"comic-viewer-claude/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MetaHandler 元数据处理器（作者、标签、分类）
type MetaHandler struct {
	repo *repository.Repository
}

// NewMetaHandler 创建元数据处理器
func NewMetaHandler(repo *repository.Repository) *MetaHandler {
	return &MetaHandler{repo: repo}
}

// GetAuthors GET /api/v2/authors — 获取去重作者列表
func (h *MetaHandler) GetAuthors(c *gin.Context) {
	authors, err := h.repo.GetDistinctAuthors()
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "获取作者列表失败: "+err.Error())
		return
	}
	Success(c, authors)
}

// GetAuthorComics GET /api/v2/authors/:id/comics — 获取同作者作品
func (h *MetaHandler) GetAuthorComics(c *gin.Context) {
	authorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的作者 ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	comics, total, err := h.repo.GetComicsByAuthorID(authorID, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "获取作者作品失败: "+err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	PagedSuccess(c, comics, page, totalPages, total)
}

// GetTags GET /api/v2/tags — 获取标签列表
func (h *MetaHandler) GetTags(c *gin.Context) {
	tags, err := h.repo.GetAllTags()
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "获取标签列表失败: "+err.Error())
		return
	}
	Success(c, tags)
}

// GetCategories GET /api/v2/categories — 获取分类列表
func (h *MetaHandler) GetCategories(c *gin.Context) {
	categories, err := h.repo.GetCategories()
	if err != nil {
		Error(c, http.StatusInternalServerError, CodeInternal, "获取分类列表失败: "+err.Error())
		return
	}
	Success(c, categories)
}
