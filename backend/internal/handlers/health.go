package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/config"
)

type HealthHandler struct {
	cfg config.Config
}

func NewHealthHandler(cfg config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (handler *HealthHandler) Health(c *gin.Context) {
	respond(c, http.StatusOK, gin.H{
		"name":    handler.cfg.AppName,
		"version": handler.cfg.AppVersion,
		"status":  "ok",
	}, okMeta())
}
