package repository

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user with the given identifier already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrAddressNotFound     = errors.New("address not found")
	ErrDatabaseConnection  = errors.New("database connection error")
	ErrDatabaseQuery       = errors.New("database query failed")
	ErrForeignKeyViolation = errors.New("related record not found")
	ErrInvalidData         = errors.New("invalid data provided")
)
