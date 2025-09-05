package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// CourseCompletionRepository implements the CourseCompletionRepository interface
type CourseCompletionRepository struct {
	*BaseRepository[models.CourseCompletion]
}

// NewCourseCompletionRepository creates a new course completion repository
func NewCourseCompletionRepository(db *sql.DB) *CourseCompletionRepository {
	return &CourseCompletionRepository{
		BaseRepository: NewBaseRepository[models.CourseCompletion](db, "course_completions"),
	}
}

// Create inserts a new course completion into the database
func (r *CourseCompletionRepository) Create(ctx context.Context, completion *models.CourseCompletion) error {
	query := `
		INSERT INTO course_completions (user_id, course_id, completed_at, completion_rate)
		VALUES ($1, $2, $3, $4)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		completion.UserID,
		completion.CourseID,
		completion.CompletedAt,
		completion.CompletionRate,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves a course completion by user_id and course_id (composite key)
func (r *CourseCompletionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CourseCompletion, error) {
	return nil, fmt.Errorf("GetByID not implemented for course completions - use GetByUserAndCourse")
}

// Update updates an existing course completion
func (r *CourseCompletionRepository) Update(ctx context.Context, completion *models.CourseCompletion) error {
	query := `
		UPDATE course_completions
		SET completed_at = $3, completion_rate = $4
		WHERE user_id = $1 AND course_id = $2
	`
	
	_, err := r.db.ExecContext(ctx, query,
		completion.UserID,
		completion.CourseID,
		completion.CompletedAt,
		completion.CompletionRate,
	)
	
	return handleDatabaseError(err)
}

// Delete removes a course completion
func (r *CourseCompletionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("Delete not implemented for course completions - use DeleteByUserAndCourse")
}

// List retrieves course completions with pagination
func (r *CourseCompletionRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.CourseCompletion, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT user_id, course_id, completed_at, completion_rate
		FROM course_completions
		ORDER BY completed_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanCourseCompletion)
}

// GetByUser retrieves course completions by user with pagination
func (r *CourseCompletionRepository) GetByUser(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.CourseCompletion, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT user_id, course_id, completed_at, completion_rate
		FROM course_completions
		WHERE user_id = $1
		ORDER BY completed_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	// Execute count query with user parameter
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	query, _ := buildPaginationQuery(baseQuery, pagination)
	
	// Execute query with user parameter
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []models.CourseCompletion
	for rows.Next() {
		completion, err := r.scanCourseCompletion(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *completion)
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

// GetByUserAndCourse retrieves a specific course completion
func (r *CourseCompletionRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*models.CourseCompletion, error) {
	query := `
		SELECT user_id, course_id, completed_at, completion_rate
		FROM course_completions
		WHERE user_id = $1 AND course_id = $2
	`
	
	completion := &models.CourseCompletion{}
	err := r.db.QueryRowContext(ctx, query, userID, courseID).Scan(
		&completion.UserID,
		&completion.CourseID,
		&completion.CompletedAt,
		&completion.CompletionRate,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return completion, nil
}

// GetUserCompletionsWithDetails retrieves user course completions with details
func (r *CourseCompletionRepository) GetUserCompletionsWithDetails(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.CourseCompletionResponse, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT 
			cc.user_id, cc.course_id, cc.completed_at, cc.completion_rate,
			c.title as course_title,
			(SELECT COUNT(*) FROM lessons WHERE course_id = c.id) as total_lessons,
			(SELECT COUNT(*) FROM progress p JOIN lessons l ON p.lesson_id = l.id WHERE p.user_id = cc.user_id AND l.course_id = c.id) as completed_lessons
		FROM course_completions cc
		JOIN courses c ON cc.course_id = c.id
		WHERE cc.user_id = $1
		ORDER BY cc.completed_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	// Execute count query with user parameter
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	query, _ := buildPaginationQuery(baseQuery, pagination)
	
	// Execute query with user parameter
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []models.CourseCompletionResponse
	for rows.Next() {
		var completion models.CourseCompletionResponse
		
		err := rows.Scan(
			&completion.UserID,
			&completion.CourseID,
			&completion.CompletedAt,
			&completion.CompletionRate,
			&completion.CourseTitle,
			&completion.TotalLessons,
			&completion.CompletedLessons,
		)
		
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		results = append(results, completion)
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

// DeleteByUserAndCourse removes a course completion by user and course
func (r *CourseCompletionRepository) DeleteByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) error {
	query := `
		DELETE FROM course_completions
		WHERE user_id = $1 AND course_id = $2
	`
	
	result, err := r.db.ExecContext(ctx, query, userID, courseID)
	if err != nil {
		return handleDatabaseError(err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return handleDatabaseError(err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("course completion not found")
	}
	
	return nil
}

// scanCourseCompletion scans a course completion from database rows
func (r *CourseCompletionRepository) scanCourseCompletion(rows *sql.Rows) (*models.CourseCompletion, error) {
	completion := &models.CourseCompletion{}
	err := rows.Scan(
		&completion.UserID,
		&completion.CourseID,
		&completion.CompletedAt,
		&completion.CompletionRate,
	)
	
	if err != nil {
		return nil, err
	}
	
	return completion, nil
}
