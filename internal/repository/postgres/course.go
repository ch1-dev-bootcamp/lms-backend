package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// CourseRepository implements the CourseRepository interface
type CourseRepository struct {
	*BaseRepository[models.Course]
}

// NewCourseRepository creates a new course repository
func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{
		BaseRepository: NewBaseRepository[models.Course](db, "courses"),
	}
}

// Create inserts a new course into the database
func (r *CourseRepository) Create(ctx context.Context, course *models.Course) error {
	query := `
		INSERT INTO courses (id, title, description, instructor_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		course.ID,
		course.Title,
		course.Description,
		course.InstructorID,
		course.Status,
		course.CreatedAt,
		course.UpdatedAt,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves a course by ID
func (r *CourseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	query := `
		SELECT id, title, description, instructor_id, status, created_at, updated_at
		FROM courses
		WHERE id = $1
	`
	
	course := &models.Course{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&course.ID,
		&course.Title,
		&course.Description,
		&course.InstructorID,
		&course.Status,
		&course.CreatedAt,
		&course.UpdatedAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return course, nil
}

// Update updates an existing course
func (r *CourseRepository) Update(ctx context.Context, course *models.Course) error {
	query := `
		UPDATE courses
		SET title = $2, description = $3, status = $4, updated_at = $5
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query,
		course.ID,
		course.Title,
		course.Description,
		course.Status,
		course.UpdatedAt,
	)
	
	return handleDatabaseError(err)
}

// List retrieves courses with pagination
func (r *CourseRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.Course, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, title, description, instructor_id, status, created_at, updated_at
		FROM courses
		ORDER BY created_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanCourse)
}

// GetByInstructor retrieves courses by instructor with pagination
func (r *CourseRepository) GetByInstructor(ctx context.Context, instructorID uuid.UUID, pagination models.PaginationRequest) ([]models.Course, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, title, description, instructor_id, status, created_at, updated_at
		FROM courses
		WHERE instructor_id = $1
		ORDER BY created_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	// Execute count query with instructor parameter
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, instructorID).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	query, _ := buildPaginationQuery(baseQuery, pagination)
	
	// Execute query with instructor parameter
	rows, err := r.db.QueryContext(ctx, query, instructorID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []models.Course
	for rows.Next() {
		course, err := r.scanCourse(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *course)
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

// GetByStatus retrieves courses by status with pagination
func (r *CourseRepository) GetByStatus(ctx context.Context, status string, pagination models.PaginationRequest) ([]models.Course, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, title, description, instructor_id, status, created_at, updated_at
		FROM courses
		WHERE status = $1
		ORDER BY created_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	// Execute count query with status parameter
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	query, _ := buildPaginationQuery(baseQuery, pagination)
	
	// Execute query with status parameter
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []models.Course
	for rows.Next() {
		course, err := r.scanCourse(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *course)
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

// Search searches courses by title and description
func (r *CourseRepository) Search(ctx context.Context, query string, pagination models.PaginationRequest) ([]models.Course, *models.PaginationResponse, error) {
	searchTerm := "%" + strings.ToLower(query) + "%"
	
	baseQuery := `
		SELECT id, title, description, instructor_id, status, created_at, updated_at
		FROM courses
		WHERE LOWER(title) LIKE $1 OR LOWER(description) LIKE $1
		ORDER BY created_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	// Execute count query with search parameter
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, searchTerm).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	paginatedQuery, _ := buildPaginationQuery(baseQuery, pagination)
	
	// Execute query with search parameter
	rows, err := r.db.QueryContext(ctx, paginatedQuery, searchTerm)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []models.Course
	for rows.Next() {
		course, err := r.scanCourse(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *course)
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

// GetWithDetails retrieves course with additional details
func (r *CourseRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*models.CourseDetailResponse, error) {
	query := `
		SELECT 
			c.id, c.title, c.description, c.instructor_id, c.status, c.created_at, c.updated_at,
			u.name as instructor_name,
			(SELECT COUNT(*) FROM lessons WHERE course_id = c.id) as lesson_count,
			(SELECT COUNT(*) FROM enrollments WHERE course_id = c.id) as enrollment_count
		FROM courses c
		LEFT JOIN users u ON c.instructor_id = u.id
		WHERE c.id = $1
	`
	
	var courseDetail models.CourseDetailResponse
	var instructorName sql.NullString
	var lessonCount, enrollmentCount int
	var description sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&courseDetail.ID,
		&courseDetail.Title,
		&description,
		&courseDetail.InstructorID,
		&courseDetail.Status,
		&courseDetail.CreatedAt,
		&courseDetail.UpdatedAt,
		&instructorName,
		&lessonCount,
		&enrollmentCount,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	if description.Valid {
		courseDetail.Description = &description.String
	}
	courseDetail.InstructorName = instructorName.String
	courseDetail.LessonCount = lessonCount
	courseDetail.EnrollmentCount = enrollmentCount
	
	return &courseDetail, nil
}

// scanCourse scans a course from database rows
func (r *CourseRepository) scanCourse(rows *sql.Rows) (*models.Course, error) {
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
		return nil, err
	}
	
	return course, nil
}
