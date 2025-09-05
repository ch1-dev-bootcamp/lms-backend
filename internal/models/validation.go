package models

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator instance
var validate *validator.Validate

// InitValidator initializes the validator instance
func InitValidator() {
	validate = validator.New()
}

// ValidateStruct validates a struct using the validator
func ValidateStruct(s interface{}) error {
	if validate == nil {
		InitValidator()
	}
	return validate.Struct(s)
}

// GetValidationErrors returns detailed validation errors
func GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError
	
	if err == nil {
		return validationErrors
	}

	// Check if it's a validation error
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErr {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fieldErr.Field(),
				Message: getValidationMessage(fieldErr),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
			})
		}
	} else {
		// Generic error
		validationErrors = append(validationErrors, ValidationError{
			Field:   "general",
			Message: err.Error(),
		})
	}

	return validationErrors
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), strings.Replace(fe.Param(), " ", ", ", -1))
	case "omitempty":
		return fmt.Sprintf("%s is invalid", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// ValidateUserRequest validates a user request
func ValidateUserRequest(req interface{}) error {
	return ValidateStruct(req)
}

// ValidateCourseRequest validates a course request
func ValidateCourseRequest(req interface{}) error {
	return ValidateStruct(req)
}

// ValidateLessonRequest validates a lesson request
func ValidateLessonRequest(req interface{}) error {
	return ValidateStruct(req)
}

// ValidateEnrollmentRequest validates an enrollment request
func ValidateEnrollmentRequest(req interface{}) error {
	return ValidateStruct(req)
}

// ValidateProgressRequest validates a progress request
func ValidateProgressRequest(req interface{}) error {
	return ValidateStruct(req)
}

// ValidatePrerequisiteRequest validates a prerequisite request
func ValidatePrerequisiteRequest(req interface{}) error {
	return ValidateStruct(req)
}

// ValidateCertificateRequest validates a certificate request
func ValidateCertificateRequest(req interface{}) error {
	return ValidateStruct(req)
}
