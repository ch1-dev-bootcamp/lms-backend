package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthCheck)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that the response contains the expected fields
	body := strings.TrimSpace(rr.Body.String())
	expected := `{"service":"lms-backend","status":"healthy"}`
	if body != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", body, expected)
	}
}

func TestAPIRoot(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(APIRoot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that the response contains the expected fields
	body := strings.TrimSpace(rr.Body.String())
	expected := `{"message":"LMS API v1","version":"1.0.0"}`
	if body != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", body, expected)
	}
}
