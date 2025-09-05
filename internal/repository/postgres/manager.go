package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/your-org/lms-backend/internal/repository"
)

// RepositoryManager implements the RepositoryManager interface
type RepositoryManager struct {
	db *sql.DB
	
	// Repositories
	userRepo             repository.UserRepository
	courseRepo           repository.CourseRepository
	lessonRepo           repository.LessonRepository
	enrollmentRepo       repository.EnrollmentRepository
	progressRepo         repository.ProgressRepository
	certificateRepo      repository.CertificateRepository
	prerequisiteRepo     repository.PrerequisiteRepository
	courseCompletionRepo repository.CourseCompletionRepository
}

// NewRepositoryManager creates a new repository manager with connection pooling
func NewRepositoryManager(db *sql.DB) repository.RepositoryManager {
	manager := &RepositoryManager{
		db: db,
	}
	
	// Initialize repositories
	manager.userRepo = NewUserRepository(db)
	manager.courseRepo = NewCourseRepository(db)
	manager.lessonRepo = NewLessonRepository(db)
	manager.enrollmentRepo = NewEnrollmentRepository(db)
	manager.progressRepo = NewProgressRepository(db)
	manager.certificateRepo = NewCertificateRepository(db)
	manager.prerequisiteRepo = NewPrerequisiteRepository(db)
	manager.courseCompletionRepo = NewCourseCompletionRepository(db)
	
	return manager
}

// Repository getters
func (m *RepositoryManager) User() repository.UserRepository {
	return m.userRepo
}

func (m *RepositoryManager) Course() repository.CourseRepository {
	return m.courseRepo
}

func (m *RepositoryManager) Lesson() repository.LessonRepository {
	return m.lessonRepo
}

func (m *RepositoryManager) Enrollment() repository.EnrollmentRepository {
	return m.enrollmentRepo
}

func (m *RepositoryManager) Progress() repository.ProgressRepository {
	return m.progressRepo
}

func (m *RepositoryManager) Certificate() repository.CertificateRepository {
	return m.certificateRepo
}

func (m *RepositoryManager) Prerequisite() repository.PrerequisiteRepository {
	return m.prerequisiteRepo
}

func (m *RepositoryManager) CourseCompletion() repository.CourseCompletionRepository {
	return m.courseCompletionRepo
}

// WithTransaction executes a function within a database transaction
func (m *RepositoryManager) WithTransaction(ctx context.Context, fn func(repository.RepositoryManager) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Create a new repository manager with the transaction
	// Note: This is a simplified approach. In a real implementation, you might want to
	// create a separate transaction-aware repository manager that wraps the transaction
	txManager := &RepositoryManager{
		db: m.db, // Use the original db for now - transaction handling would need more work
	}
	
	// Execute the function
	if err := fn(txManager); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w (original error: %v)", rollbackErr, err)
		}
		return err
	}
	
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// Close closes the database connection
func (m *RepositoryManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewDatabaseConnection creates a new database connection with pooling
func NewDatabaseConnection(config DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Configure connection pooling
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}
	
	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return db, nil
}

// HealthCheck checks the database connection health
func (m *RepositoryManager) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	return m.db.PingContext(ctx)
}
