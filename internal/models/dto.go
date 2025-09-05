package models

import (
	"time"

	"github.com/google/uuid"
)

// Authentication DTOs
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"securepassword123"`
}

type LoginResponse struct {
	Message string    `json:"message" example:"Login successful"`
	Token   string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID  uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status  string    `json:"status" example:"success"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"securepassword123"`
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Role     string `json:"role" validate:"oneof=admin instructor student" example:"student"`
}

type RegisterResponse struct {
	Message string    `json:"message" example:"User registered successfully"`
	UserID  uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status  string    `json:"status" example:"success"`
}

type LogoutResponse struct {
	Message string `json:"message" example:"Logout successful"`
	Status  string `json:"status" example:"success"`
}

// Pagination DTOs are defined in common.go

// Enhanced Course DTOs
type CourseListResponse struct {
	Courses []CourseResponse `json:"courses"`
	Pagination PaginationResponse `json:"pagination"`
}

type CourseDetailResponse struct {
	CourseResponse
	InstructorName string `json:"instructor_name" example:"Jane Smith"`
	LessonCount    int    `json:"lesson_count" example:"12"`
	EnrollmentCount int   `json:"enrollment_count" example:"150"`
}

// Enhanced Lesson DTOs
type LessonListResponse struct {
	Lessons []LessonResponse `json:"lessons"`
	Pagination PaginationResponse `json:"pagination"`
}

type LessonDetailResponse struct {
	LessonResponse
	CourseTitle string `json:"course_title" example:"Introduction to Go Programming"`
	Duration    int    `json:"duration" example:"30"` // in minutes
}

// Enhanced Enrollment DTOs
type EnrollmentListResponse struct {
	Enrollments []EnrollmentDetailResponse `json:"enrollments"`
	Pagination  PaginationResponse         `json:"pagination"`
}

type EnrollmentDetailResponse struct {
	EnrollmentResponse
	CourseTitle string  `json:"course_title" example:"Introduction to Go Programming"`
	UserName    string  `json:"user_name" example:"John Doe"`
	Progress    float64 `json:"progress" example:"25.5"`
	Status      string  `json:"status" example:"enrolled"`
}

type CreateEnrollmentResponse struct {
	Message       string    `json:"message" example:"Enrollment successful"`
	EnrollmentID  uuid.UUID `json:"enrollment_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID        uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseID      uuid.UUID `json:"course_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status        string    `json:"status" example:"enrolled"`
	EnrolledAt    time.Time `json:"enrolled_at" example:"2023-01-01T00:00:00Z"`
}

// Enhanced Progress DTOs
type ProgressListResponse struct {
	Progress []ProgressDetailResponse `json:"progress"`
	Total    int                      `json:"total" example:"5"`
}

type ProgressDetailResponse struct {
	UserID           uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseID         uuid.UUID `json:"course_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CourseTitle      string    `json:"course_title" example:"Introduction to Go Programming"`
	TotalLessons     int       `json:"total_lessons" example:"12"`
	CompletedLessons int       `json:"completed_lessons" example:"3"`
	CompletionRate   float64   `json:"completion_rate" example:"25.0"`
	LastActivity     time.Time `json:"last_activity" example:"2023-01-01T00:00:00Z"`
	EnrolledAt       time.Time `json:"enrolled_at" example:"2023-01-01T00:00:00Z"`
}

type CompleteLessonRequest struct {
	UserID   string `json:"user_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	LessonID string `json:"lesson_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CompleteLessonResponse struct {
	Message        string    `json:"message" example:"Lesson completed successfully"`
	ProgressID     uuid.UUID `json:"progress_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LessonID       uuid.UUID `json:"lesson_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CompletionRate float64   `json:"completion_rate" example:"100.0"`
	CompletedAt    time.Time `json:"completed_at" example:"2023-01-01T00:00:00Z"`
	Status         string    `json:"status" example:"success"`
}

// Enhanced Certificate DTOs
type CertificateListResponse struct {
	Certificates []CertificateDetailResponse `json:"certificates"`
	Pagination   PaginationResponse          `json:"pagination"`
}

type CertificateDetailResponse struct {
	CertificateResponse
	UserName   string `json:"user_name" example:"John Doe"`
	CourseTitle string `json:"course_title" example:"Introduction to Go Programming"`
	Status     string `json:"status" example:"valid"`
}

type VerifyCertificateResponse struct {
	CertificateID string    `json:"certificate_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Valid         bool      `json:"valid" example:"true"`
	UserName      string    `json:"user_name" example:"John Doe"`
	CourseTitle   string    `json:"course_title" example:"Introduction to Go Programming"`
	IssuedAt      time.Time `json:"issued_at" example:"2023-01-01T00:00:00Z"`
	VerifiedAt    time.Time `json:"verified_at" example:"2023-01-01T00:00:00Z"`
}

// Common Response DTOs are defined in common.go

// Profile DTOs
type ProfileResponse struct {
	UserResponse
	EnrollmentCount int `json:"enrollment_count" example:"5"`
	CompletedCourses int `json:"completed_courses" example:"2"`
	CertificatesCount int `json:"certificates_count" example:"2"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=100" example:"John Doe"`
	Email string `json:"email,omitempty" validate:"omitempty,email" example:"john.doe@example.com"`
}

type UpdateProfileResponse struct {
	Message string    `json:"message" example:"Profile updated successfully"`
	UserID  uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status  string    `json:"status" example:"success"`
}
