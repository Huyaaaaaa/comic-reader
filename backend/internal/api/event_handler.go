package api

import (
	"comic-viewer-claude/internal/event"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EventHandler SSE 事件处理器
type EventHandler struct {
	eventService *event.Service
}

// NewEventHandler 创建事件处理器
func NewEventHandler(eventService *event.Service) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// StreamEvents GET /api/events — SSE 连接端点
func (h *EventHandler) StreamEvents(c *gin.Context) {
	clientID := uuid.New().String()
	eventChan := h.eventService.Register(clientID)
	defer h.eventService.Unregister(clientID)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	c.Stream(func(w io.Writer) bool {
		select {
		case evt, ok := <-eventChan:
			if !ok {
				return false
			}
			data, _ := json.Marshal(evt)
			c.SSEvent("message", string(data))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
