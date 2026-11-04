package customers

import "errors"

var (
	ErrConflict     = errors.New("customer already exists")
	ErrInvalidInput = errors.New("invalid customer input")
	ErrNotFound     = errors.New("customer not found")
)
