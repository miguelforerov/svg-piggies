package productimages

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid product image input")
	ErrReferenceNotFound = errors.New("product not found")
	ErrConflict          = errors.New("product image conflicts with an existing record")
)
