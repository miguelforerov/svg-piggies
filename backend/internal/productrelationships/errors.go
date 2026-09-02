package productrelationships

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid product relationship input")
	ErrNotFound          = errors.New("product relationship not found")
	ErrReferenceNotFound = errors.New("product or related product not found")
	ErrConflict          = errors.New("product relationship conflicts with an existing record")
)
