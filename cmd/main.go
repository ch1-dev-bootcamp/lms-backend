package main

import (
	"log"
	"net/http"

	"github.com/your-org/lms-backend/internal/handlers"
	"github.com/your-org/lms-backend/internal/middleware"
	"github.com/your-org/lms-backend/pkg/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handlers.HealthCheck)

	// API routes (placeholder for now)
	mux.HandleFunc("/api/v1/", handlers.APIRoot)

	// Apply middleware
	handler := middleware.Logging(middleware.CORS(mux))

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting LMS server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
