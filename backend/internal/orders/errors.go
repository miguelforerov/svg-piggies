package orders

import "errors"

var (
	ErrConflict           = errors.New("order conflicts with existing data")
	ErrEmptyCart          = errors.New("cart is empty")
	ErrInvalidInput       = errors.New("invalid order input")
	ErrNotFound           = errors.New("order or cart not found")
	ErrProductUnavailable = errors.New("cart contains an unavailable product")
)
