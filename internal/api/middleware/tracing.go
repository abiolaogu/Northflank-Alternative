// Package middleware provides request tracing
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey is the context key for request ID
const RequestIDKey = "X-Request-ID"

// Tracing provides request tracing middleware
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get or generate request ID
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDKey, requestID)

		// Record request timing
		start := time.Now()

		c.Next()

		// Add timing header
		latency := time.Since(start)
		c.Header("X-Response-Time", latency.String())
	}
}

// RequestLogger logs request details
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Log fields available for structured logging
		_ = map[string]interface{}{
			"request_id": c.GetString(RequestIDKey),
			"status":     status,
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency":    latency.String(),
			"latency_ms": latency.Milliseconds(),
			"size":       c.Writer.Size(),
		}
	}
}
