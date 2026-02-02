package postgresql

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kareemhamed001/e-commerce/services/UserService/internal/repository"
)

// mapPostgresError maps Postgres-specific errors to readable repository errors
func mapPostgresError(err error) error {
	if err == nil {
		return nil
	}

	// Check for Postgres-specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return repository.ErrUserAlreadyExists
		case "23503": // foreign_key_violation
			return repository.ErrForeignKeyViolation
		case "23502": // not_null_violation
			return repository.ErrInvalidData
		case "23514": // check_violation
			return repository.ErrInvalidData
		case "08000", "08003", "08006": // connection errors
			return repository.ErrDatabaseConnection
		default:
			return repository.ErrDatabaseQuery
		}
	}

	// Return the original error if it's not a Postgres error
	return repository.ErrDatabaseQuery
}
