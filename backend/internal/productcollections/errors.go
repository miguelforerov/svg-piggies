package productcollections

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid product collection input")
	ErrNotFound          = errors.New("product collection not found")
	ErrReferenceNotFound = errors.New("product or collection not found")
	ErrConflict          = errors.New("product is already assigned to the collection")
)
