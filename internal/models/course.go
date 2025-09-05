package models

import (
	"time"

	"github.com/google/uuid"
)

// Course represents a course in the system
type Course struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	Description  *string   `json:"description" db:"description"`
	InstructorID uuid.UUID `json:"instructor_id" db:"instructor_id"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CourseStatus constants
const (
	CourseStatusDraft     = "draft"
	CourseStatusPublished = "published"
	CourseStatusArchived  = "archived"
)

// CreateCourseRequest represents the request payload for creating a course
type CreateCourseRequest struct {
	Title        string  `json:"title" validate:"required,min=3,max=200" example:"Introduction to Go Programming"`
	Description  *string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Learn the basics of Go programming language"`
	InstructorID string  `json:"instructor_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       string  `json:"status,omitempty" validate:"omitempty,oneof=draft published archived" example:"published"`
}

// UpdateCourseRequest represents the request payload for updating a course
type UpdateCourseRequest struct {
	Title       string  `json:"title,omitempty" validate:"omitempty,min=3,max=200" example:"Introduction to Go Programming"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Learn the basics of Go programming language"`
	Status      string  `json:"status,omitempty" validate:"omitempty,oneof=draft published archived" example:"published"`
}

// CourseResponse represents the response payload for course data
type CourseResponse struct {
	ID           uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title        string    `json:"title" example:"Introduction to Go Programming"`
	Description  *string   `json:"description" example:"Learn the basics of Go programming language"`
	InstructorID uuid.UUID `json:"instructor_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       string    `json:"status" example:"published"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// ToResponse converts a Course to CourseResponse
func (c *Course) ToResponse() CourseResponse {
	return CourseResponse{
		ID:           c.ID,
		Title:        c.Title,
		Description:  c.Description,
		InstructorID: c.InstructorID,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
