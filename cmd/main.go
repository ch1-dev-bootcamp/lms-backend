package main

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/lms-backend/internal/database"
	"github.com/your-org/lms-backend/internal/handlers"
	"github.com/your-org/lms-backend/internal/middleware"
	"github.com/your-org/lms-backend/pkg/config"
	"github.com/your-org/lms-backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Initialize(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	logger.Info("Starting LMS Backend application")

	// Initialize database connection
	logger.Info("Initializing database connection...")
	dbConfig := database.NewConnectionConfig(cfg)
	if err := database.Connect(dbConfig); err != nil {
		logger.Fatal("Failed to connect to database", logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Data)
	}
	defer database.Close()

	logger.Info("Database connected successfully!")

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.Logging())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	// Health check endpoints
	r.GET("/health", handlers.HealthCheck)
	r.GET("/health/database", handlers.DatabaseHealth)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", handlers.APIRoot)
		// Future API endpoints will be added here
	}

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	logger.WithField("port", port).Info("Starting LMS server")
	logger.Fatal("Server stopped", logger.WithFields(map[string]interface{}{
		"error": r.Run(":" + port).Error(),
	}).Data)
}
