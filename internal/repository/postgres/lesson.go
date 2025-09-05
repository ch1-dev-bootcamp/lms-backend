package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// LessonRepository implements the LessonRepository interface
type LessonRepository struct {
	*BaseRepository[models.Lesson]
}

// NewLessonRepository creates a new lesson repository
func NewLessonRepository(db *sql.DB) *LessonRepository {
	return &LessonRepository{
		BaseRepository: NewBaseRepository[models.Lesson](db, "lessons"),
	}
}

// Create inserts a new lesson into the database
func (r *LessonRepository) Create(ctx context.Context, lesson *models.Lesson) error {
	query := `
		INSERT INTO lessons (id, course_id, title, content, order_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		lesson.ID,
		lesson.CourseID,
		lesson.Title,
		lesson.Content,
		lesson.OrderNumber,
		lesson.CreatedAt,
		lesson.UpdatedAt,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves a lesson by ID
func (r *LessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	query := `
		SELECT id, course_id, title, content, order_number, created_at, updated_at
		FROM lessons
		WHERE id = $1
	`
	
	lesson := &models.Lesson{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Content,
		&lesson.OrderNumber,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return lesson, nil
}

// Update updates an existing lesson
func (r *LessonRepository) Update(ctx context.Context, lesson *models.Lesson) error {
	query := `
		UPDATE lessons
		SET title = $2, content = $3, order_number = $4, updated_at = $5
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query,
		lesson.ID,
		lesson.Title,
		lesson.Content,
		lesson.OrderNumber,
		lesson.UpdatedAt,
	)
	
	return handleDatabaseError(err)
}

// List retrieves lessons with pagination
func (r *LessonRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.Lesson, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, course_id, title, content, order_number, created_at, updated_at
		FROM lessons
		ORDER BY order_number ASC, created_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanLesson)
}

// GetByCourse retrieves lessons by course with pagination
func (r *LessonRepository) GetByCourse(ctx context.Context, courseID uuid.UUID, pagination models.PaginationRequest) ([]models.Lesson, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, course_id, title, content, order_number, created_at, updated_at
		FROM lessons
		WHERE course_id = $1
		ORDER BY order_number ASC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	// Execute count query with course parameter
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, courseID).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	query, _ := buildPaginationQuery(baseQuery, pagination)
	
	// Execute query with course parameter
	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []models.Lesson
	for rows.Next() {
		lesson, err := r.scanLesson(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *lesson)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Build pagination response
	paginationResp := &models.PaginationResponse{
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		Total:      total,
		TotalPages: (total + pagination.PageSize - 1) / pagination.PageSize,
	}

	return results, paginationResp, nil
}

// GetByOrder retrieves a lesson by course and order number
func (r *LessonRepository) GetByOrder(ctx context.Context, courseID uuid.UUID, orderNumber int) (*models.Lesson, error) {
	query := `
		SELECT id, course_id, title, content, order_number, created_at, updated_at
		FROM lessons
		WHERE course_id = $1 AND order_number = $2
	`
	
	lesson := &models.Lesson{}
	err := r.db.QueryRowContext(ctx, query, courseID, orderNumber).Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Content,
		&lesson.OrderNumber,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return lesson, nil
}

// ReorderLessons updates the order of lessons in a course
func (r *LessonRepository) ReorderLessons(ctx context.Context, courseID uuid.UUID, lessonOrders map[uuid.UUID]int) error {
	// This would typically be implemented as a transaction
	// For now, we'll update each lesson individually
	for lessonID, order := range lessonOrders {
		query := `
			UPDATE lessons
			SET order_number = $2, updated_at = NOW()
			WHERE id = $1 AND course_id = $3
		`
		
		_, err := r.db.ExecContext(ctx, query, lessonID, order, courseID)
		if err != nil {
			return handleDatabaseError(err)
		}
	}
	
	return nil
}

// GetWithDetails retrieves lesson with additional details
func (r *LessonRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*models.LessonDetailResponse, error) {
	query := `
		SELECT 
			l.id, l.course_id, l.title, l.content, l.order_number, l.created_at, l.updated_at,
			c.title as course_title,
			30 as duration
		FROM lessons l
		LEFT JOIN courses c ON l.course_id = c.id
		WHERE l.id = $1
	`
	
	var lessonDetail models.LessonDetailResponse
	var courseTitle sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lessonDetail.ID,
		&lessonDetail.CourseID,
		&lessonDetail.Title,
		&lessonDetail.Content,
		&lessonDetail.OrderNumber,
		&lessonDetail.CreatedAt,
		&lessonDetail.UpdatedAt,
		&courseTitle,
		&lessonDetail.Duration,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	lessonDetail.CourseTitle = courseTitle.String
	
	return &lessonDetail, nil
}

// scanLesson scans a lesson from database rows
func (r *LessonRepository) scanLesson(rows *sql.Rows) (*models.Lesson, error) {
	lesson := &models.Lesson{}
	err := rows.Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Content,
		&lesson.OrderNumber,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return lesson, nil
}
