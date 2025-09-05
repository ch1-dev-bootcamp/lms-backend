package models

import (
	"time"

	"github.com/google/uuid"
)

// Progress represents a user's progress on a lesson
type Progress struct {
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	LessonID    uuid.UUID `json:"lesson_id" db:"lesson_id"`
	CompletedAt time.Time `json:"completed_at" db:"completed_at"`
}

// CreateProgressRequest represents the request payload for creating progress
type CreateProgressRequest struct {
	LessonID string `json:"lesson_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ProgressResponse represents the response payload for progress data
type ProgressResponse struct {
	UserID      uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LessonID    uuid.UUID `json:"lesson_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CompletedAt time.Time `json:"completed_at" example:"2023-01-01T00:00:00Z"`
}

// ToResponse converts a Progress to ProgressResponse
func (p *Progress) ToResponse() ProgressResponse {
	return ProgressResponse{
		UserID:      p.UserID,
		LessonID:    p.LessonID,
		CompletedAt: p.CompletedAt,
	}
}
