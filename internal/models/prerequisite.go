package models

import (
	"github.com/google/uuid"
)

// Prerequisite represents a course prerequisite relationship
type Prerequisite struct {
	CourseID         uuid.UUID `json:"course_id" db:"course_id"`
	RequiredCourseID uuid.UUID `json:"required_course_id" db:"required_course_id"`
}

// CreatePrerequisiteRequest represents the request payload for creating a prerequisite
type CreatePrerequisiteRequest struct {
	CourseID         string `json:"course_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	RequiredCourseID string `json:"required_course_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// PrerequisiteResponse represents the response payload for prerequisite data
type PrerequisiteResponse struct {
	CourseID         uuid.UUID `json:"course_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RequiredCourseID uuid.UUID `json:"required_course_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ToResponse converts a Prerequisite to PrerequisiteResponse
func (p *Prerequisite) ToResponse() PrerequisiteResponse {
	return PrerequisiteResponse{
		CourseID:         p.CourseID,
		RequiredCourseID: p.RequiredCourseID,
	}
}
