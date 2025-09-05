package main

import (
	"github.com/gin-gonic/gin"
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

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.Logging())
	r.Use(middleware.CORS())
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
		
		// User routes
		users := v1.Group("/users")
		{
			users.GET("/profile", handlers.GetProfile)
			users.PUT("/profile", middleware.ValidateRequest[models.UpdateProfileRequest](), handlers.UpdateProfile)
		}
		
		// Course routes
		courses := v1.Group("/courses")
		{
			courses.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListCourses)
			courses.POST("/", middleware.ValidateRequest[models.CreateCourseRequest](), handlers.CreateCourse)
			courses.GET("/:id", handlers.GetCourse)
			courses.PUT("/:id", middleware.ValidateRequest[models.UpdateCourseRequest](), handlers.UpdateCourse)
			courses.DELETE("/:id", handlers.DeleteCourse)
		}
		
		// Lesson routes
		lessons := v1.Group("/lessons")
		{
			lessons.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListLessons)
			lessons.POST("/", middleware.ValidateRequest[models.CreateLessonRequest](), handlers.CreateLesson)
			lessons.GET("/:id", handlers.GetLesson)
			lessons.PUT("/:id", middleware.ValidateRequest[models.UpdateLessonRequest](), handlers.UpdateLesson)
			lessons.DELETE("/:id", handlers.DeleteLesson)
		}
		
		// Enrollment routes
		enrollments := v1.Group("/enrollments")
		{
			enrollments.POST("/", middleware.ValidateRequest[models.CreateEnrollmentRequest](), handlers.Enroll)
			enrollments.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListEnrollments)
			enrollments.GET("/:id", handlers.GetEnrollment)
		}
		
		// Progress routes
		progress := v1.Group("/progress")
		{
			progress.POST("/complete", middleware.ValidateRequest[models.CompleteLessonRequest](), handlers.CompleteLesson)
			progress.GET("/", handlers.GetProgress)
			progress.GET("/:user_id", handlers.GetUserProgress)
		}
		
		// Certificate routes
		certificates := v1.Group("/certificates")
		{
			certificates.GET("/", middleware.ValidateQuery[models.PaginationRequest](), handlers.ListCertificates)
			certificates.GET("/:id", handlers.GetCertificate)
			certificates.GET("/verify/:id", handlers.VerifyCertificate)
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
