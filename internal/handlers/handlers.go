package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/auth"
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

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
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
	courses, paginationResp, err := courseRepo.List(c.Request.Context(), pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_failed",
			Message: "Failed to retrieve courses",
			Code:    http.StatusInternalServerError,
		})
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

	// Parse instructor ID from request
	instructorID, err := uuid.Parse(req.InstructorID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid instructor ID format"))
		return
	}

	// Verify instructor exists
	userRepo := database.GetRepoManager().User()
	_, err = userRepo.GetByID(c.Request.Context(), instructorID)
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

	courseID := uuid.New()
	course := &models.Course{
		ID:           courseID,
		Title:        req.Title,
		Description:  req.Description,
		InstructorID: instructorID,
		Status:       req.Status,
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid course ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	req, exists := middleware.GetValidatedRequest[models.UpdateCourseRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	courseRepo := database.GetRepoManager().Course()
	course, err := courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "course_not_found",
			Message: "Course not found",
			Code:    http.StatusNotFound,
		})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to update course",
			Code:    http.StatusInternalServerError,
		})
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid course ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	courseRepo := database.GetRepoManager().Course()
	if err := courseRepo.Delete(c.Request.Context(), courseID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete course",
			Code:    http.StatusInternalServerError,
		})
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

	// Verify course exists
	courseRepo := database.GetRepoManager().Course()
	_, err = courseRepo.GetByID(c.Request.Context(), courseID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewForeignKeyViolationError("Course not found"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	lessonID := uuid.New()
	lesson := &models.Lesson{
		ID:          lessonID,
		CourseID:    courseID,
		Title:       req.Title,
		Content:     req.Content,
		OrderNumber: req.OrderNumber,
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

	lessonRepo := database.GetRepoManager().Lesson()
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

// Enrollment handlers
func Enroll(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CreateEnrollmentRequest](c)
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
				middleware.AbortWithError(c, apperrors.NewForeignKeyViolationError("User not found"))
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
				middleware.AbortWithError(c, apperrors.NewForeignKeyViolationError("Course not found"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
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

	// Get progress (simplified for now)
	progress := 0.0 // TODO: Calculate actual progress from progress table

	response := models.EnrollmentDetailResponse{
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

// Progress handlers
func CompleteLesson(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CompleteLessonRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	// Parse user and lesson IDs
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid user ID format"))
		return
	}

	lessonID, err := uuid.Parse(req.LessonID)
	if err != nil {
		middleware.AbortWithError(c, apperrors.NewInvalidFormatError("Invalid lesson ID format"))
		return
	}

	// Verify user exists
	userRepo := database.GetRepoManager().User()
	_, err = userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewForeignKeyViolationError("User not found"))
				return
			}
		}
		middleware.ErrorHandlerFunc(c, err)
		return
	}

	// Verify lesson exists
	lessonRepo := database.GetRepoManager().Lesson()
	_, err = lessonRepo.GetByID(c.Request.Context(), lessonID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrorCodeNotFound {
				middleware.AbortWithError(c, apperrors.NewForeignKeyViolationError("Lesson not found"))
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

	response := models.CompleteLessonResponse{
		Message:        "Lesson completed successfully",
		ProgressID:     uuid.New(), // TODO: Return actual progress ID
		UserID:         userID,
		LessonID:       lessonID,
		CompletionRate: 100.0,
		CompletedAt:    time.Now(),
		Status:         "success",
	}
	c.JSON(http.StatusOK, response)
}

func GetProgress(c *gin.Context) {
	response := models.ProgressDetailResponse{
		UserID:           uuid.New(),
		CourseID:         uuid.New(),
		CourseTitle:      "Introduction to Go Programming",
		TotalLessons:     12,
		CompletedLessons: 3,
		CompletionRate:   25.0,
		LastActivity:     time.Now().AddDate(0, 0, -1),
		EnrolledAt:       time.Now().AddDate(0, -1, 0),
	}
	c.JSON(http.StatusOK, response)
}

func GetUserProgress(c *gin.Context) {
	userID := c.Param("user_id")
	progress := []models.ProgressDetailResponse{
		{
			UserID:           uuid.MustParse(userID),
			CourseID:         uuid.New(),
			CourseTitle:      "Introduction to Go Programming",
			TotalLessons:     12,
			CompletedLessons: 3,
			CompletionRate:   25.0,
			LastActivity:     time.Now().AddDate(0, 0, -1),
			EnrolledAt:       time.Now().AddDate(0, -1, 0),
		},
		{
			UserID:           uuid.MustParse(userID),
			CourseID:         uuid.New(),
			CourseTitle:      "Advanced Web Development",
			TotalLessons:     15,
			CompletedLessons: 15,
			CompletionRate:   100.0,
			LastActivity:     time.Now().AddDate(0, 0, -5),
			EnrolledAt:       time.Now().AddDate(0, -2, 0),
		},
	}

	response := models.ProgressListResponse{
		Progress: progress,
		Total:    len(progress),
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
	certificateID := c.Param("id")
	response := models.CertificateDetailResponse{
		CertificateResponse: models.CertificateResponse{
			ID:              uuid.MustParse(certificateID),
			UserID:          uuid.New(),
			CourseID:        uuid.New(),
			IssuedAt:        time.Now().AddDate(0, -1, 0),
			CertificateCode: "CERT-" + uuid.New().String()[:8],
		},
		UserName:    "John Doe",
		CourseTitle: "Introduction to Go Programming",
		Status:      "valid",
	}
	c.JSON(http.StatusOK, response)
}

func VerifyCertificate(c *gin.Context) {
	certificateID := c.Param("id")
	response := models.VerifyCertificateResponse{
		CertificateID: certificateID,
		Valid:         true,
		UserName:      "John Doe",
		CourseTitle:   "Introduction to Go Programming",
		IssuedAt:      time.Now().AddDate(0, -1, 0),
		VerifiedAt:    time.Now(),
	}
	c.JSON(http.StatusOK, response)
}