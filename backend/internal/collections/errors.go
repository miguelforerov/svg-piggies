package collections

import "errors"

var (
	ErrInvalidInput = errors.New("invalid collection input")
	ErrNotFound     = errors.New("collection not found")
	ErrConflict     = errors.New("collection conflicts with an existing record")
)
