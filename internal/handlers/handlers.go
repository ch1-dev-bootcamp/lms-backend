package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lms-backend/internal/database"
)

// HealthCheck handles health check requests
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "lms-backend",
	})
}

// APIRoot handles API root requests
func APIRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "LMS API v1",
		"version": "1.0.0",
	})
}

// DatabaseHealth handles database health check requests
func DatabaseHealth(c *gin.Context) {
	if err := database.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"service":   "database",
			"error":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "database",
	})
}