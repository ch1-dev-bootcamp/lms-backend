package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// CertificateRepository implements the CertificateRepository interface
type CertificateRepository struct {
	*BaseRepository[models.Certificate]
}

// NewCertificateRepository creates a new certificate repository
func NewCertificateRepository(db *sql.DB) *CertificateRepository {
	return &CertificateRepository{
		BaseRepository: NewBaseRepository[models.Certificate](db, "certificates"),
	}
}

// Create inserts a new certificate into the database
func (r *CertificateRepository) Create(ctx context.Context, certificate *models.Certificate) error {
	query := `
		INSERT INTO certificates (id, user_id, course_id, issued_at, certificate_code)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		certificate.ID,
		certificate.UserID,
		certificate.CourseID,
		certificate.IssuedAt,
		certificate.CertificateCode,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves a certificate by ID
func (r *CertificateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Certificate, error) {
	query := `
		SELECT id, user_id, course_id, issued_at, certificate_code
		FROM certificates
		WHERE id = $1
	`
	
	certificate := &models.Certificate{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&certificate.ID,
		&certificate.UserID,
		&certificate.CourseID,
		&certificate.IssuedAt,
		&certificate.CertificateCode,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return certificate, nil
}

// Update updates an existing certificate
func (r *CertificateRepository) Update(ctx context.Context, certificate *models.Certificate) error {
	query := `
		UPDATE certificates
		SET user_id = $2, course_id = $3, issued_at = $4, certificate_code = $5
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query,
		certificate.ID,
		certificate.UserID,
		certificate.CourseID,
		certificate.IssuedAt,
		certificate.CertificateCode,
	)
	
	return handleDatabaseError(err)
}

// List retrieves certificates with pagination
func (r *CertificateRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.Certificate, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, user_id, course_id, issued_at, certificate_code
		FROM certificates
		ORDER BY issued_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanCertificate)
}

// GetByUser retrieves certificates by user with pagination
func (r *CertificateRepository) GetByUser(ctx context.Context, userID uuid.UUID, pagination models.PaginationRequest) ([]models.Certificate, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, user_id, course_id, issued_at, certificate_code
		FROM certificates
		WHERE user_id = $1
		ORDER BY issued_at DESC
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
	var results []models.Certificate
	for rows.Next() {
		certificate, err := r.scanCertificate(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *certificate)
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

// GetByCourse retrieves certificates by course with pagination
func (r *CertificateRepository) GetByCourse(ctx context.Context, courseID uuid.UUID, pagination models.PaginationRequest) ([]models.Certificate, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, user_id, course_id, issued_at, certificate_code
		FROM certificates
		WHERE course_id = $1
		ORDER BY issued_at DESC
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
	var results []models.Certificate
	for rows.Next() {
		certificate, err := r.scanCertificate(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *certificate)
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

// GetByUserAndCourse retrieves a certificate by user and course
func (r *CertificateRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*models.Certificate, error) {
	query := `
		SELECT id, user_id, course_id, issued_at, certificate_code
		FROM certificates
		WHERE user_id = $1 AND course_id = $2
	`
	
	certificate := &models.Certificate{}
	err := r.db.QueryRowContext(ctx, query, userID, courseID).Scan(
		&certificate.ID,
		&certificate.UserID,
		&certificate.CourseID,
		&certificate.IssuedAt,
		&certificate.CertificateCode,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return certificate, nil
}

// GetByCode retrieves a certificate by its verification code
func (r *CertificateRepository) GetByCode(ctx context.Context, code string) (*models.Certificate, error) {
	query := `
		SELECT id, user_id, course_id, issued_at, certificate_code
		FROM certificates
		WHERE certificate_code = $1
	`
	
	certificate := &models.Certificate{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&certificate.ID,
		&certificate.UserID,
		&certificate.CourseID,
		&certificate.IssuedAt,
		&certificate.CertificateCode,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return certificate, nil
}

// GetWithDetails retrieves certificate with additional details
func (r *CertificateRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*models.CertificateDetailResponse, error) {
	query := `
		SELECT 
			c.id, c.user_id, c.course_id, c.issued_at, c.certificate_code,
			u.name as user_name,
			co.title as course_title,
			'valid' as status
		FROM certificates c
		LEFT JOIN users u ON c.user_id = u.id
		LEFT JOIN courses co ON c.course_id = co.id
		WHERE c.id = $1
	`
	
	var certificateDetail models.CertificateDetailResponse
	var userName, courseTitle, status sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&certificateDetail.ID,
		&certificateDetail.UserID,
		&certificateDetail.CourseID,
		&certificateDetail.IssuedAt,
		&certificateDetail.CertificateCode,
		&userName,
		&courseTitle,
		&status,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	certificateDetail.UserName = userName.String
	certificateDetail.CourseTitle = courseTitle.String
	certificateDetail.Status = status.String
	
	return &certificateDetail, nil
}

// VerifyCertificate verifies a certificate and returns verification details
func (r *CertificateRepository) VerifyCertificate(ctx context.Context, id uuid.UUID) (*models.VerifyCertificateResponse, error) {
	query := `
		SELECT 
			c.id, c.issued_at,
			u.name as user_name,
			co.title as course_title
		FROM certificates c
		LEFT JOIN users u ON c.user_id = u.id
		LEFT JOIN courses co ON c.course_id = co.id
		WHERE c.id = $1
	`
	
	var verifyResponse models.VerifyCertificateResponse
	var userName, courseTitle sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&verifyResponse.CertificateID,
		&verifyResponse.IssuedAt,
		&userName,
		&courseTitle,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	verifyResponse.Valid = true
	verifyResponse.UserName = userName.String
	verifyResponse.CourseTitle = courseTitle.String
	verifyResponse.VerifiedAt = time.Now()
	
	return &verifyResponse, nil
}

// scanCertificate scans a certificate from database rows
func (r *CertificateRepository) scanCertificate(rows *sql.Rows) (*models.Certificate, error) {
	certificate := &models.Certificate{}
	err := rows.Scan(
		&certificate.ID,
		&certificate.UserID,
		&certificate.CourseID,
		&certificate.IssuedAt,
		&certificate.CertificateCode,
	)
	
	if err != nil {
		return nil, err
	}
	
	return certificate, nil
}
