package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/services"
)

type SyncHandler struct {
	service *services.SyncService
}

type syncHeadRequest struct {
	SourceID *int64 `json:"source_id"`
	Pages    int    `json:"pages"`
}

type syncComicDetailRequest struct {
	SourceID *int64 `json:"source_id"`
}

func NewSyncHandler(service *services.SyncService) *SyncHandler {
	return &SyncHandler{service: service}
}

func (handler *SyncHandler) SyncHead(c *gin.Context) {
	var request syncHeadRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&request); err != nil {
			respondError(c, http.StatusBadRequest, err.Error(), okMeta())
			return
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()
	result, err := handler.service.SyncHead(ctx, request.SourceID, request.Pages)
	if err != nil {
		respondError(c, http.StatusBadGateway, err.Error(), Meta{Offline: false, Stale: false, Source: "remote", Message: "头部同步失败"})
		return
	}
	respond(c, http.StatusOK, result, Meta{Offline: false, Stale: false, Source: "hybrid", Message: "已完成头部同步"})
}

func (handler *SyncHandler) SyncComicDetail(c *gin.Context) {
	comicID, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	var request syncComicDetailRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&request); err != nil {
			respondError(c, http.StatusBadRequest, err.Error(), okMeta())
			return
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()
	result, err := handler.service.SyncComicDetail(ctx, request.SourceID, comicID)
	if err != nil {
		respondError(c, http.StatusBadGateway, err.Error(), Meta{Offline: false, Stale: false, Source: "remote", Message: "详情补全失败"})
		return
	}
	respond(c, http.StatusOK, result, Meta{Offline: false, Stale: false, Source: "hybrid", Message: "详情已补全"})
}
