package producttypes

import "errors"

var (
	ErrInvalidInput = errors.New("invalid product type input")
	ErrNotFound     = errors.New("product type not found")
	ErrConflict     = errors.New("product type conflicts with an existing record")
)
