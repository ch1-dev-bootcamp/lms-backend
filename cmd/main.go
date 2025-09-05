package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lms-backend/internal/auth"
	"github.com/your-org/lms-backend/internal/database"
	"github.com/your-org/lms-backend/internal/handlers"
	"github.com/your-org/lms-backend/internal/middleware"
	"github.com/your-org/lms-backend/internal/models"
	"github.com/your-org/lms-backend/pkg/config"
	"github.com/your-org/lms-backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Initialize(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	logger.Info("Starting LMS Backend application")

	// Initialize database connection
	logger.Info("Initializing database connection...")
	dbConfig := database.NewConnectionConfig(cfg)
	if err := database.Connect(dbConfig); err != nil {
		logger.Fatal("Failed to connect to database", logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Data)
	}
	defer database.Close()

	logger.Info("Database connected successfully!")

	// Initialize JWT manager
	tokenDuration, err := time.ParseDuration(cfg.JWT.TokenDuration)
	if err != nil {
		logger.Fatal("Invalid JWT token duration", logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Data)
	}
	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey, tokenDuration)
	handlers.SetJWTManager(jwtManager)

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.Logging())
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Recovery())

	// Health check endpoints
	r.GET("/health", handlers.HealthCheck)
	r.GET("/health/database", handlers.DatabaseHealth)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", handlers.APIRoot)
		
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", middleware.ValidateRequest[models.RegisterRequest](), handlers.Register)
			auth.POST("/login", middleware.ValidateRequest[models.LoginRequest](), handlers.Login)
			auth.POST("/logout", handlers.Logout)
		}
		
		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("/profile", handlers.GetProfile)
			users.PUT("/profile", middleware.ValidateRequest[models.UpdateProfileRequest](), handlers.UpdateProfile)
			users.DELETE("/:id", handlers.DeleteUser)
		}
		
		// Course routes (public)
		courses := v1.Group("/courses")
		{
			courses.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListCourses)
			courses.GET("/search", middleware.ValidateQuery[models.PaginationRequest](), handlers.SearchCourses)
			courses.GET("/:id", handlers.GetCourse)
		}
		
		// Course management routes (protected - instructors and admins)
		courseMgmt := v1.Group("/courses")
		courseMgmt.Use(middleware.AuthMiddleware(jwtManager))
		courseMgmt.Use(middleware.RequireRole(models.RoleInstructor))
		{
			courseMgmt.POST("/", middleware.ValidateRequest[models.CreateCourseRequest](), handlers.CreateCourse)
			courseMgmt.PUT("/:id", middleware.ValidateRequest[models.UpdateCourseRequest](), handlers.UpdateCourse)
			courseMgmt.DELETE("/:id", handlers.DeleteCourse)
			courseMgmt.GET("/my-courses", middleware.ValidateQuery[models.PaginationRequest](), handlers.InstructorCourseList)
		}
		
		// Lesson routes (public)
		lessons := v1.Group("/lessons")
		{
			lessons.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListLessons)
			lessons.GET("/:id", handlers.GetLesson)
		}
		
		// Course lessons routes (public for published courses)
		courseLessons := v1.Group("/courses/:course_id/lessons")
		{
			courseLessons.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListLessonsByCourse)
		}
		
		// Lesson management routes (protected - instructors and admins)
		lessonMgmt := v1.Group("/lessons")
		lessonMgmt.Use(middleware.AuthMiddleware(jwtManager))
		lessonMgmt.Use(middleware.RequireRole(models.RoleInstructor))
		{
			lessonMgmt.POST("/", middleware.ValidateRequest[models.CreateLessonRequest](), handlers.CreateLesson)
			lessonMgmt.PUT("/:id", middleware.ValidateRequest[models.UpdateLessonRequest](), handlers.UpdateLesson)
			lessonMgmt.DELETE("/:id", handlers.DeleteLesson)
		}
		
		// Course lesson management routes (protected - instructors and admins)
		courseLessonMgmt := v1.Group("/courses/:course_id/lessons")
		courseLessonMgmt.Use(middleware.AuthMiddleware(jwtManager))
		courseLessonMgmt.Use(middleware.RequireRole(models.RoleInstructor))
		{
			courseLessonMgmt.POST("/", middleware.ValidateRequest[models.CreateLessonRequest](), handlers.CreateLesson)
			courseLessonMgmt.PUT("/:id", middleware.ValidateRequest[models.UpdateLessonRequest](), handlers.UpdateLesson)
			courseLessonMgmt.DELETE("/:id", handlers.DeleteLesson)
			courseLessonMgmt.PUT("/reorder", handlers.ReorderLessons)
		}
		
		// Enrollment routes (public for enrollment, protected for management)
		enrollments := v1.Group("/enrollments")
		{
			enrollments.POST("/", middleware.AuthMiddleware(jwtManager), middleware.ValidateRequest[models.CreateEnrollmentRequest](), handlers.Enroll)
			enrollments.GET("/my-enrollments", middleware.AuthMiddleware(jwtManager), middleware.ValidateQuery[models.PaginationRequest](), handlers.ListUserEnrollments)
			enrollments.DELETE("/courses/:course_id", middleware.AuthMiddleware(jwtManager), handlers.Unenroll)
		}
		
		// Admin enrollment management routes
		enrollmentMgmt := v1.Group("/enrollments")
		enrollmentMgmt.Use(middleware.AuthMiddleware(jwtManager))
		enrollmentMgmt.Use(middleware.RequireRole(models.RoleAdmin))
		{
			enrollmentMgmt.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListEnrollments)
			enrollmentMgmt.GET("/:id", handlers.GetEnrollment)
			enrollmentMgmt.PUT("/:id", middleware.ValidateRequest[models.UpdateEnrollmentRequest](), handlers.UpdateEnrollment)
			enrollmentMgmt.DELETE("/:id", handlers.DeleteEnrollment)
		}
		
		// Progress routes (protected - authenticated users only)
		progress := v1.Group("/progress")
		progress.Use(middleware.AuthMiddleware(jwtManager))
		{
			progress.POST("/complete", middleware.ValidateRequest[models.CompleteLessonRequest](), handlers.CompleteLesson)
			progress.GET("/my-progress", handlers.GetUserProgress)
			progress.GET("/courses/:course_id", handlers.GetCourseProgress)
		}
		
		// Course completion routes (protected - authenticated users)
		completions := v1.Group("/completions")
		completions.Use(middleware.AuthMiddleware(jwtManager))
		{
			completions.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListUserCompletions)
			completions.GET("/courses/:course_id", handlers.GetCourseCompletion)
		}
		
		// Prerequisite routes (protected - instructors and admins only)
		prerequisites := v1.Group("/prerequisites")
		prerequisites.Use(middleware.AuthMiddleware(jwtManager))
		prerequisites.Use(middleware.RequireRole(models.RoleInstructor))
		{
			prerequisites.POST("/", middleware.ValidateRequest[models.CreatePrerequisiteRequest](), handlers.CreatePrerequisite)
			prerequisites.GET("/courses/:course_id", handlers.ListPrerequisites)
			prerequisites.DELETE("/courses/:course_id/required/:required_course_id", handlers.DeletePrerequisite)
		}
		
		// Prerequisite check routes (protected - authenticated users)
		prerequisiteCheck := v1.Group("/prerequisites")
		prerequisiteCheck.Use(middleware.AuthMiddleware(jwtManager))
		{
			prerequisiteCheck.GET("/check/courses/:course_id", handlers.CheckPrerequisites)
		}
		
		// Certificate routes
		certificates := v1.Group("/certificates")
		{
			certificates.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListCertificates)
			certificates.GET("/:id", handlers.GetCertificate)
			certificates.GET("/verify/:id", handlers.VerifyCertificate)
		}
		
		// Certificate management routes (admin only)
		certificateMgmt := v1.Group("/certificates")
		certificateMgmt.Use(middleware.AuthMiddleware(jwtManager))
		certificateMgmt.Use(middleware.RequireRole(models.RoleAdmin))
		{
			certificateMgmt.POST("/", middleware.ValidateRequest[models.CreateCertificateRequest](), handlers.CreateCertificate)
		}
	}

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	logger.WithField("port", port).Info("Starting LMS server")
	logger.Fatal("Server stopped", logger.WithFields(map[string]interface{}{
		"error": r.Run(":" + port).Error(),
	}).Data)
}
