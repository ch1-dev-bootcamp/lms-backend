package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/auth"
	"github.com/your-org/lms-backend/internal/certificate"
	"github.com/your-org/lms-backend/internal/database"
	apperrors "github.com/your-org/lms-backend/internal/errors"
	"github.com/your-org/lms-backend/internal/middleware"
	"github.com/your-org/lms-backend/internal/models"
)

// Global JWT manager - will be initialized in main
var jwtManager *auth.JWTManager

// SetJWTManager sets the global JWT manager
func SetJWTManager(manager *auth.JWTManager) {
	jwtManager = manager
}


// HealthCheck handles health check requests
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "lms-backend",
	})
}

// APIRoot handles API root requests
func APIRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "LMS API v1",
		"version": "1.0.0",
	})
}

// DatabaseHealth handles database health check requests
func DatabaseHealth(c *gin.Context) {
	if err := database.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"service":   "database",
			"error":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "database",
	})
}

// Authentication handlers
func Register(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.RegisterRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Check if user already exists
	userRepo := database.GetRepoManager().User()
	existingUser, err := userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err == nil && existingUser != nil {
		middleware.AbortWithError(c, apperrors.NewUserExistsError())
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Create new user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		Role:         req.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := userRepo.Create(c.Request.Context(), user); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.RegisterResponse{
		Message: "User registered successfully",
		UserID:  userID,
		Status:  "success",
	}
	c.JSON(http.StatusCreated, response)
}

func Login(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.LoginRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Find user by email
	userRepo := database.GetRepoManager().User()
	user, err := userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
			// Check if it's a not found error
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		if appErr.Code == apperrors.ErrorCodeNotFound {
			middleware.AbortWithError(c, apperrors.NewValidationError("Invalid email or password"))
			return
		}
	}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		middleware.AbortWithError(c, apperrors.NewValidationError("Invalid email or password"))
		return
	}

	// Generate JWT token
	token, err := jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.LoginResponse{
		Message: "Login successful",
		Token:   token,
		UserID:  user.ID,
		Status:  "success",
	}
	c.JSON(http.StatusOK, response)
}

func Logout(c *gin.Context) {
	response := models.LogoutResponse{
		Message: "Logout successful",
		Status:  "success",
	}
	c.JSON(http.StatusOK, response)
}

// User handlers
func GetProfile(c *gin.Context) {
	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}
	
	userRepo := database.GetRepoManager().User()
	user, err := userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
			// Check if it's a not found error
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		if appErr.Code == apperrors.ErrorCodeNotFound {
			middleware.AbortWithError(c, apperrors.NewUserNotFoundError())
			return
		}
	}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Get user statistics
	enrollmentRepo := database.GetRepoManager().Enrollment()
	enrollments, _, err := enrollmentRepo.GetByUser(c.Request.Context(), userID, models.PaginationRequest{Page: 1, PageSize: 100})
	if err != nil {
		// Log error but don't fail the request
		enrollments = []models.Enrollment{}
	}

	certificateRepo := database.GetRepoManager().Certificate()
	certificates, _, err := certificateRepo.GetByUser(c.Request.Context(), userID, models.PaginationRequest{Page: 1, PageSize: 100})
	if err != nil {
		// Log error but don't fail the request
		certificates = []models.Certificate{}
	}

	response := models.ProfileResponse{
		UserResponse: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		EnrollmentCount:   len(enrollments),
		CompletedCourses:  len(enrollments), // Same as enrollments for now
		CertificatesCount: len(certificates),
	}
	c.JSON(http.StatusOK, response)
}

func UpdateProfile(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.UpdateProfileRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}
	
	userRepo := database.GetRepoManager().User()
	user, err := userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewUserNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Update user fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	user.UpdatedAt = time.Now()

	if err := userRepo.Update(c.Request.Context(), user); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.UpdateProfileResponse{
		Message: "Profile updated successfully",
		UserID:  user.ID,
		Status:  "success",
	}
	c.JSON(http.StatusOK, response)
}

func DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	userRepo := database.GetRepoManager().User()
	_, err = userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewUserNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	if err := userRepo.Delete(c.Request.Context(), userID); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "User deleted successfully",
		Data: gin.H{
			"user_id": userID,
		},
	}
	c.JSON(http.StatusOK, response)
}

