package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"github.com/huyaaaaaa/hehuan-reader/internal/services"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

type SourcesHandler struct {
	store         *store.SourceStore
	healthService *services.SourceHealthService
	events        *services.EventBroker
}

type createSourceRequest struct {
	Name         string `json:"name" binding:"required"`
	BaseURL      string `json:"base_url" binding:"required"`
	NavigatorURL string `json:"navigator_url"`
	Priority     int    `json:"priority"`
	Enabled      *bool  `json:"enabled"`
}

type updateSourceRequest struct {
	Name         *string `json:"name"`
	BaseURL      *string `json:"base_url"`
	NavigatorURL *string `json:"navigator_url"`
	Priority     *int    `json:"priority"`
	Enabled      *bool   `json:"enabled"`
}

func NewSourcesHandler(store *store.SourceStore, healthService *services.SourceHealthService, events *services.EventBroker) *SourcesHandler {
	return &SourcesHandler{store: store, healthService: healthService, events: events}
}

func (handler *SourcesHandler) List(c *gin.Context) {
	items, err := handler.store.List()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"sources": items}, okMeta())
}

func (handler *SourcesHandler) Create(c *gin.Context) {
	var request createSourceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	item := &models.SourceSite{
		Name:         request.Name,
		BaseURL:      request.BaseURL,
		NavigatorURL: request.NavigatorURL,
		Priority:     request.Priority,
		Enabled:      enabled,
		Status:       "unknown",
	}
	if err := handler.store.Create(item); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	handler.events.Publish("source.created", gin.H{"source_id": item.ID, "name": item.Name})
	respond(c, http.StatusCreated, item, okMeta())
}

func (handler *SourcesHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid source id", okMeta())
		return
	}
	var request updateSourceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	changes := map[string]any{}
	if request.Name != nil {
		changes["name"] = *request.Name
	}
	if request.BaseURL != nil {
		changes["base_url"] = *request.BaseURL
	}
	if request.NavigatorURL != nil {
		changes["navigator_url"] = *request.NavigatorURL
	}
	if request.Priority != nil {
		changes["priority"] = *request.Priority
	}
	if request.Enabled != nil {
		changes["enabled"] = *request.Enabled
	}
	if len(changes) == 0 {
		respondError(c, http.StatusBadRequest, "empty update", okMeta())
		return
	}
	if err := handler.store.Update(id, changes); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	handler.events.Publish("source.updated", gin.H{"source_id": id})
	respond(c, http.StatusOK, gin.H{"id": id}, okMeta())
}

func (handler *SourcesHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid source id", okMeta())
		return
	}
	if err := handler.store.Delete(id); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	handler.events.Publish("source.deleted", gin.H{"source_id": id})
	respond(c, http.StatusOK, gin.H{"id": id}, okMeta())
}

func (handler *SourcesHandler) Check(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid source id", okMeta())
		return
	}
	item, err := handler.store.Get(id)
	if err != nil {
		respondError(c, http.StatusNotFound, err.Error(), okMeta())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	check, checkErr := handler.healthService.CheckSource(ctx, item, "manual")
	if checkErr != nil {
		respondError(c, http.StatusBadGateway, checkErr.Error(), Meta{Offline: false, Stale: false, Source: "remote", Message: "源站当前不可用"})
		return
	}
	respond(c, http.StatusOK, check, okMeta())
}
