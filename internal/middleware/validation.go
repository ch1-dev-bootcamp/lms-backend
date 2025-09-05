package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/your-org/lms-backend/internal/models"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string           `json:"error"`
	Message string           `json:"message"`
	Details []ValidationError `json:"details"`
}

// ValidateRequest validates request body against a struct
func ValidateRequest[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		
		// Bind JSON to struct
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid JSON format",
				Code:    400,
			})
			c.Abort()
			return
		}

		// Validate struct
		if err := validate.Struct(req); err != nil {
			var validationErrors []ValidationError
			
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, ValidationError{
					Field:   err.Field(),
					Message: getValidationMessage(err),
					Value:   err.Value().(string),
				})
			}

			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Error:   "validation_failed",
				Message: "Request validation failed",
				Details: validationErrors,
			})
			c.Abort()
			return
		}

		// Store validated request in context
		c.Set("validated_request", req)
		c.Next()
	}
}

// ValidateQuery validates query parameters
func ValidateQuery[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		
		// Bind query parameters to struct
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_query",
				Message: "Invalid query parameters",
				Code:    400,
			})
			c.Abort()
			return
		}

		// Validate struct
		if err := validate.Struct(req); err != nil {
			var validationErrors []ValidationError
			
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, ValidationError{
					Field:   err.Field(),
					Message: getValidationMessage(err),
					Value:   err.Value().(string),
				})
			}

			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Error:   "validation_failed",
				Message: "Query validation failed",
				Details: validationErrors,
			})
			c.Abort()
			return
		}

		// Store validated request in context
		c.Set("validated_query", req)
		c.Next()
	}
}

// getValidationMessage returns a user-friendly validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "uuid":
		return "Must be a valid UUID"
	case "oneof":
		return "Value must be one of the allowed options"
	case "numeric":
		return "Must be a number"
	case "gte":
		return "Value must be greater than or equal to " + err.Param()
	case "lte":
		return "Value must be less than or equal to " + err.Param()
	default:
		return "Invalid value"
	}
}

// GetValidatedRequest retrieves validated request from context
func GetValidatedRequest[T any](c *gin.Context) (T, bool) {
	var req T
	if val, exists := c.Get("validated_request"); exists {
		req = val.(T)
		return req, true
	}
	return req, false
}

// GetValidatedQuery retrieves validated query from context
func GetValidatedQuery[T any](c *gin.Context) (T, bool) {
	var req T
	if val, exists := c.Get("validated_query"); exists {
		req = val.(T)
		return req, true
	}
	return req, false
}
