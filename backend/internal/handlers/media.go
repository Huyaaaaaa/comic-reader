package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/services"
	"gorm.io/gorm"
)

type MediaHandler struct {
	service *services.ImageProxyService
}

func NewMediaHandler(service *services.ImageProxyService) *MediaHandler {
	return &MediaHandler{service: service}
}

func (handler *MediaHandler) ProxyCover(c *gin.Context) {
	comicID, err := strconv.ParseInt(strings.TrimSpace(c.Query("comic_id")), 10, 64)
	if err != nil || comicID <= 0 {
		c.String(http.StatusBadRequest, "invalid comic_id")
		return
	}

	path, err := handler.service.EnsureCoverLocal(c.Request.Context(), comicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "cover not found")
			return
		}
		c.String(http.StatusBadGateway, err.Error())
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.File(path)
}

func (handler *MediaHandler) ProxyImage(c *gin.Context) {
	comicID, err := strconv.ParseInt(strings.TrimSpace(c.Query("comic_id")), 10, 64)
	if err != nil || comicID <= 0 {
		c.String(http.StatusBadRequest, "invalid comic_id")
		return
	}
	sortIndex, err := strconv.Atoi(strings.TrimSpace(c.Query("sort")))
	if err != nil || sortIndex < 0 {
		c.String(http.StatusBadRequest, "invalid sort")
		return
	}

	path, err := handler.service.EnsureImageLocal(c.Request.Context(), comicID, sortIndex)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "image not found")
			return
		}
		c.String(http.StatusBadGateway, err.Error())
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.File(path)
}
