package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/your-org/lms-backend/internal/errors"
	"github.com/your-org/lms-backend/internal/models"
)

// BaseRepository provides common database operations and helper functions.
// Note: Most CRUD operations (Create, GetByID, Update, List) are implemented
// in individual repository files, not in this base repository.
type BaseRepository[T any] struct {
	db    *sql.DB
	table string
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](db *sql.DB, table string) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:    db,
		table: table,
	}
}

// Delete removes a record by ID
// This is a generic implementation that works for all tables with an 'id' column
func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.table)
	_, err := r.db.ExecContext(ctx, query, id)
	return handleDatabaseError(err)
}

// Helper functions for common database operations

// buildPaginationQuery builds a pagination query with LIMIT and OFFSET
func buildPaginationQuery(baseQuery string, pagination models.PaginationRequest) (string, []interface{}) {
	offset := pagination.GetOffset()
	query := fmt.Sprintf("%s LIMIT %d OFFSET %d", baseQuery, pagination.PageSize, offset)
	return query, []interface{}{}
}

// buildCountQuery builds a count query for pagination
func buildCountQuery(baseQuery string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_query", baseQuery)
}

// executePaginationQuery executes a query with pagination and returns results with pagination metadata
func executePaginationQuery[T any](
	ctx context.Context,
	db *sql.DB,
	baseQuery string,
	countQuery string,
	pagination models.PaginationRequest,
	scanFunc func(*sql.Rows) (*T, error),
) ([]T, *models.PaginationResponse, error) {
	// Get total count
	var total int
	err := db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build pagination query
	query, _ := buildPaginationQuery(baseQuery, pagination)

	// Execute query
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []T
	for rows.Next() {
		entity, err := scanFunc(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, *entity)
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

// handleDatabaseError handles common database errors and converts them to appropriate application errors
func handleDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	// Handle PostgreSQL specific errors
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // unique_violation
			return errors.NewDuplicateEntryError(pqErr.Detail)
		case "23503": // foreign_key_violation
			return errors.NewForeignKeyViolationError(pqErr.Detail)
		case "23502": // not_null_violation
			return errors.NewConstraintViolationError(fmt.Sprintf("required field is missing: %s", pqErr.Column))
		case "23514": // check_violation
			return errors.NewConstraintViolationError(pqErr.Detail)
		default:
			return errors.NewDatabaseError(err, "database operation")
		}
	}

	// Handle common SQL errors
	switch err {
	case sql.ErrNoRows:
		return errors.NewNotFoundError("record not found")
	case sql.ErrConnDone:
		return errors.NewServiceUnavailableError("database")
	default:
		return errors.NewDatabaseError(err, "database operation")
	}
}

