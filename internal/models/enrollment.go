package models

import (
	"time"

	"github.com/google/uuid"
)

// Enrollment represents a user's enrollment in a course
type Enrollment struct {
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	CourseID   uuid.UUID `json:"course_id" db:"course_id"`
	EnrolledAt time.Time `json:"enrolled_at" db:"enrolled_at"`
}

// CreateEnrollmentRequest represents the request payload for creating an enrollment
type CreateEnrollmentRequest struct {
	UserID   string `json:"user_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseID string `json:"course_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateEnrollmentRequest represents the request payload for updating an enrollment
type UpdateEnrollmentRequest struct {
	Status string `json:"status,omitempty" example:"completed"`
}

// EnrollmentResponse represents the response payload for enrollment data
type EnrollmentResponse struct {
	UserID     uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseID   uuid.UUID `json:"course_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	EnrolledAt time.Time `json:"enrolled_at" example:"2023-01-01T00:00:00Z"`
}

// ToResponse converts an Enrollment to EnrollmentResponse
func (e *Enrollment) ToResponse() EnrollmentResponse {
	return EnrollmentResponse{
		UserID:     e.UserID,
		CourseID:   e.CourseID,
		EnrolledAt: e.EnrolledAt,
	}
}
