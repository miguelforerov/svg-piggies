package productproducttypes

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid product product type input")
	ErrNotFound          = errors.New("product product type not found")
	ErrReferenceNotFound = errors.New("product or product type not found")
	ErrConflict          = errors.New("product already has this product type")
)
