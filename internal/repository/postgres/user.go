package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/lms-backend/internal/models"
)

// UserRepository implements the UserRepository interface
type UserRepository struct {
	*BaseRepository[models.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository[models.User](db, "users"),
	}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	
	return handleDatabaseError(err)
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $2, name = $3, role = $4, updated_at = $5
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Role,
		user.UpdatedAt,
	)
	
	return handleDatabaseError(err)
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, pagination models.PaginationRequest) ([]models.User, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	return executePaginationQuery(ctx, r.db, baseQuery, countQuery, pagination, r.scanUser)
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, handleDatabaseError(err)
	}
	
	return user, nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query, userID, passwordHash)
	return handleDatabaseError(err)
}

// GetByRole retrieves users by role with pagination
func (r *UserRepository) GetByRole(ctx context.Context, role string, pagination models.PaginationRequest) ([]models.User, *models.PaginationResponse, error) {
	baseQuery := `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC
	`
	
	countQuery := buildCountQuery(baseQuery)
	
	// Execute count query with role parameter
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, role).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	query, _ := buildPaginationQuery(baseQuery, pagination)
	
	// Execute query with role parameter
	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []models.User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *user)
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

// scanUser scans a user from database rows
func (r *UserRepository) scanUser(rows *sql.Rows) (*models.User, error) {
	user := &models.User{}
	err := rows.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}