// Course handlers
func ListCourses(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	courseRepo := database.GetRepoManager().Course()
	// Only show published courses for public listing
	courses, paginationResp, err := courseRepo.GetByStatus(c.Request.Context(), models.CourseStatusPublished, pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to response format
	courseResponses := make([]models.CourseResponse, len(courses))
	for i, course := range courses {
		courseResponses[i] = models.CourseResponse{
			ID:           course.ID,
			Title:        course.Title,
			Description:  course.Description,
			InstructorID: course.InstructorID,
			Status:       course.Status,
			CreatedAt:    course.CreatedAt,
		}
	}

	response := models.CourseListResponse{
		Courses:    courseResponses,
		Pagination: *paginationResp,
	}
	c.JSON(http.StatusOK, response)
}

func CreateCourse(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CreateCourseRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Extract instructor ID from JWT token (authenticated user)
	instructorIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	instructorID, err := uuid.Parse(instructorIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid instructor ID format"))
		return
	}

	// Verify instructor exists and has instructor or admin role
	userRepo := database.GetRepoManager().User()
	instructor, err := userRepo.GetByID(c.Request.Context(), instructorID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewForeignKeyViolationError("Instructor not found"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check if user has instructor or admin role
	if instructor.Role != models.RoleInstructor && instructor.Role != models.RoleAdmin {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only instructors and admins can create courses"))
		return
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = models.CourseStatusDraft
	}

	courseID := uuid.New()
	course := &models.Course{
		ID:           courseID,
		Title:        req.Title,
		Description:  req.Description,
		InstructorID: instructorID,
		Status:       status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	courseRepo := database.GetRepoManager().Course()
	if err := courseRepo.Create(c.Request.Context(), course); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Course created successfully",
		Data: gin.H{
			"course_id": courseID,
		},
	}
	c.JSON(http.StatusCreated, response)
}

func GetCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	courseRepo := database.GetRepoManager().Course()
	courseDetail, err := courseRepo.GetWithDetails(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.CourseDetailResponse{
		CourseResponse: models.CourseResponse{
			ID:           courseDetail.ID,
			Title:        courseDetail.Title,
			Description:  courseDetail.Description,
			InstructorID: courseDetail.InstructorID,
			Status:       courseDetail.Status,
			CreatedAt:    courseDetail.CreatedAt,
			UpdatedAt:    courseDetail.UpdatedAt,
		},
		InstructorName:  courseDetail.InstructorName,
		LessonCount:     courseDetail.LessonCount,
		EnrollmentCount: courseDetail.EnrollmentCount,
	}
	c.JSON(http.StatusOK, response)
}

func UpdateCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	req, exists := middleware.GetValidatedRequest[models.UpdateCourseRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only instructor or admin can update
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can update this course"))
		return
	}

	// Update course fields
	if req.Title != "" {
		course.Title = req.Title
	}
	if req.Description != nil && *req.Description != "" {
		course.Description = req.Description
	}
	if req.Status != "" {
		course.Status = req.Status
	}
	course.UpdatedAt = time.Now()

	if err := courseRepo.Update(c.Request.Context(), course); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Course updated successfully",
		Data: gin.H{
			"course_id": courseID,
		},
	}
	c.JSON(http.StatusOK, response)
}

func DeleteCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only instructor or admin can delete
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can delete this course"))
		return
	}

	if err := courseRepo.Delete(c.Request.Context(), courseID); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Course deleted successfully",
		Data: gin.H{
			"course_id": courseID,
		},
	}
	c.JSON(http.StatusOK, response)
}

