package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// PrerequisiteRepository implements the PrerequisiteRepository interface
type PrerequisiteRepository struct {
	*BaseRepository[models.Prerequisite]
}

// NewPrerequisiteRepository creates a new prerequisite repository
func NewPrerequisiteRepository(db *sql.DB) *PrerequisiteRepository {
	return &PrerequisiteRepository{
		BaseRepository: NewBaseRepository[models.Prerequisite](db, "prerequisites"),
	}
}

// Create inserts a new prerequisite into the database
func (r *PrerequisiteRepository) Create(ctx context.Context, prerequisite *models.Prerequisite) error {
	query := `
		INSERT INTO prerequisites (course_id, prerequisite_course_id)
		VALUES ($1, $2)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		prerequisite.CourseID,
		prerequisite.RequiredCourseID,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves a prerequisite by course_id and prerequisite_course_id (composite key)
func (r *PrerequisiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Prerequisite, error) {
	return nil, fmt.Errorf("GetByID not implemented for prerequisites - use GetByCourseAndPrerequisite")
}

// Update updates an existing prerequisite
func (r *PrerequisiteRepository) Update(ctx context.Context, prerequisite *models.Prerequisite) error {
	query := `
		UPDATE prerequisites
		SET prerequisite_course_id = $3
		WHERE course_id = $1 AND prerequisite_course_id = $2
	`
	
	_, err := r.db.ExecContext(ctx, query,
		prerequisite.CourseID,
		prerequisite.RequiredCourseID,
		prerequisite.RequiredCourseID, // This is a placeholder - actual implementation would need different logic
	)
	
	return handleDatabaseError(err)
}

// Delete removes a prerequisite
func (r *PrerequisiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("Delete not implemented for prerequisites - use DeleteByCourseAndPrerequisite")
}

// List retrieves prerequisites with pagination
func (r *PrerequisiteRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.Prerequisite, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT course_id, prerequisite_course_id
		FROM prerequisites
		ORDER BY course_id, prerequisite_course_id
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanPrerequisite)
}

// GetByCourse retrieves prerequisites for a course
func (r *PrerequisiteRepository) GetByCourse(ctx context.Context, courseID uuid.UUID) ([]models.Prerequisite, error) {
	query := `
		SELECT course_id, prerequisite_course_id
		FROM prerequisites
		WHERE course_id = $1
		ORDER BY prerequisite_course_id
	`
	
	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []models.Prerequisite
	for rows.Next() {
		prerequisite, err := r.scanPrerequisite(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *prerequisite)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetPrerequisiteCourses retrieves prerequisite courses for a course
func (r *PrerequisiteRepository) GetPrerequisiteCourses(ctx context.Context, courseID uuid.UUID) ([]models.Course, error) {
	query := `
		SELECT c.id, c.title, c.description, c.instructor_id, c.status, c.created_at, c.updated_at
		FROM courses c
		JOIN prerequisites p ON c.id = p.prerequisite_course_id
		WHERE p.course_id = $1
		ORDER BY c.title
	`
	
	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []models.Course
	for rows.Next() {
		course := &models.Course{}
		err := rows.Scan(
			&course.ID,
			&course.Title,
			&course.Description,
			&course.InstructorID,
			&course.Status,
			&course.CreatedAt,
			&course.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		results = append(results, *course)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// CheckPrerequisites checks if a user has completed all prerequisites for a course
func (r *PrerequisiteRepository) CheckPrerequisites(ctx context.Context, userID, courseID uuid.UUID) (bool, []uuid.UUID, error) {
	query := `
		SELECT p.prerequisite_course_id
		FROM prerequisites p
		LEFT JOIN progress pr ON pr.user_id = $1 AND pr.lesson_id IN (
			SELECT l.id FROM lessons l WHERE l.course_id = p.prerequisite_course_id
		)
		WHERE p.course_id = $2
		GROUP BY p.prerequisite_course_id
		HAVING COUNT(pr.lesson_id) < (
			SELECT COUNT(*) FROM lessons l WHERE l.course_id = p.prerequisite_course_id
		)
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID, courseID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var missingPrerequisites []uuid.UUID
	for rows.Next() {
		var prerequisiteID uuid.UUID
		err := rows.Scan(&prerequisiteID)
		if err != nil {
			return false, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		missingPrerequisites = append(missingPrerequisites, prerequisiteID)
	}

	if err = rows.Err(); err != nil {
		return false, nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return len(missingPrerequisites) == 0, missingPrerequisites, nil
}

// scanPrerequisite scans a prerequisite from database rows
func (r *PrerequisiteRepository) scanPrerequisite(rows *sql.Rows) (*models.Prerequisite, error) {
	prerequisite := &models.Prerequisite{}
	err := rows.Scan(
		&prerequisite.CourseID,
		&prerequisite.RequiredCourseID,
	)
	
	if err != nil {
		return nil, err
	}
	
	return prerequisite, nil
}
