package errors

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Validation errors
	ErrorCodeValidationFailed ErrorCode = "validation_failed"
	ErrorCodeInvalidRequest   ErrorCode = "invalid_request"
	ErrorCodeInvalidFormat    ErrorCode = "invalid_format"

	// Not found errors
	ErrorCodeNotFound         ErrorCode = "not_found"
	ErrorCodeUserNotFound     ErrorCode = "user_not_found"
	ErrorCodeCourseNotFound   ErrorCode = "course_not_found"
	ErrorCodeLessonNotFound   ErrorCode = "lesson_not_found"
	ErrorCodeEnrollmentNotFound ErrorCode = "enrollment_not_found"

	// Conflict errors
	ErrorCodeConflict         ErrorCode = "conflict"
	ErrorCodeUserExists       ErrorCode = "user_exists"
	ErrorCodeAlreadyEnrolled  ErrorCode = "already_enrolled"

	// Database errors
	ErrorCodeDatabaseError    ErrorCode = "database_error"
	ErrorCodeDuplicateEntry   ErrorCode = "duplicate_entry"
	ErrorCodeForeignKeyViolation ErrorCode = "foreign_key_violation"
	ErrorCodeConstraintViolation ErrorCode = "constraint_violation"

	// Business logic errors
	ErrorCodeInvalidOperation ErrorCode = "invalid_operation"
	ErrorCodeInsufficientPermissions ErrorCode = "insufficient_permissions"
	ErrorCodePrerequisitesNotMet ErrorCode = "prerequisites_not_met"

	// Server errors
	ErrorCodeInternalError    ErrorCode = "internal_error"
	ErrorCodeServiceUnavailable ErrorCode = "service_unavailable"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
	}
}

// NewWithDetails creates a new AppError with details
func NewWithDetails(code ErrorCode, message, details string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
		Cause:      err,
	}
}

// WrapWithDetails wraps an existing error with additional context and details
func WrapWithDetails(err error, code ErrorCode, message, details string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
		Cause:      err,
	}
}

// SetRequestID sets the request ID for tracing
func (e *AppError) SetRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// Predefined error constructors for common errors

// Validation errors
func NewValidationError(message string) *AppError {
	return New(ErrorCodeValidationFailed, message, http.StatusBadRequest)
}

func NewValidationErrorWithDetails(message, details string) *AppError {
	return NewWithDetails(ErrorCodeValidationFailed, message, details, http.StatusBadRequest)
}

func NewInvalidRequestError(message string) *AppError {
	return New(ErrorCodeInvalidRequest, message, http.StatusBadRequest)
}

func NewInvalidFormatError(message string) *AppError {
	return New(ErrorCodeInvalidFormat, message, http.StatusBadRequest)
}

// Not found errors
func NewNotFoundError(message string) *AppError {
	return New(ErrorCodeNotFound, message, http.StatusNotFound)
}

func NewUserNotFoundError() *AppError {
	return New(ErrorCodeUserNotFound, "User not found", http.StatusNotFound)
}

func NewCourseNotFoundError() *AppError {
	return New(ErrorCodeCourseNotFound, "Course not found", http.StatusNotFound)
}

func NewLessonNotFoundError() *AppError {
	return New(ErrorCodeLessonNotFound, "Lesson not found", http.StatusNotFound)
}

func NewEnrollmentNotFoundError() *AppError {
	return New(ErrorCodeEnrollmentNotFound, "Enrollment not found", http.StatusNotFound)
}

// Conflict errors
func NewConflictError(message string) *AppError {
	return New(ErrorCodeConflict, message, http.StatusConflict)
}

func NewUserExistsError() *AppError {
	return New(ErrorCodeUserExists, "User with this email already exists", http.StatusConflict)
}

func NewAlreadyEnrolledError() *AppError {
	return New(ErrorCodeAlreadyEnrolled, "User is already enrolled in this course", http.StatusConflict)
}

// Database errors
func NewDatabaseError(err error, operation string) *AppError {
	return Wrap(err, ErrorCodeDatabaseError, fmt.Sprintf("Database error during %s", operation), http.StatusInternalServerError)
}

func NewDuplicateEntryError(details string) *AppError {
	return NewWithDetails(ErrorCodeDuplicateEntry, "Duplicate entry", details, http.StatusConflict)
}

func NewForeignKeyViolationError(details string) *AppError {
	return NewWithDetails(ErrorCodeForeignKeyViolation, "Referenced record not found", details, http.StatusBadRequest)
}

func NewConstraintViolationError(details string) *AppError {
	return NewWithDetails(ErrorCodeConstraintViolation, "Constraint violation", details, http.StatusBadRequest)
}

// Business logic errors
func NewInvalidOperationError(message string) *AppError {
	return New(ErrorCodeInvalidOperation, message, http.StatusUnprocessableEntity)
}

func NewInsufficientPermissionsError() *AppError {
	return New(ErrorCodeInsufficientPermissions, "Insufficient permissions", http.StatusForbidden)
}

func NewPrerequisitesNotMetError(prerequisites []string) *AppError {
	details := fmt.Sprintf("Missing prerequisites: %v", prerequisites)
	return NewWithDetails(ErrorCodePrerequisitesNotMet, "Course prerequisites not met", details, http.StatusUnprocessableEntity)
}

// Server errors
func NewInternalError(err error, operation string) *AppError {
	return Wrap(err, ErrorCodeInternalError, fmt.Sprintf("Internal error during %s", operation), http.StatusInternalServerError)
}

func NewServiceUnavailableError(service string) *AppError {
	return New(ErrorCodeServiceUnavailable, fmt.Sprintf("%s service is unavailable", service), http.StatusServiceUnavailable)
}

// ErrorResponse represents the error response sent to clients
type ErrorResponse struct {
	Error     ErrorCode `json:"error"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ToResponse converts AppError to ErrorResponse
func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Error:     e.Code,
		Message:   e.Message,
		Details:   e.Details,
		RequestID: e.RequestID,
		Timestamp: e.Timestamp,
	}
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}
