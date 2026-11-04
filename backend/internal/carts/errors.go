package carts

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid cart input")
	ErrNotFound           = errors.New("cart not found")
	ErrProductUnavailable = errors.New("product is unavailable")
	ErrConflict           = errors.New("cart conflicts with an existing record")
)
