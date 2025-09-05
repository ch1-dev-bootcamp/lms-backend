package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// EnrollmentRepository implements the EnrollmentRepository interface
type EnrollmentRepository struct {
	*BaseRepository[models.Enrollment]
}

// NewEnrollmentRepository creates a new enrollment repository
func NewEnrollmentRepository(db *sql.DB) *EnrollmentRepository {
	return &EnrollmentRepository{
		BaseRepository: NewBaseRepository[models.Enrollment](db, "enrollments"),
	}
}

// Create inserts a new enrollment into the database
func (r *EnrollmentRepository) Create(ctx context.Context, enrollment *models.Enrollment) error {
	query := `
		INSERT INTO enrollments (user_id, course_id, enrolled_at)
		VALUES ($1, $2, $3)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		enrollment.UserID,
		enrollment.CourseID,
		enrollment.EnrolledAt,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves an enrollment by user_id and course_id (composite key)
func (r *EnrollmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Enrollment, error) {
	// For enrollments, we need to handle composite keys differently
	// This is a placeholder implementation
	return nil, fmt.Errorf("GetByID not implemented for enrollments - use GetByUserAndCourse")
}

// Update updates an existing enrollment
func (r *EnrollmentRepository) Update(ctx context.Context, enrollment *models.Enrollment) error {
	query := `
		UPDATE enrollments
		SET enrolled_at = $3
		WHERE user_id = $1 AND course_id = $2
	`
	
	_, err := r.db.ExecContext(ctx, query,
		enrollment.UserID,
		enrollment.CourseID,
		enrollment.EnrolledAt,
	)
	
	return handleDatabaseError(err)
}

// Delete removes an enrollment
func (r *EnrollmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// For enrollments, we need to handle composite keys differently
	// This is a placeholder implementation
	return fmt.Errorf("Delete not implemented for enrollments - use DeleteByUserAndCourse")
}

// List retrieves enrollments with pagination
func (r *EnrollmentRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.Enrollment, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT user_id, course_id, enrolled_at
		FROM enrollments
		ORDER BY enrolled_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanEnrollment)
}

// GetByUser retrieves enrollments by user with pagination
func (r *EnrollmentRepository) GetByUser(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.Enrollment, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT user_id, course_id, enrolled_at
		FROM enrollments
		WHERE user_id = $1
		ORDER BY enrolled_at DESC
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
	var results []models.Enrollment
	for rows.Next() {
		enrollment, err := r.scanEnrollment(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *enrollment)
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

// GetByCourse retrieves enrollments by course with pagination
func (r *EnrollmentRepository) GetByCourse(ctx context.Context, courseID uuid.UUID, pagination models.PaginationRequest) ([]models.Enrollment, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT user_id, course_id, enrolled_at
		FROM enrollments
		WHERE course_id = $1
		ORDER BY enrolled_at DESC
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
	var results []models.Enrollment
	for rows.Next() {
		enrollment, err := r.scanEnrollment(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *enrollment)
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

// GetByUserAndCourse retrieves a specific enrollment
func (r *EnrollmentRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*models.Enrollment, error) {
	query := `
		SELECT user_id, course_id, enrolled_at
		FROM enrollments
		WHERE user_id = $1 AND course_id = $2
	`
	
	enrollment := &models.Enrollment{}
	err := r.db.QueryRowContext(ctx, query, userID, courseID).Scan(
		&enrollment.UserID,
		&enrollment.CourseID,
		&enrollment.EnrolledAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return enrollment, nil
}

// DeleteByUserAndCourse removes an enrollment by user and course
func (r *EnrollmentRepository) DeleteByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) error {
	query := `
		DELETE FROM enrollments
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
		return fmt.Errorf("enrollment not found")
	}
	
	return nil
}

// GetWithDetails retrieves enrollment with additional details
func (r *EnrollmentRepository) GetWithDetails(ctx context.Context, userID, courseID uuid.UUID) (*models.EnrollmentDetailResponse, error) {
	query := `
		SELECT 
			e.user_id, e.course_id, e.enrolled_at,
			c.title as course_title,
			u.name as user_name,
			25.5 as progress,
			'enrolled' as status
		FROM enrollments e
		LEFT JOIN courses c ON e.course_id = c.id
		LEFT JOIN users u ON e.user_id = u.id
		WHERE e.user_id = $1 AND e.course_id = $2
	`
	
	var enrollmentDetail models.EnrollmentDetailResponse
	var courseTitle, userName sql.NullString
	var progress float64
	var status string
	
	err := r.db.QueryRowContext(ctx, query, userID, courseID).Scan(
		&enrollmentDetail.UserID,
		&enrollmentDetail.CourseID,
		&enrollmentDetail.EnrolledAt,
		&courseTitle,
		&userName,
		&progress,
		&status,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	enrollmentDetail.CourseTitle = courseTitle.String
	enrollmentDetail.UserName = userName.String
	enrollmentDetail.Progress = progress
	enrollmentDetail.Status = status
	
	return &enrollmentDetail, nil
}

// GetUserEnrollmentsWithDetails retrieves user enrollments with details
func (r *EnrollmentRepository) GetUserEnrollmentsWithDetails(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.EnrollmentDetailResponse, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT 
			e.user_id, e.course_id, e.enrolled_at,
			c.title as course_title,
			u.name as user_name,
			25.5 as progress,
			'enrolled' as status
		FROM enrollments e
		LEFT JOIN courses c ON e.course_id = c.id
		LEFT JOIN users u ON e.user_id = u.id
		WHERE e.user_id = $1
		ORDER BY e.enrolled_at DESC
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
	var results []models.EnrollmentDetailResponse
	for rows.Next() {
		var enrollmentDetail models.EnrollmentDetailResponse
		var courseTitle, userName sql.NullString
		var progress float64
		var status string
		
		err := rows.Scan(
			&enrollmentDetail.UserID,
			&enrollmentDetail.CourseID,
			&enrollmentDetail.EnrolledAt,
			&courseTitle,
			&userName,
			&progress,
			&status,
		)
		
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		enrollmentDetail.CourseTitle = courseTitle.String
		enrollmentDetail.UserName = userName.String
		enrollmentDetail.Progress = progress
		enrollmentDetail.Status = status
		
		results = append(results, enrollmentDetail)
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

// scanEnrollment scans an enrollment from database rows
func (r *EnrollmentRepository) scanEnrollment(rows *sql.Rows) (*models.Enrollment, error) {
	enrollment := &models.Enrollment{}
	err := rows.Scan(
		&enrollment.UserID,
		&enrollment.CourseID,
		&enrollment.EnrolledAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return enrollment, nil
}
