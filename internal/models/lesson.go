package models

import (
	"time"

	"github.com/google/uuid"
)

// Lesson represents a lesson in a course
type Lesson struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CourseID    uuid.UUID `json:"course_id" db:"course_id"`
	Title       string    `json:"title" db:"title"`
	Content     *string   `json:"content" db:"content"`
	OrderNumber int       `json:"order_number" db:"order_number"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateLessonRequest represents the request payload for creating a lesson
type CreateLessonRequest struct {
	CourseID    string  `json:"course_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title       string  `json:"title" validate:"required,min=3,max=200" example:"Getting Started with Go"`
	Content     *string `json:"content,omitempty" validate:"omitempty,max=5000" example:"This lesson covers the basics of Go programming..."`
	OrderNumber int     `json:"order_number,omitempty" validate:"omitempty,min=1" example:"1"`
}

// UpdateLessonRequest represents the request payload for updating a lesson
type UpdateLessonRequest struct {
	Title       string  `json:"title,omitempty" validate:"omitempty,min=3,max=200" example:"Getting Started with Go"`
	Content     *string `json:"content,omitempty" validate:"omitempty,max=5000" example:"This lesson covers the basics of Go programming..."`
	OrderNumber int     `json:"order_number,omitempty" validate:"omitempty,min=1" example:"1"`
}

// LessonResponse represents the response payload for lesson data
type LessonResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseID    uuid.UUID `json:"course_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title       string    `json:"title" example:"Getting Started with Go"`
	Content     *string   `json:"content" example:"This lesson covers the basics of Go programming..."`
	OrderNumber int       `json:"order_number" example:"1"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// ToResponse converts a Lesson to LessonResponse
func (l *Lesson) ToResponse() LessonResponse {
	return LessonResponse{
		ID:          l.ID,
		CourseID:    l.CourseID,
		Title:       l.Title,
		Content:     l.Content,
		OrderNumber: l.OrderNumber,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}
