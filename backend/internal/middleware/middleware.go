package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS(allowOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否在允许列表中
		allowed := false
		for _, allowOrigin := range allowOrigins {
			if origin == allowOrigin || allowOrigin == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		// 简单的日志输出
		if statusCode >= 400 {
			c.Writer.Header().Set("X-Request-Time", latency.String())
		}

		// 这里可以使用 zap logger 记录更详细的日志
		_ = latency
		_ = statusCode
		_ = method
		_ = clientIP
		_ = path
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}
