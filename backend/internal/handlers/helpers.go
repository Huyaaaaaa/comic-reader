package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Offline bool   `json:"offline"`
	Stale   bool   `json:"stale"`
	Source  string `json:"source"`
	Message string `json:"message,omitempty"`
}

func respond(c *gin.Context, status int, data interface{}, meta Meta) {
	c.JSON(status, gin.H{"data": data, "meta": meta})
}

func respondError(c *gin.Context, status int, message string, meta Meta) {
	c.JSON(status, gin.H{"error": message, "meta": meta})
}

func okMeta() Meta {
	return Meta{Offline: false, Stale: false, Source: "local_cache"}
}

func noContentMeta() Meta {
	return Meta{Offline: false, Stale: false, Source: "local_cache", Message: "暂无数据"}
}

const defaultUserID = 1

var _ = http.StatusOK
