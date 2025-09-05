package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Hidden from JSON for security
	Name         string    `json:"name" db:"name"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole constants
const (
	RoleAdmin      = "admin"
	RoleInstructor = "instructor"
	RoleStudent    = "student"
)

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"securepassword123"`
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Role     string `json:"role" validate:"oneof=admin instructor student" example:"student"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Email string `json:"email,omitempty" validate:"omitempty,email" example:"john.doe@example.com"`
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=100" example:"John Doe"`
	Role  string `json:"role,omitempty" validate:"omitempty,oneof=admin instructor student" example:"student"`
}

// UserResponse represents the response payload for user data
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Name      string    `json:"name" example:"John Doe"`
	Role      string    `json:"role" example:"student"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
