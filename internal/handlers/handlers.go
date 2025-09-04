package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheck handles health check requests
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]string{
		"status": "healthy",
		"service": "lms-backend",
	}
	
	json.NewEncoder(w).Encode(response)
}

// APIRoot handles API root requests
func APIRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]string{
		"message": "LMS API v1",
		"version": "1.0.0",
	}
	
	json.NewEncoder(w).Encode(response)
}