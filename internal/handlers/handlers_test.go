package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a Gin router
	r := gin.New()
	r.GET("/health", HealthCheck)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Perform the request
	r.ServeHTTP(w, req)
	
	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	assert.Contains(t, w.Body.String(), "lms-backend")
}

func TestAPIRoot(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a Gin router
	r := gin.New()
	r.GET("/api/v1/", APIRoot)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/api/v1/", nil)
	w := httptest.NewRecorder()
	
	// Perform the request
	r.ServeHTTP(w, req)
	
	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "LMS API v1")
	assert.Contains(t, w.Body.String(), "1.0.0")
}
