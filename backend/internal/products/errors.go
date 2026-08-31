package products

import "errors"

var (
	ErrInvalidInput = errors.New("invalid product input")
	ErrNotFound     = errors.New("product not found")
	ErrConflict     = errors.New("product conflicts with an existing record")
)
