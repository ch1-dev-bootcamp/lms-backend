package models

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel contains common fields for all entities
type BaseModel struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `json:"page" form:"page" validate:"min=1" example:"1"`
	PageSize int `json:"page_size" form:"page_size" validate:"min=1,max=100" example:"10"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"validation_failed"`
	Message string `json:"message,omitempty" example:"Invalid input data"`
	Code    int    `json:"code,omitempty" example:"400"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Email is required"`
	Value   string `json:"value,omitempty" example:"invalid-email"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string           `json:"error" example:"validation_failed"`
	Details []ValidationError `json:"details"`
}

// GetOffset calculates the offset for pagination
func (p *PaginationRequest) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return (p.Page - 1) * p.PageSize
}

// CalculateTotalPages calculates total pages for pagination
func (p *PaginationResponse) CalculateTotalPages() {
	if p.PageSize > 0 {
		p.TotalPages = (p.Total + p.PageSize - 1) / p.PageSize
	}
}
