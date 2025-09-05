package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// BaseRepository defines common operations for all repositories
type BaseRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pagination models.PaginationRequest) ([]T, *models.PaginationResponse, error)
}

// UserRepository defines operations for user management
type UserRepository interface {
	BaseRepository[models.User]
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	GetByRole(ctx context.Context, role string, pagination models.PaginationRequest) ([]models.User, *models.PaginationResponse, error)
}

// CourseRepository defines operations for course management
type CourseRepository interface {
	BaseRepository[models.Course]
	GetByInstructor(ctx context.Context, instructorID uuid.UUID, pagination models.PaginationRequest) ([]models.Course, *models.PaginationResponse, error)
	GetByStatus(ctx context.Context, status string, pagination models.PaginationRequest) ([]models.Course, *models.PaginationResponse, error)
	Search(ctx context.Context, query string, pagination models.PaginationRequest) ([]models.Course, *models.PaginationResponse, error)
	GetWithDetails(ctx context.Context, id uuid.UUID) (*models.CourseDetailResponse, error)
}

// LessonRepository defines operations for lesson management
type LessonRepository interface {
	BaseRepository[models.Lesson]
	GetByCourse(ctx context.Context, courseID uuid.UUID, pagination models.PaginationRequest) ([]models.Lesson, *models.PaginationResponse, error)
	GetByOrder(ctx context.Context, courseID uuid.UUID, orderNumber int) (*models.Lesson, error)
	ReorderLessons(ctx context.Context, courseID uuid.UUID, lessonOrders map[uuid.UUID]int) error
	GetWithDetails(ctx context.Context, id uuid.UUID) (*models.LessonDetailResponse, error)
}

// EnrollmentRepository defines operations for enrollment management
type EnrollmentRepository interface {
	BaseRepository[models.Enrollment]
	GetByUser(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.Enrollment, *models.PaginationResponse, error)
	GetByCourse(ctx context.Context, courseID uuid.UUID, pagination models.PaginationRequest) ([]models.Enrollment, *models.PaginationResponse, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*models.Enrollment, error)
	DeleteByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) error
	GetWithDetails(ctx context.Context, userID, courseID uuid.UUID) (*models.EnrollmentDetailResponse, error)
	GetUserEnrollmentsWithDetails(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.EnrollmentDetailResponse, *models.PaginationResponse, error)
}

// ProgressRepository defines operations for progress tracking
type ProgressRepository interface {
	BaseRepository[models.Progress]
	GetByUser(ctx context.Context, userID uuid.UUID) ([]models.Progress, error)
	GetByLesson(ctx context.Context, lessonID uuid.UUID) ([]models.Progress, error)
	GetByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) (*models.Progress, error)
	GetUserProgress(ctx context.Context, userID uuid.UUID) ([]models.ProgressDetailResponse, error)
	GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (*models.ProgressDetailResponse, error)
	GetCompletionRate(ctx context.Context, userID, courseID uuid.UUID) (float64, error)
}

// CertificateRepository defines operations for certificate management
type CertificateRepository interface {
	BaseRepository[models.Certificate]
	GetByUser(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.Certificate, *models.PaginationResponse, error)
	GetByCourse(ctx context.Context, courseID uuid.UUID, pagination models.PaginationRequest) ([]models.Certificate, *models.PaginationResponse, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*models.Certificate, error)
	GetByCode(ctx context.Context, code string) (*models.Certificate, error)
	GetWithDetails(ctx context.Context, id uuid.UUID) (*models.CertificateDetailResponse, error)
	VerifyCertificate(ctx context.Context, id uuid.UUID) (*models.VerifyCertificateResponse, error)
}

// PrerequisiteRepository defines operations for course prerequisites
type PrerequisiteRepository interface {
	BaseRepository[models.Prerequisite]
	GetByCourse(ctx context.Context, courseID uuid.UUID) ([]models.Prerequisite, error)
	GetPrerequisiteCourses(ctx context.Context, courseID uuid.UUID) ([]models.Course, error)
	DeleteByCourseAndPrerequisite(ctx context.Context, courseID, requiredCourseID uuid.UUID) error
	CheckPrerequisites(ctx context.Context, userID, courseID uuid.UUID) (bool, []uuid.UUID, error)
}

// CourseCompletionRepository defines operations for course completion tracking
type CourseCompletionRepository interface {
	BaseRepository[models.CourseCompletion]
	GetByUser(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.CourseCompletion, *models.PaginationResponse, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*models.CourseCompletion, error)
	GetUserCompletionsWithDetails(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.CourseCompletionResponse, *models.PaginationResponse, error)
	DeleteByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) error
}

// RepositoryManager manages all repositories
type RepositoryManager interface {
	User() UserRepository
	Course() CourseRepository
	Lesson() LessonRepository
	Enrollment() EnrollmentRepository
	Progress() ProgressRepository
	Certificate() CertificateRepository
	Prerequisite() PrerequisiteRepository
	CourseCompletion() CourseCompletionRepository
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(RepositoryManager) error) error
	Close() error
}
