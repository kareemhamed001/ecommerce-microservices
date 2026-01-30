package repository

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user with the given identifier already exists")
)
