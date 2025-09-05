package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/database"
	"github.com/your-org/lms-backend/internal/middleware"
	"github.com/your-org/lms-backend/internal/models"
)

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
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "user_exists",
			Message: "User with this email already exists",
			Code:    http.StatusConflict,
		})
		return
	}

	// Create new user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: "hashed_password_" + req.Password, // TODO: Implement proper password hashing
		Name:         req.Name,
		Role:         req.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "registration_failed",
			Message: "Failed to create user account",
			Code:    http.StatusInternalServerError,
		})
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
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// TODO: Implement proper password verification
	// For now, just check if password matches our mock format
	expectedHash := "hashed_password_" + req.Password
	if user.PasswordHash != expectedHash {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// TODO: Generate real JWT token
	token := "jwt-token-" + uuid.New().String()[:8]

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
	// TODO: Extract user ID from JWT token
	// For now, use a mock user ID
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	
	userRepo := database.GetRepoManager().User()
	user, err := userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "user_not_found",
			Message: "User profile not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Get user statistics
	enrollmentRepo := database.GetRepoManager().Enrollment()
	enrollments, _, err := enrollmentRepo.GetByUser(c.Request.Context(), userID, models.PaginationRequest{Page: 1, PageSize: 100})
	if err != nil {
		enrollments = []models.Enrollment{}
	}

	certificateRepo := database.GetRepoManager().Certificate()
	certificates, _, err := certificateRepo.GetByUser(c.Request.Context(), userID, models.PaginationRequest{Page: 1, PageSize: 100})
	if err != nil {
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

	// TODO: Extract user ID from JWT token
	// For now, use a mock user ID
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	
	userRepo := database.GetRepoManager().User()
	user, err := userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "user_not_found",
			Message: "User profile not found",
			Code:    http.StatusNotFound,
		})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to update user profile",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := models.UpdateProfileResponse{
		Message: "Profile updated successfully",
		UserID:  user.ID,
		Status:  "success",
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_instructor_id",
			Message: "Invalid instructor ID format",
			Code:    http.StatusBadRequest,
		})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: "Failed to create course",
			Code:    http.StatusInternalServerError,
		})
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid course ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	courseRepo := database.GetRepoManager().Course()
	courseDetail, err := courseRepo.GetWithDetails(c.Request.Context(), courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "course_not_found",
			Message: "Course not found",
			Code:    http.StatusNotFound,
		})
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

	lessons := []models.LessonResponse{
		{
			ID:          uuid.New(),
			CourseID:    uuid.New(),
			Title:       "Getting Started with Go",
			Content:     stringPtr("Introduction to Go syntax and basic concepts"),
			OrderNumber: 1,
			CreatedAt:   time.Now().AddDate(0, -2, 0),
		},
		{
			ID:          uuid.New(),
			CourseID:    uuid.New(),
			Title:       "Variables and Types",
			Content:     stringPtr("Understanding Go's type system"),
			OrderNumber: 2,
			CreatedAt:   time.Now().AddDate(0, -2, 0),
		},
	}

	response := models.LessonListResponse{
		Lessons: lessons,
		Pagination: models.PaginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			Total:      len(lessons),
			TotalPages: (len(lessons) + pagination.PageSize - 1) / pagination.PageSize,
		},
	}
	c.JSON(http.StatusOK, response)
}

func CreateLesson(c *gin.Context) {
	_, exists := middleware.GetValidatedRequest[models.CreateLessonRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	response := models.SuccessResponse{
		Message: "Lesson created successfully",
		Data: gin.H{
			"lesson_id": uuid.New(),
		},
	}
	c.JSON(http.StatusCreated, response)
}

func GetLesson(c *gin.Context) {
	lessonID := c.Param("id")
	response := models.LessonDetailResponse{
		LessonResponse: models.LessonResponse{
			ID:          uuid.MustParse(lessonID),
			CourseID:    uuid.New(),
			Title:       "Getting Started with Go",
			Content:     stringPtr("This lesson covers the basics of Go programming..."),
			OrderNumber: 1,
			CreatedAt:   time.Now().AddDate(0, -2, 0),
			UpdatedAt:   time.Now(),
		},
		CourseTitle: "Introduction to Go Programming",
		Duration:    30,
	}
	c.JSON(http.StatusOK, response)
}

func UpdateLesson(c *gin.Context) {
	lessonID := c.Param("id")
	_, exists := middleware.GetValidatedRequest[models.UpdateLessonRequest](c)
	if !exists {
		return // Error already handled by middleware
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
	lessonID := c.Param("id")
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

	response := models.CreateEnrollmentResponse{
		Message:      "Enrollment successful",
		EnrollmentID: uuid.New(),
		UserID:       uuid.MustParse(req.UserID),
		CourseID:     uuid.MustParse(req.CourseID),
		Status:       "enrolled",
		EnrolledAt:   time.Now(),
	}
	c.JSON(http.StatusCreated, response)
}

func ListEnrollments(c *gin.Context) {
	pagination, exists := middleware.GetValidatedQuery[models.PaginationRequest](c)
	if !exists {
		pagination = models.PaginationRequest{Page: 1, PageSize: 10}
	}

	enrollments := []models.EnrollmentDetailResponse{
		{
			EnrollmentResponse: models.EnrollmentResponse{
				UserID:     uuid.New(),
				CourseID:   uuid.New(),
				EnrolledAt: time.Now().AddDate(0, -1, 0),
			},
			CourseTitle: "Introduction to Go Programming",
			UserName:    "John Doe",
			Progress:    25.5,
			Status:      "enrolled",
		},
		{
			EnrollmentResponse: models.EnrollmentResponse{
				UserID:     uuid.New(),
				CourseID:   uuid.New(),
				EnrolledAt: time.Now().AddDate(0, -2, 0),
			},
			CourseTitle: "Advanced Web Development",
			UserName:    "Jane Smith",
			Progress:    100.0,
			Status:      "completed",
		},
	}

	response := models.EnrollmentListResponse{
		Enrollments: enrollments,
		Pagination: models.PaginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			Total:      len(enrollments),
			TotalPages: (len(enrollments) + pagination.PageSize - 1) / pagination.PageSize,
		},
	}
	c.JSON(http.StatusOK, response)
}

func GetEnrollment(c *gin.Context) {
	response := models.EnrollmentDetailResponse{
		EnrollmentResponse: models.EnrollmentResponse{
			UserID:     uuid.New(),
			CourseID:   uuid.New(),
			EnrolledAt: time.Now().AddDate(0, -1, 0),
		},
		CourseTitle: "Introduction to Go Programming",
		UserName:    "John Doe",
		Progress:    25.5,
		Status:      "enrolled",
	}
	c.JSON(http.StatusOK, response)
}

// Progress handlers
func CompleteLesson(c *gin.Context) {
	req, exists := middleware.GetValidatedRequest[models.CompleteLessonRequest](c)
	if !exists {
		return // Error already handled by middleware
	}

	response := models.CompleteLessonResponse{
		Message:        "Lesson completed successfully",
		ProgressID:     uuid.New(),
		UserID:         uuid.MustParse(req.UserID),
		LessonID:       uuid.MustParse(req.LessonID),
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