// InstructorCourseList handles listing courses for instructors (all their courses)
func InstructorCourseList(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	// Extract instructor ID from JWT token
	instructorIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	instructorID, err := uuid.Parse(instructorIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid instructor ID format"))
		return
	}

	courseRepo := database.GetRepoManager().Course()
	courses, paginationResp, err := courseRepo.GetByInstructor(c.Request.Context(), instructorID, pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to response format
	courseResponses := make([]models.CourseResponse, len(courses))
	for i, course := range courses {
		courseResponses[i] = models.CourseResponse{
			ID:           course.ID,
			Title:        course.Title,
			Description:  course.Description,
			InstructorID: course.InstructorID,
			Status:       course.Status,
			CreatedAt:    course.CreatedAt,
			UpdatedAt:    course.UpdatedAt,
		}
	}

	response := models.CourseListResponse{
		Courses:    courseResponses,
		Pagination: *paginationResp,
	}
	c.JSON(http.StatusOK, response)
}

// SearchCourses handles searching published courses
func SearchCourses(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	query := c.Query("q")
	if query == "" {
		middleware.AbortWithError(c, apperrors.NewValidationError("Search query is required"))
		return
	}

	courseRepo := database.GetRepoManager().Course()
	courses, paginationResp, err := courseRepo.Search(c.Request.Context(), query, pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to response format
	courseResponses := make([]models.CourseResponse, len(courses))
	for i, course := range courses {
		courseResponses[i] = models.CourseResponse{
			ID:           course.ID,
			Title:        course.Title,
			Description:  course.Description,
			InstructorID: course.InstructorID,
			Status:       course.Status,
			CreatedAt:    course.CreatedAt,
			UpdatedAt:    course.UpdatedAt,
		}
	}

	response := models.CourseListResponse{
		Courses:    courseResponses,
		Pagination: *paginationResp,
	}
	c.JSON(http.StatusOK, response)
}

// Lesson handlers
func ListLessons(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	lessonRepo := database.GetRepoManager().Lesson()
	lessons, paginationResp, err := lessonRepo.List(c.Request.Context(), pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to response format
	lessonResponses := make([]models.LessonResponse, len(lessons))
	for i, lesson := range lessons {
		lessonResponses[i] = models.LessonResponse{
			ID:          lesson.ID,
			CourseID:    lesson.CourseID,
			Title:       lesson.Title,
			Content:     lesson.Content,
			OrderNumber: lesson.OrderNumber,
			CreatedAt:   lesson.CreatedAt,
			UpdatedAt:   lesson.UpdatedAt,
		}
	}

	response := models.LessonListResponse{
		Lessons:    lessonResponses,
		Pagination: *paginationResp,
	}
	c.JSON(http.StatusOK, response)
}

// ListLessonsByCourse handles listing lessons for a specific course
func ListLessonsByCourse(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	// Verify course exists
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check if course is published or user has access
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		// Public access - only show published courses
		if course.Status != models.CourseStatusPublished {
			middleware.AbortWithError(c, apperrors.NewForbiddenError("Course is not published"))
			return
		}
	} else {
		// Authenticated user - check if they have access
		userIDStr, exists := middleware.GetUserIDFromContext(c)
		if exists {
			userID, err := uuid.Parse(userIDStr)
			if err == nil && userRole != models.RoleAdmin && course.InstructorID != userID && course.Status != models.CourseStatusPublished {
				middleware.AbortWithError(c, apperrors.NewForbiddenError("Access denied to this course"))
				return
			}
		}
	}

	lessonRepo := database.GetRepoManager().Lesson()
	lessons, paginationResp, err := lessonRepo.GetByCourse(c.Request.Context(), courseID, pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to response format
	lessonResponses := make([]models.LessonResponse, len(lessons))
	for i, lesson := range lessons {
		lessonResponses[i] = models.LessonResponse{
			ID:          lesson.ID,
			CourseID:    lesson.CourseID,
			Title:       lesson.Title,
			Content:     lesson.Content,
			OrderNumber: lesson.OrderNumber,
			CreatedAt:   lesson.CreatedAt,
			UpdatedAt:   lesson.UpdatedAt,
		}
	}

	response := models.LessonListResponse{
		Lessons:    lessonResponses,
		Pagination: *paginationResp,
	}
	c.JSON(http.StatusOK, response)
}

func CreateLesson(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CreateLessonRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Parse course ID from request
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Verify course exists and check ownership
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only course instructor or admin can create lessons
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can create lessons"))
		return
	}

	// Auto-assign order number if not provided
	orderNumber := req.OrderNumber
	if orderNumber == 0 {
		lessonRepo := database.GetRepoManager().Lesson()
		existingLessons, _, err := lessonRepo.GetByCourse(c.Request.Context(), courseID, models.PaginationRequest{Page: 1, PageSize: 1000})
		if err != nil {
			middleware.ErrorHandlerFunc(c, err)
			return
		}
		orderNumber = len(existingLessons) + 1
	}

	lessonID := uuid.New()
	lesson := &models.Lesson{
		ID:          lessonID,
		CourseID:    courseID,
		Title:       req.Title,
		Content:     req.Content,
		OrderNumber: orderNumber,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	lessonRepo := database.GetRepoManager().Lesson()
	if err := lessonRepo.Create(c.Request.Context(), lesson); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Lesson created successfully",
		Data: gin.H{
			"lesson_id": lessonID,
		},
	}
	c.JSON(http.StatusCreated, response)
}

func GetLesson(c *gin.Context) {
	lessonIDStr := c.Param("id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid lesson ID format"))
		return
	}

	lessonRepo := database.GetRepoManager().Lesson()
	lesson, err := lessonRepo.GetByID(c.Request.Context(), lessonID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewLessonNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Get course details for additional information
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), lesson.CourseID)
	if err != nil {
		// If course not found, still return lesson but without course title
		course = &models.Course{Title: "Unknown Course"}
	}

	response := models.LessonDetailResponse{
		LessonResponse: models.LessonResponse{
			ID:          lesson.ID,
			CourseID:    lesson.CourseID,
			Title:       lesson.Title,
			Content:     lesson.Content,
			OrderNumber: lesson.OrderNumber,
			CreatedAt:   lesson.CreatedAt,
			UpdatedAt:   lesson.UpdatedAt,
		},
		CourseTitle: course.Title,
		Duration:    30, // TODO: Calculate actual duration or store in database
	}
	c.JSON(http.StatusOK, response)
}

func UpdateLesson(c *gin.Context) {
	lessonIDStr := c.Param("id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid lesson ID format"))
		return
	}

	req, exists := middleware.GetValidatedRequest[models.UpdateLessonRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	lessonRepo := database.GetRepoManager().Lesson()
	lesson, err := lessonRepo.GetByID(c.Request.Context(), lessonID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewLessonNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership through course
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), lesson.CourseID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only course instructor or admin can update lessons
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can update lessons"))
		return
	}

	// Update lesson fields
	if req.Title != "" {
		lesson.Title = req.Title
	}
	if req.Content != nil {
		lesson.Content = req.Content
	}
	if req.OrderNumber != 0 {
		lesson.OrderNumber = req.OrderNumber
	}
	lesson.UpdatedAt = time.Now()

	if err := lessonRepo.Update(c.Request.Context(), lesson); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Lesson updated successfully",
		Data: gin.H{
			"lesson_id": lessonID,
		},
	}
	c.JSON(http.StatusOK, response)
}

