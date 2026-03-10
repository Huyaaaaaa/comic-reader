package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/services"
)

type EventsHandler struct {
	broker *services.EventBroker
}

func NewEventsHandler(broker *services.EventBroker) *EventsHandler {
	return &EventsHandler{broker: broker}
}

func (handler *EventsHandler) Stream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	channel := handler.broker.Subscribe()
	defer handler.broker.Unsubscribe(channel)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		respondError(c, http.StatusInternalServerError, "stream unsupported", okMeta())
		return
	}

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-channel:
			payload, _ := json.Marshal(event)
			fmt.Fprintf(c.Writer, "event: %s\n", event.Type)
			fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
