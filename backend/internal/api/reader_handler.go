package api

import (
	"comic-viewer-claude/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReaderHandler struct {
	readerService *service.ReaderSessionService
}

func NewReaderHandler(readerService *service.ReaderSessionService) *ReaderHandler {
	return &ReaderHandler{readerService: readerService}
}

func (h *ReaderHandler) CreateSession(c *gin.Context) {
	var req struct {
		ComicID int `json:"comic_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	session, err := h.readerService.CreateSession(req.ComicID)
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	Success(c, session)
}

func (h *ReaderHandler) GetSession(c *gin.Context) {
	session, err := h.readerService.GetSession(c.Param("id"))
	if err != nil {
		Error(c, http.StatusNotFound, CodeNotFound, err.Error())
		return
	}

	Success(c, session)
}

func (h *ReaderHandler) UpdateFocus(c *gin.Context) {
	var req struct {
		Sort int `json:"sort" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	if err := h.readerService.UpdateFocus(c.Param("id"), req.Sort); err != nil {
		Error(c, http.StatusNotFound, CodeNotFound, err.Error())
		return
	}

	Success(c, nil)
}

func (h *ReaderHandler) ServeImage(c *gin.Context) {
	sortValue, err := strconv.Atoi(c.Param("sort"))
	if err != nil {
		Error(c, http.StatusBadRequest, CodeBadRequest, "无效的页码")
		return
	}

	path, err := h.readerService.ServeImagePath(c.Param("id"), sortValue, 3*time.Second)
	if err != nil {
		Error(c, http.StatusTooEarly, CodeBadRequest, err.Error())
		return
	}

	c.File(path)
}
