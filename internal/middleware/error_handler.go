package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	apperrors "github.com/your-org/lms-backend/internal/errors"
	"github.com/your-org/lms-backend/pkg/logger"
)

// ErrorHandler is a centralized error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
					// Handle different types of errors
		switch e := err.Err.(type) {
		case *apperrors.AppError:
			handleAppError(c, e)
		case error:
			handleGenericError(c, e)
		default:
			handleUnknownError(c, err.Err)
		}
		}
	}
}

// handleAppError handles application-specific errors
func handleAppError(c *gin.Context, appErr *apperrors.AppError) {
	// Set request ID if not already set
	if appErr.RequestID == "" {
		requestID := getRequestID(c)
		appErr.SetRequestID(requestID)
	}

	// Log the error
	logError(c, appErr)

	// Send error response
	c.JSON(appErr.HTTPStatus, appErr.ToResponse())
}

// handleGenericError handles generic Go errors
func handleGenericError(c *gin.Context, err error) {
	requestID := getRequestID(c)
	
	// Convert generic error to AppError
	appErr := apperrors.NewInternalError(err, "request processing").
		SetRequestID(requestID)

	// Log the error
	logError(c, appErr)

	// Send error response
	c.JSON(appErr.HTTPStatus, appErr.ToResponse())
}

// handleUnknownError handles unknown error types
func handleUnknownError(c *gin.Context, _ interface{}) {
	requestID := getRequestID(c)
	
	// Create a generic internal error
	appErr := apperrors.NewInternalError(
		errors.New("unknown error type"),
		"request processing",
	).SetRequestID(requestID)

	// Log the error
	logError(c, appErr)

	// Send error response
	c.JSON(appErr.HTTPStatus, appErr.ToResponse())
}

// getRequestID gets or generates a request ID
func getRequestID(c *gin.Context) string {
	// Try to get request ID from header
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Try to get from context
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	// Generate new request ID
	requestID := apperrors.GenerateRequestID()
	c.Set("request_id", requestID)
	return requestID
}

// logError logs the error with appropriate level
func logError(c *gin.Context, appErr *apperrors.AppError) {
	// Create log entry with context
	entry := logger.WithFields(map[string]interface{}{
		"request_id":    appErr.RequestID,
		"error_code":    appErr.Code,
		"error_message": appErr.Message,
		"error_details": appErr.Details,
		"http_status":   appErr.HTTPStatus,
		"method":        c.Request.Method,
		"path":          c.Request.URL.Path,
		"client_ip":     c.ClientIP(),
		"user_agent":    c.Request.UserAgent(),
	})

	// Add cause if available
	if appErr.Cause != nil {
		entry = entry.WithField("cause", appErr.Cause.Error())
	}

	// Log based on HTTP status
	switch {
	case appErr.HTTPStatus >= 500:
		entry.Error("Application Error")
	case appErr.HTTPStatus >= 400:
		entry.Warn("Client Error")
	default:
		entry.Info("Application Error")
	}
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := getRequestID(c)
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ErrorHandlerFunc is a helper function to handle errors in handlers
func ErrorHandlerFunc(c *gin.Context, err error) {
	// Convert error to AppError if it's not already
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		// Error is already an AppError
	} else {
		appErr = apperrors.NewInternalError(err, "handler processing")
	}

	// Set request ID
	requestID := getRequestID(c)
	appErr.SetRequestID(requestID)

	// Add error to Gin context
	c.Error(appErr)
}

// AbortWithError aborts the request with an error
func AbortWithError(c *gin.Context, err *apperrors.AppError) {
	requestID := getRequestID(c)
	appErr := err.SetRequestID(requestID)
	
	// Log the error
	logError(c, appErr)
	
	// Abort with error
	c.AbortWithStatusJSON(appErr.HTTPStatus, appErr.ToResponse())
}

// AbortWithValidationError aborts with validation error
func AbortWithValidationError(c *gin.Context, message string) {
	AbortWithError(c, apperrors.NewValidationError(message))
}

// AbortWithValidationErrorWithDetails aborts with validation error and details
func AbortWithValidationErrorWithDetails(c *gin.Context, message, details string) {
	AbortWithError(c, apperrors.NewValidationErrorWithDetails(message, details))
}

// AbortWithNotFoundError aborts with not found error
func AbortWithNotFoundError(c *gin.Context, message string) {
	AbortWithError(c, apperrors.NewNotFoundError(message))
}

// AbortWithConflictError aborts with conflict error
func AbortWithConflictError(c *gin.Context, message string) {
	AbortWithError(c, apperrors.NewConflictError(message))
}

// AbortWithInternalError aborts with internal error
func AbortWithInternalError(c *gin.Context, err error, operation string) {
	AbortWithError(c, apperrors.NewInternalError(err, operation))
}
