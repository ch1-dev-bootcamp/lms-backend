package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lms-backend/internal/database"
	"github.com/your-org/lms-backend/internal/handlers"
	"github.com/your-org/lms-backend/internal/middleware"
	"github.com/your-org/lms-backend/pkg/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	log.Println("Initializing database connection...")
	dbConfig := database.NewConnectionConfig(cfg)
	if err := database.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Database connected successfully!")

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

	log.Printf("Starting LMS server on port %s", port)
	log.Fatal(r.Run(":" + port))
}
