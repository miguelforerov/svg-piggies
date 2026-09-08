package productimages

import "time"

type ProductImage struct {
	ID               string
	ProductID        string
	R2ObjectKey      string
	OriginalFilename string
	ContentType      string
	FileSizeBytes    int64
	WidthPX          *int
	HeightPX         *int
	AltText          *string
	DisplayOrder     int
	IsPrimary        bool
	R2ETag           *string `db:"r2_etag"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NewProductImage struct {
	R2ObjectKey      string
	OriginalFilename string
	ContentType      string
	FileSizeBytes    int64
	WidthPX          *int
	HeightPX         *int
	AltText          *string
	DisplayOrder     int
	IsPrimary        bool
	R2ETag           *string
}

type AddProductImagesInput struct {
	ProductID string
	Images    []NewProductImage
}
