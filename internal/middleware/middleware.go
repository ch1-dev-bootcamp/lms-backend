package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lms-backend/pkg/logger"
)

// Logging middleware logs HTTP requests
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Create log entry
		entry := logger.WithFields(map[string]interface{}{
			"method":      method,
			"path":        path,
			"raw_query":   raw,
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   clientIP,
			"body_size":   bodySize,
		})

		// Log based on status code
		if statusCode >= 500 {
			entry.Error("HTTP Request")
		} else if statusCode >= 400 {
			entry.Warn("HTTP Request")
		} else {
			entry.Info("HTTP Request")
		}
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		
		c.Next()
	}
}
