package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/services"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

type SettingsHandler struct {
	store  *store.SettingsStore
	events *services.EventBroker
}

type updateSettingRequest struct {
	Value interface{} `json:"value" binding:"required"`
}

func NewSettingsHandler(store *store.SettingsStore, events *services.EventBroker) *SettingsHandler {
	return &SettingsHandler{store: store, events: events}
}

func (handler *SettingsHandler) List(c *gin.Context) {
	items, err := handler.store.List()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	values := make(map[string]string, len(items))
	for key, item := range items {
		values[key] = item.ValueJSON
	}
	respond(c, http.StatusOK, values, okMeta())
}

func (handler *SettingsHandler) Update(c *gin.Context) {
	key := c.Param("key")
	var request updateSettingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	encoded, err := json.Marshal(request.Value)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	if err := handler.store.Update(key, string(encoded)); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	handler.events.Publish("settings.updated", gin.H{"key": key, "value": request.Value})
	respond(c, http.StatusOK, gin.H{"key": key}, okMeta())
}
