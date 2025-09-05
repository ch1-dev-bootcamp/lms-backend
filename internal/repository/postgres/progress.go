package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// ProgressRepository implements the ProgressRepository interface
type ProgressRepository struct {
	*BaseRepository[models.Progress]
}

// NewProgressRepository creates a new progress repository
func NewProgressRepository(db *sql.DB) *ProgressRepository {
	return &ProgressRepository{
		BaseRepository: NewBaseRepository[models.Progress](db, "progress"),
	}
}

// Create inserts a new progress record into the database
func (r *ProgressRepository) Create(ctx context.Context, progress *models.Progress) error {
	query := `
		INSERT INTO progress (user_id, lesson_id, completed_at)
		VALUES ($1, $2, $3)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		progress.UserID,
		progress.LessonID,
		progress.CompletedAt,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves a progress record by user_id and lesson_id (composite key)
func (r *ProgressRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Progress, error) {
	return nil, fmt.Errorf("GetByID not implemented for progress - use GetByUserAndLesson")
}

// Update updates an existing progress record
func (r *ProgressRepository) Update(ctx context.Context, progress *models.Progress) error {
	query := `
		UPDATE progress
		SET completed_at = $3
		WHERE user_id = $1 AND lesson_id = $2
	`
	
	_, err := r.db.ExecContext(ctx, query,
		progress.UserID,
		progress.LessonID,
		progress.CompletedAt,
	)
	
	return handleDatabaseError(err)
}

// Delete removes a progress record
func (r *ProgressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("Delete not implemented for progress - use DeleteByUserAndLesson")
}

// List retrieves progress records with pagination
func (r *ProgressRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.Progress, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT user_id, lesson_id, completed_at
		FROM progress
		ORDER BY completed_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanProgress)
}

// GetByUser retrieves progress records by user
func (r *ProgressRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.Progress, error) {
	query := `
		SELECT user_id, lesson_id, completed_at
		FROM progress
		WHERE user_id = $1
		ORDER BY completed_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []models.Progress
	for rows.Next() {
		progress, err := r.scanProgress(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *progress)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetByLesson retrieves progress records by lesson
func (r *ProgressRepository) GetByLesson(ctx context.Context, lessonID uuid.UUID) ([]models.Progress, error) {
	query := `
		SELECT user_id, lesson_id, completed_at
		FROM progress
		WHERE lesson_id = $1
		ORDER BY completed_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []models.Progress
	for rows.Next() {
		progress, err := r.scanProgress(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *progress)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetByUserAndLesson retrieves a specific progress record
func (r *ProgressRepository) GetByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) (*models.Progress, error) {
	query := `
		SELECT user_id, lesson_id, completed_at
		FROM progress
		WHERE user_id = $1 AND lesson_id = $2
	`
	
	progress := &models.Progress{}
	err := r.db.QueryRowContext(ctx, query, userID, lessonID).Scan(
		&progress.UserID,
		&progress.LessonID,
		&progress.CompletedAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return progress, nil
}

// GetUserProgress retrieves user progress with details
func (r *ProgressRepository) GetUserProgress(ctx context.Context, userID uuid.UUID) ([]models.ProgressDetailResponse, error) {
	query := `
		SELECT 
			p.user_id, c.id as course_id, c.title as course_title,
			(SELECT COUNT(*) FROM lessons WHERE course_id = c.id) as total_lessons,
			COUNT(p.lesson_id) as completed_lessons,
			(COUNT(p.lesson_id)::float / (SELECT COUNT(*) FROM lessons WHERE course_id = c.id) * 100) as completion_rate,
			MAX(p.completed_at) as last_activity,
			e.enrolled_at
		FROM progress p
		JOIN lessons l ON p.lesson_id = l.id
		JOIN courses c ON l.course_id = c.id
		JOIN enrollments e ON e.user_id = p.user_id AND e.course_id = c.id
		WHERE p.user_id = $1
		GROUP BY p.user_id, c.id, c.title, e.enrolled_at
		ORDER BY last_activity DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []models.ProgressDetailResponse
	for rows.Next() {
		var progressDetail models.ProgressDetailResponse
		var lastActivity, enrolledAt sql.NullTime
		
		err := rows.Scan(
			&progressDetail.UserID,
			&progressDetail.CourseID,
			&progressDetail.CourseTitle,
			&progressDetail.TotalLessons,
			&progressDetail.CompletedLessons,
			&progressDetail.CompletionRate,
			&lastActivity,
			&enrolledAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		if lastActivity.Valid {
			progressDetail.LastActivity = lastActivity.Time
		}
		if enrolledAt.Valid {
			progressDetail.EnrolledAt = enrolledAt.Time
		}
		
		results = append(results, progressDetail)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetCourseProgress retrieves course progress for a user
func (r *ProgressRepository) GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (*models.ProgressDetailResponse, error) {
	query := `
		SELECT 
			$1 as user_id, $2 as course_id, c.title as course_title,
			(SELECT COUNT(*) FROM lessons WHERE course_id = c.id) as total_lessons,
			COUNT(p.lesson_id) as completed_lessons,
			(COUNT(p.lesson_id)::float / (SELECT COUNT(*) FROM lessons WHERE course_id = c.id) * 100) as completion_rate,
			MAX(p.completed_at) as last_activity,
			e.enrolled_at
		FROM courses c
		LEFT JOIN progress p ON p.user_id = $1 AND p.lesson_id IN (SELECT id FROM lessons WHERE course_id = c.id)
		LEFT JOIN enrollments e ON e.user_id = $1 AND e.course_id = c.id
		WHERE c.id = $2
		GROUP BY c.id, c.title, e.enrolled_at
	`
	
	var progressDetail models.ProgressDetailResponse
	var lastActivity, enrolledAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, userID, courseID).Scan(
		&progressDetail.UserID,
		&progressDetail.CourseID,
		&progressDetail.CourseTitle,
		&progressDetail.TotalLessons,
		&progressDetail.CompletedLessons,
		&progressDetail.CompletionRate,
		&lastActivity,
		&enrolledAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	if lastActivity.Valid {
		progressDetail.LastActivity = lastActivity.Time
	}
	if enrolledAt.Valid {
		progressDetail.EnrolledAt = enrolledAt.Time
	}
	
	return &progressDetail, nil
}

// GetCompletionRate calculates completion rate for a user in a course
func (r *ProgressRepository) GetCompletionRate(ctx context.Context, userID, courseID uuid.UUID) (float64, error) {
	query := `
		SELECT 
			(COUNT(p.lesson_id)::float / (SELECT COUNT(*) FROM lessons WHERE course_id = $2) * 100) as completion_rate
		FROM progress p
		JOIN lessons l ON p.lesson_id = l.id
		WHERE p.user_id = $1 AND l.course_id = $2
	`
	
	var completionRate float64
	err := r.db.QueryRowContext(ctx, query, userID, courseID).Scan(&completionRate)
	if err != nil {
		return 0, handleDatabaseError(err)
	}
	
	return completionRate, nil
}

// scanProgress scans a progress record from database rows
func (r *ProgressRepository) scanProgress(rows *sql.Rows) (*models.Progress, error) {
	progress := &models.Progress{}
	err := rows.Scan(
		&progress.UserID,
		&progress.LessonID,
		&progress.CompletedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return progress, nil
}
