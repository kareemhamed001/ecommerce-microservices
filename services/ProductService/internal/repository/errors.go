package repository

import "errors"

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrDatabaseConnection  = errors.New("database connection error")
	ErrDatabaseQuery       = errors.New("database query failed")
	ErrForeignKeyViolation = errors.New("related record not found")
	ErrInvalidData         = errors.New("invalid data provided")
)