func DeleteLesson(c *gin.Context) {
	lessonIDStr := c.Param("id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid lesson ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	lessonRepo := database.GetRepoManager().Lesson()
	lesson, err := lessonRepo.GetByID(c.Request.Context(), lessonID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewLessonNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership through course
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), lesson.CourseID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only course instructor or admin can delete lessons
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can delete lessons"))
		return
	}

	if err := lessonRepo.Delete(c.Request.Context(), lessonID); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Lesson deleted successfully",
		Data: gin.H{
			"lesson_id": lessonID,
		},
	}
	c.JSON(http.StatusOK, response)
}

// ReorderLessons handles reordering lessons within a course
func ReorderLessons(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Check course ownership
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only course instructor or admin can reorder lessons
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can reorder lessons"))
		return
	}

	// Parse reorder request
	var req struct {
		LessonOrders map[string]int `json:"lesson_orders" validate:"required" example:"{\"lesson-id-1\": 1, \"lesson-id-2\": 2}"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidRequestError("Invalid JSON format"))
		return
	}

	// Convert string keys to UUIDs
	lessonOrders := make(map[uuid.UUID]int)
	for lessonIDStr, order := range req.LessonOrders {
		lessonID, err := uuid.Parse(lessonIDStr)
		if err != nil {
			middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid lesson ID format: "+lessonIDStr))
			return
		}
		lessonOrders[lessonID] = order
	}

	// Reorder lessons
	lessonRepo := database.GetRepoManager().Lesson()
	if err := lessonRepo.ReorderLessons(c.Request.Context(), courseID, lessonOrders); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Lessons reordered successfully",
		Data: gin.H{
			"course_id": courseID,
		},
	}
	c.JSON(http.StatusOK, response)
}

// Enrollment handlers
func Enroll(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CreateEnrollmentRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Parse course ID from request
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Verify course exists and is published
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check if course is published
	if course.Status != models.CourseStatusPublished {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Course is not published and cannot be enrolled in"))
		return
	}

	// Check prerequisites
	prerequisiteRepo := database.GetRepoManager().Prerequisite()
	meetsPrerequisites, missingPrerequisites, err := prerequisiteRepo.CheckPrerequisites(c.Request.Context(), userID, courseID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	if !meetsPrerequisites {
		middleware.AbortWithError(c, apperrors.NewValidationError("Prerequisites not met. Missing required courses: " + fmt.Sprintf("%v", missingPrerequisites)))
		return
	}

	// Check if already enrolled
	enrollmentRepo := database.GetRepoManager().Enrollment()
	existingEnrollment, err := enrollmentRepo.GetByUserAndCourse(c.Request.Context(), userID, courseID)
	if err == nil && existingEnrollment != nil {
		middleware.AbortWithError(c, apperrors.NewAlreadyEnrolledError())
		return
	}

	enrollmentID := uuid.New()
	enrollment := &models.Enrollment{
		UserID:     userID,
		CourseID:   courseID,
		EnrolledAt: time.Now(),
	}

	if err := enrollmentRepo.Create(c.Request.Context(), enrollment); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.CreateEnrollmentResponse{
		Message:      "Enrollment successful",
		EnrollmentID: enrollmentID,
		UserID:       userID,
		CourseID:     courseID,
		Status:       "enrolled",
		EnrolledAt:   enrollment.EnrolledAt,
	}
	c.JSON(http.StatusCreated, response)
}

// ListUserEnrollments handles listing enrollments for the authenticated user
func ListUserEnrollments(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	enrollmentRepo := database.GetRepoManager().Enrollment()
	enrollments, paginationResp, err := enrollmentRepo.GetByUser(c.Request.Context(), userID, pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to detailed response format
	enrollmentDetails := make([]models.EnrollmentDetailResponse, len(enrollments))
	for i, enrollment := range enrollments {
		// Get course details
		courseRepo := database.GetRepoManager().Course()
		course, err := courseRepo.GetByID(c.Request.Context(), enrollment.CourseID)
		if err != nil {
			course = &models.Course{Title: "Unknown Course"}
		}

		// Get user details
		userRepo := database.GetRepoManager().User()
		user, err := userRepo.GetByID(c.Request.Context(), enrollment.UserID)
		if err != nil {
			user = &models.User{Name: "Unknown User"}
		}

		// Get actual progress from progress table
		progressRepo := database.GetRepoManager().Progress()
		courseProgress, err := progressRepo.GetCourseProgress(c.Request.Context(), enrollment.UserID, enrollment.CourseID)
		progress := 0.0
		status := "enrolled"
		if err == nil && courseProgress != nil {
			progress = courseProgress.CompletionRate
			if progress >= 100.0 {
				status = "completed"
			} else if progress > 0.0 {
				status = "in_progress"
			}
		}

		enrollmentDetails[i] = models.EnrollmentDetailResponse{
			EnrollmentResponse: models.EnrollmentResponse{
				UserID:     enrollment.UserID,
				CourseID:   enrollment.CourseID,
				EnrolledAt: enrollment.EnrolledAt,
			},
			CourseTitle: course.Title,
			UserName:    user.Name,
			Progress:    progress,
			Status:      status,
		}
	}

	response := models.EnrollmentListResponse{
		Enrollments: enrollmentDetails,
		Pagination:  *paginationResp,
	}
	c.JSON(http.StatusOK, response)
}

func ListEnrollments(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	enrollmentRepo := database.GetRepoManager().Enrollment()
	enrollments, paginationResp, err := enrollmentRepo.List(c.Request.Context(), pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to detailed response format
	enrollmentDetails := make([]models.EnrollmentDetailResponse, len(enrollments))
	for i, enrollment := range enrollments {
		// Get user details
		userRepo := database.GetRepoManager().User()
		user, err := userRepo.GetByID(c.Request.Context(), enrollment.UserID)
		if err != nil {
			user = &models.User{Name: "Unknown User"}
		}

		// Get course details
		courseRepo := database.GetRepoManager().Course()
		course, err := courseRepo.GetByID(c.Request.Context(), enrollment.CourseID)
		if err != nil {
			course = &models.Course{Title: "Unknown Course"}
		}

		// Get progress (simplified for now)
		progress := 0.0 // TODO: Calculate actual progress from progress table

		enrollmentDetails[i] = models.EnrollmentDetailResponse{
			EnrollmentResponse: models.EnrollmentResponse{
				UserID:     enrollment.UserID,
				CourseID:   enrollment.CourseID,
				EnrolledAt: enrollment.EnrolledAt,
			},
			CourseTitle: course.Title,
			UserName:    user.Name,
			Progress:    progress,
			Status:      "enrolled", // TODO: Calculate status based on progress
		}
	}

	response := models.EnrollmentListResponse{
		Enrollments: enrollmentDetails,
		Pagination:  *paginationResp,
	}
	c.JSON(http.StatusOK, response)
}

func GetEnrollment(c *gin.Context) {
	enrollmentIDStr := c.Param("id")
	enrollmentID, err := uuid.Parse(enrollmentIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid enrollment ID format"))
		return
	}

	enrollmentRepo := database.GetRepoManager().Enrollment()
	enrollment, err := enrollmentRepo.GetByID(c.Request.Context(), enrollmentID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewEnrollmentNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Get user details
	userRepo := database.GetRepoManager().User()
	user, err := userRepo.GetByID(c.Request.Context(), enrollment.UserID)
	if err != nil {
		user = &models.User{Name: "Unknown User"}
	}

	// Get course details
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), enrollment.CourseID)
	if err != nil {
		course = &models.Course{Title: "Unknown Course"}
	}

	// Get actual progress from progress table
	progressRepo := database.GetRepoManager().Progress()
	courseProgress, err := progressRepo.GetCourseProgress(c.Request.Context(), enrollment.UserID, enrollment.CourseID)
	progress := 0.0
	status := "enrolled"
	if err == nil && courseProgress != nil {
		progress = courseProgress.CompletionRate
		if progress >= 100.0 {
			status = "completed"
		} else if progress > 0.0 {
			status = "in_progress"
		}
	}

	response := models.EnrollmentDetailResponse{
		EnrollmentResponse: models.EnrollmentResponse{
			UserID:     enrollment.UserID,
			CourseID:   enrollment.CourseID,
			EnrolledAt: enrollment.EnrolledAt,
		},
		CourseTitle: course.Title,
		UserName:    user.Name,
		Progress:    progress,
		Status:      status,
	}
	c.JSON(http.StatusOK, response)
}

func UpdateEnrollment(c *gin.Context) {
	enrollmentIDStr := c.Param("id")
	enrollmentID, err := uuid.Parse(enrollmentIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid enrollment ID format"))
		return
	}

	req, exists := middleware.GetValidatedRequest[models.UpdateEnrollmentRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	enrollmentRepo := database.GetRepoManager().Enrollment()
	enrollment, err := enrollmentRepo.GetByID(c.Request.Context(), enrollmentID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewEnrollmentNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Update enrollment fields if provided
	if req.Status != "" {
		// For now, we'll just update the enrollment timestamp
		enrollment.EnrolledAt = time.Now()
	}

	if err := enrollmentRepo.Update(c.Request.Context(), enrollment); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Enrollment updated successfully",
		Data: gin.H{
			"enrollment_id": enrollmentID,
		},
	}
	c.JSON(http.StatusOK, response)
}

func DeleteEnrollment(c *gin.Context) {
	enrollmentIDStr := c.Param("id")
	enrollmentID, err := uuid.Parse(enrollmentIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid enrollment ID format"))
		return
	}

	enrollmentRepo := database.GetRepoManager().Enrollment()
	if err := enrollmentRepo.Delete(c.Request.Context(), enrollmentID); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Enrollment deleted successfully",
		Data: gin.H{
			"enrollment_id": enrollmentID,
		},
	}
	c.JSON(http.StatusOK, response)
}

// Unenroll handles unenrolling a user from a course
func Unenroll(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Check if enrollment exists
	enrollmentRepo := database.GetRepoManager().Enrollment()
	_, err = enrollmentRepo.GetByUserAndCourse(c.Request.Context(), userID, courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewEnrollmentNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Delete enrollment
	if err := enrollmentRepo.DeleteByUserAndCourse(c.Request.Context(), userID, courseID); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Successfully unenrolled from course",
		Data: gin.H{
			"user_id":   userID,
			"course_id": courseID,
		},
	}
	c.JSON(http.StatusOK, response)
}

// Progress handlers
func CompleteLesson(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CompleteLessonRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Parse lesson ID from request
	lessonID, err := uuid.Parse(req.LessonID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid lesson ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Verify lesson exists and get course info
	lessonRepo := database.GetRepoManager().Lesson()
	lesson, err := lessonRepo.GetByID(c.Request.Context(), lessonID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewLessonNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Verify user is enrolled in the course
	enrollmentRepo := database.GetRepoManager().Enrollment()
	_, err = enrollmentRepo.GetByUserAndCourse(c.Request.Context(), userID, lesson.CourseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewForbiddenError("User is not enrolled in this course"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check if already completed
	progressRepo := database.GetRepoManager().Progress()
	existingProgress, err := progressRepo.GetByUserAndLesson(c.Request.Context(), userID, lessonID)
	if err == nil && existingProgress != nil {
		// Update existing progress
		existingProgress.CompletedAt = time.Now()
		if err := progressRepo.Update(c.Request.Context(), existingProgress); err != nil {
			middleware.ErrorHandlerFunc(c, err)
			return
		}
	} else {
		// Create new progress record
		progress := &models.Progress{
			UserID:      userID,
			LessonID:    lessonID,
			CompletedAt: time.Now(),
		}

		if err := progressRepo.Create(c.Request.Context(), progress); err != nil {
			middleware.ErrorHandlerFunc(c, err)
			return
		}
	}

	// Get course progress for completion rate
	courseProgress, err := progressRepo.GetCourseProgress(c.Request.Context(), userID, lesson.CourseID)
	if err != nil {
		// If we can't get course progress, just return basic info
		courseProgress = &models.ProgressDetailResponse{
			CompletionRate: 0.0,
		}
	}

	// Check if course is completed (100% completion)
	courseCompleted := false
	if courseProgress.CompletionRate >= 100.0 {
		// Mark course as completed
		courseCompletionRepo := database.GetRepoManager().CourseCompletion()
		_, err := courseCompletionRepo.GetByUserAndCourse(c.Request.Context(), userID, lesson.CourseID)
		if err != nil {
			// Course not yet marked as completed, create completion record
			completion := &models.CourseCompletion{
				UserID:         userID,
				CourseID:       lesson.CourseID,
				CompletedAt:    time.Now(),
				CompletionRate: courseProgress.CompletionRate,
			}
			
			if err := courseCompletionRepo.Create(c.Request.Context(), completion); err != nil {
				// Log error but don't fail the request
				fmt.Printf("Failed to mark course as completed: %v\n", err)
			} else {
				courseCompleted = true
				
				// Generate certificate for completed course
				if err := generateCertificateForCompletedCourse(c.Request.Context(), userID, lesson.CourseID); err != nil {
					// Log error but don't fail the request
					fmt.Printf("Failed to generate certificate: %v\n", err)
				}
			}
		} else {
			// Course already marked as completed
			courseCompleted = true
		}
	}

	response := models.CompleteLessonResponse{
		Message:         "Lesson completed successfully",
		ProgressID:      lessonID, // Using lesson ID as progress identifier
		UserID:          userID,
		LessonID:        lessonID,
		CompletionRate:  courseProgress.CompletionRate,
		CompletedAt:     time.Now(),
		Status:          "success",
		CourseCompleted: courseCompleted,
	}
	c.JSON(http.StatusOK, response)
}

// GetUserProgress handles getting progress for the authenticated user
func GetUserProgress(c *gin.Context) {
	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	progressRepo := database.GetRepoManager().Progress()
	progress, err := progressRepo.GetUserProgress(c.Request.Context(), userID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.ProgressListResponse{
		Progress: progress,
		Total:    len(progress),
	}
	c.JSON(http.StatusOK, response)
}

// GetCourseProgress handles getting progress for a specific course
func GetCourseProgress(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Verify user is enrolled in the course
	enrollmentRepo := database.GetRepoManager().Enrollment()
	_, err = enrollmentRepo.GetByUserAndCourse(c.Request.Context(), userID, courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewForbiddenError("User is not enrolled in this course"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	progressRepo := database.GetRepoManager().Progress()
	progress, err := progressRepo.GetCourseProgress(c.Request.Context(), userID, courseID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	c.JSON(http.StatusOK, progress)
}

// generateCertificateForCompletedCourse generates a certificate for a completed course
func generateCertificateForCompletedCourse(ctx context.Context, userID, courseID uuid.UUID) error {
	// Check if certificate already exists
	certificateRepo := database.GetRepoManager().Certificate()
	existingCert, err := certificateRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err == nil && existingCert != nil {
		// Certificate already exists, no need to create another
		return nil
	}

	// Generate unique certificate code
	codeGenerator := certificate.NewCertificateCodeGenerator()
	certificateCode, err := codeGenerator.GenerateCodeWithUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to generate certificate code: %w", err)
	}

	// Create certificate
	cert := &models.Certificate{
		ID:              uuid.New(),
		UserID:          userID,
		CourseID:        courseID,
		IssuedAt:        time.Now(),
		CertificateCode: certificateCode,
	}

	// Save certificate to database
	if err := certificateRepo.Create(ctx, cert); err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	return nil
}

// Course completion handlers
func ListUserCompletions(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	courseCompletionRepo := database.GetRepoManager().CourseCompletion()
	completions, paginationResp, err := courseCompletionRepo.GetUserCompletionsWithDetails(c.Request.Context(), userID, pagination)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.CourseCompletionListResponse{
		Completions: completions,
		Total:       paginationResp.Total,
	}
	c.JSON(http.StatusOK, response)
}

func GetCourseCompletion(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	courseCompletionRepo := database.GetRepoManager().CourseCompletion()
	completion, err := courseCompletionRepo.GetByUserAndCourse(c.Request.Context(), userID, courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewValidationError("Course not completed"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Get course details
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		course = &models.Course{Title: "Unknown Course"}
	}

	// Get progress details
	progressRepo := database.GetRepoManager().Progress()
	courseProgress, err := progressRepo.GetCourseProgress(c.Request.Context(), userID, courseID)
	if err != nil {
		courseProgress = &models.ProgressDetailResponse{
			TotalLessons:     0,
			CompletedLessons: 0,
		}
	}

	response := models.CourseCompletionResponse{
		UserID:           completion.UserID,
		CourseID:         completion.CourseID,
		CourseTitle:      course.Title,
		CompletedAt:      completion.CompletedAt,
		CompletionRate:   completion.CompletionRate,
		TotalLessons:     courseProgress.TotalLessons,
		CompletedLessons: courseProgress.CompletedLessons,
	}
	c.JSON(http.StatusOK, response)
}

// CreateCertificate handles manual certificate creation (admin only)
func CreateCertificate(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CreateCertificateRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Parse user and course IDs
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Verify user exists
	userRepo := database.GetRepoManager().User()
	_, err = userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewValidationError("User not found"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Verify course exists
	courseRepo := database.GetRepoManager().Course()
	_, err = courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check if certificate already exists
	certificateRepo := database.GetRepoManager().Certificate()
	_, err = certificateRepo.GetByUserAndCourse(c.Request.Context(), userID, courseID)
	if err == nil {
		middleware.AbortWithError(c, apperrors.NewValidationError("Certificate already exists for this user and course"))
		return
	}

	// Generate unique certificate code
	codeGenerator := certificate.NewCertificateCodeGenerator()
	certificateCode, err := codeGenerator.GenerateCodeWithUserID(userID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Create certificate
	cert := &models.Certificate{
		ID:              uuid.New(),
		UserID:          userID,
		CourseID:        courseID,
		IssuedAt:        time.Now(),
		CertificateCode: certificateCode,
	}

	if err := certificateRepo.Create(c.Request.Context(), cert); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Certificate created successfully",
		Data:    cert.ToResponse(),
	}
	c.JSON(http.StatusCreated, response)
}

// Prerequisite handlers
func CreatePrerequisite(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CreatePrerequisiteRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Parse course IDs
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	requiredCourseID, err := uuid.Parse(req.RequiredCourseID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid required course ID format"))
		return
	}

	// Prevent self-referencing prerequisites
	if courseID == requiredCourseID {
		middleware.AbortWithError(c, apperrors.NewValidationError("Course cannot be a prerequisite for itself"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Verify course exists and check ownership
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only course instructor or admin can add prerequisites
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can add prerequisites"))
		return
	}

	// Verify required course exists
	_, err = courseRepo.GetByID(c.Request.Context(), requiredCourseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewValidationError("Required course not found"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Create prerequisite
	prerequisite := &models.Prerequisite{
		CourseID:         courseID,
		RequiredCourseID: requiredCourseID,
	}

	prerequisiteRepo := database.GetRepoManager().Prerequisite()
	if err := prerequisiteRepo.Create(c.Request.Context(), prerequisite); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Prerequisite added successfully",
		Data: gin.H{
			"course_id":          courseID,
			"required_course_id": requiredCourseID,
		},
	}
	c.JSON(http.StatusCreated, response)
}

func ListPrerequisites(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	prerequisiteRepo := database.GetRepoManager().Prerequisite()
	prerequisites, err := prerequisiteRepo.GetPrerequisiteCourses(c.Request.Context(), courseID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Convert to response format
	prerequisiteResponses := make([]models.PrerequisiteResponse, len(prerequisites))
	for i, course := range prerequisites {
		prerequisiteResponses[i] = models.PrerequisiteResponse{
			CourseID:         courseID,
			RequiredCourseID: course.ID,
		}
	}

	response := models.PrerequisiteListResponse{
		Prerequisites: prerequisiteResponses,
		Total:         len(prerequisiteResponses),
	}
	c.JSON(http.StatusOK, response)
}

func DeletePrerequisite(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	requiredCourseIDStr := c.Param("required_course_id")
	requiredCourseID, err := uuid.Parse(requiredCourseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid required course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	// Verify course exists and check ownership
	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewCourseNotFoundError())
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Check ownership - only course instructor or admin can remove prerequisites
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
		return
	}

	if userRole != models.RoleAdmin && course.InstructorID != userID {
		middleware.AbortWithError(c, apperrors.NewForbiddenError("Only the course instructor or admin can remove prerequisites"))
		return
	}

	// Delete prerequisite
	prerequisiteRepo := database.GetRepoManager().Prerequisite()
	if err := prerequisiteRepo.DeleteByCourseAndPrerequisite(c.Request.Context(), courseID, requiredCourseID); err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.SuccessResponse{
		Message: "Prerequisite removed successfully",
		Data: gin.H{
			"course_id":          courseID,
			"required_course_id": requiredCourseID,
		},
	}
	c.JSON(http.StatusOK, response)
}

// CheckPrerequisites handles checking if user meets prerequisites for a course
func CheckPrerequisites(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid course ID format"))
		return
	}

	// Extract user ID from JWT token
	userIDStr, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		middleware.AbortWithError(c, apperrors.NewUnauthorizedError("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	prerequisiteRepo := database.GetRepoManager().Prerequisite()
	meetsPrerequisites, missingPrerequisites, err := prerequisiteRepo.CheckPrerequisites(c.Request.Context(), userID, courseID)
	if err != nil {
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	response := models.PrerequisiteCheckResponse{
		MeetsPrerequisites:    meetsPrerequisites,
		MissingPrerequisites:  missingPrerequisites,
		Message:               "Prerequisites checked successfully",
	}
	c.JSON(http.StatusOK, response)
}

// Certificate handlers
func ListCertificates(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	certificates := []models.CertificateDetailResponse{
		{
			CertificateResponse: models.CertificateResponse{
				ID:              uuid.New(),
				UserID:          uuid.New(),
				CourseID:        uuid.New(),
				IssuedAt:        time.Now().AddDate(0, -1, 0),
				CertificateCode: "CERT-" + uuid.New().String()[:8],
			},
			UserName:    "John Doe",
			CourseTitle: "Introduction to Go Programming",
			Status:      "valid",
		},
		{
			CertificateResponse: models.CertificateResponse{
				ID:              uuid.New(),
				UserID:          uuid.New(),
				CourseID:        uuid.New(),
				IssuedAt:        time.Now().AddDate(0, -2, 0),
				CertificateCode: "CERT-" + uuid.New().String()[:8],
			},
			UserName:    "Jane Smith",
			CourseTitle: "Advanced Web Development",
			Status:      "valid",
		},
	}

	response := models.CertificateListResponse{
		Certificates: certificates,
		Pagination: models.PaginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			Total:      len(certificates),
			TotalPages: (len(certificates) + pagination.PageSize - 1) / pagination.PageSize,
		},
	}
	c.JSON(http.StatusOK, response)
}

func GetCertificate(c *gin.Context) {
	certificateIDStr := c.Param("id")
	certificateID, err := uuid.Parse(certificateIDStr)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid certificate ID format"))
		return
	}

	certificateRepo := database.GetRepoManager().Certificate()
	certificate, err := certificateRepo.GetWithDetails(c.Request.Context(), certificateID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewValidationError("Certificate not found"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	c.JSON(http.StatusOK, certificate)
}

func VerifyCertificate(c *gin.Context) {
	certificateCode := c.Param("id")
	
	certificateRepo := database.GetRepoManager().Certificate()
	certificate, err := certificateRepo.GetByCode(c.Request.Context(), certificateCode)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				response := models.VerifyCertificateResponse{
					CertificateID: certificateCode,
					Valid:         false,
					Message:       "Certificate not found or invalid",
				}
				c.JSON(http.StatusOK, response)
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Get user and course details
	userRepo := database.GetRepoManager().User()
	user, err := userRepo.GetByID(c.Request.Context(), certificate.UserID)
	if err != nil {
		user = &models.User{Name: "Unknown User"}
	}

	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), certificate.CourseID)
	if err != nil {
		course = &models.Course{Title: "Unknown Course"}
	}

	response := models.VerifyCertificateResponse{
		CertificateID: certificateCode,
		Valid:         true,
		UserName:      user.Name,
		CourseTitle:   course.Title,
		IssuedAt:      certificate.IssuedAt,
		VerifiedAt:    time.Now(),
		Message:       "Certificate is valid",
	}
	c.JSON(http.StatusOK, response)
}