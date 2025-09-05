package models

import (
	"time"

	"github.com/google/uuid"
)

// Certificate represents a course completion certificate
type Certificate struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	CourseID        uuid.UUID `json:"course_id" db:"course_id"`
	IssuedAt        time.Time `json:"issued_at" db:"issued_at"`
	CertificateCode string    `json:"certificate_code" db:"certificate_code"`
}

// CreateCertificateRequest represents the request payload for creating a certificate
type CreateCertificateRequest struct {
	UserID   string `json:"user_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseID string `json:"course_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CertificateResponse represents the response payload for certificate data
type CertificateResponse struct {
	ID              uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID          uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseID        uuid.UUID `json:"course_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	IssuedAt        time.Time `json:"issued_at" example:"2023-01-01T00:00:00Z"`
	CertificateCode string    `json:"certificate_code" example:"CERT-12345-ABCD"`
}

// ToResponse converts a Certificate to CertificateResponse
func (c *Certificate) ToResponse() CertificateResponse {
	return CertificateResponse{
		ID:              c.ID,
		UserID:          c.UserID,
		CourseID:        c.CourseID,
		IssuedAt:        c.IssuedAt,
		CertificateCode: c.CertificateCode,
	}
}
