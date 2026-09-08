package productimages

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestNewServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	if _, err := NewService(nil); err == nil {
		t.Fatal("NewService(nil) error = nil, want error")
	}
}

func TestAddProductImagesValidatesAndNormalizesInput(t *testing.T) {
	t.Parallel()

	width := 1200
	altText := "Party pig preview"
	etag := " etag-value "
	want := AddProductImagesInput{
		ProductID: "product-id",
		Images: []NewProductImage{{
			R2ObjectKey:      "products/product-id/party-pig.webp",
			OriginalFilename: "party-pig.webp",
			ContentType:      "image/webp",
			FileSizeBytes:    4096,
			WidthPX:          &width,
			AltText:          &altText,
			DisplayOrder:     2,
			IsPrimary:        true,
			R2ETag:           stringPointer("etag-value"),
		}},
	}
	repository := &repositoryStub{
		createMany: func(
			_ context.Context,
			input AddProductImagesInput,
		) ([]ProductImage, error) {
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("CreateMany() input = %#v, want %#v", input, want)
			}
			return []ProductImage{{ID: "image-id", ProductID: input.ProductID}}, nil
		},
	}
	service := newTestService(t, repository)

	result, err := service.AddProductImages(context.Background(), AddProductImagesInput{
		ProductID: " product-id ",
		Images: []NewProductImage{{
			R2ObjectKey:      " products/product-id/party-pig.webp ",
			OriginalFilename: " party-pig.webp ",
			ContentType:      " image/webp ",
			FileSizeBytes:    4096,
			WidthPX:          &width,
			AltText:          &altText,
			DisplayOrder:     2,
			IsPrimary:        true,
			R2ETag:           &etag,
		}},
	})
	if err != nil {
		t.Fatalf("AddProductImages() error = %v", err)
	}
	if len(result) != 1 || result[0].ID != "image-id" {
		t.Fatalf("AddProductImages() = %#v", result)
	}
}

func TestAddProductImagesValidation(t *testing.T) {
	t.Parallel()

	validImage := NewProductImage{
		R2ObjectKey:      "products/product-id/image.webp",
		OriginalFilename: "image.webp",
		ContentType:      "image/webp",
		FileSizeBytes:    1,
	}
	nonPositive := 0
	tests := []struct {
		name  string
		input AddProductImagesInput
	}{
		{name: "missing product", input: AddProductImagesInput{Images: []NewProductImage{validImage}}},
		{name: "missing images", input: AddProductImagesInput{ProductID: "product-id"}},
		{
			name: "missing object key",
			input: AddProductImagesInput{
				ProductID: "product-id",
				Images:    []NewProductImage{{OriginalFilename: "image.webp", ContentType: "image/webp"}},
			},
		},
		{
			name: "non-image content type",
			input: AddProductImagesInput{
				ProductID: "product-id",
				Images: []NewProductImage{{
					R2ObjectKey: "key", OriginalFilename: "image.txt", ContentType: "text/plain",
				}},
			},
		},
		{
			name: "negative file size",
			input: AddProductImagesInput{
				ProductID: "product-id",
				Images: []NewProductImage{{
					R2ObjectKey: "key", OriginalFilename: "image.webp",
					ContentType: "image/webp", FileSizeBytes: -1,
				}},
			},
		},
		{
			name: "invalid dimensions",
			input: AddProductImagesInput{
				ProductID: "product-id",
				Images: []NewProductImage{{
					R2ObjectKey: "key", OriginalFilename: "image.webp",
					ContentType: "image/webp", WidthPX: &nonPositive,
				}},
			},
		},
		{
			name: "duplicate keys",
			input: AddProductImagesInput{
				ProductID: "product-id",
				Images:    []NewProductImage{validImage, validImage},
			},
		},
		{
			name: "multiple primary images",
			input: AddProductImagesInput{
				ProductID: "product-id",
				Images: []NewProductImage{
					withPrimary(validImage, "first"),
					withPrimary(validImage, "second"),
				},
			},
		},
	}

	service := newTestService(t, &repositoryStub{})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.AddProductImages(context.Background(), test.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestAddProductImagesPreservesRepositoryError(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{
		createMany: func(context.Context, AddProductImagesInput) ([]ProductImage, error) {
			return nil, ErrConflict
		},
	})
	_, err := service.AddProductImages(context.Background(), AddProductImagesInput{
		ProductID: "product-id",
		Images: []NewProductImage{{
			R2ObjectKey: "key", OriginalFilename: "image.webp", ContentType: "image/webp",
		}},
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

type repositoryStub struct {
	createMany func(context.Context, AddProductImagesInput) ([]ProductImage, error)
}

func (r *repositoryStub) CreateMany(
	ctx context.Context,
	input AddProductImagesInput,
) ([]ProductImage, error) {
	if r.createMany == nil {
		return []ProductImage{}, nil
	}
	return r.createMany(ctx, input)
}

func newTestService(t *testing.T, repository Repository) *Service {
	t.Helper()
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func stringPointer(value string) *string {
	return &value
}

func withPrimary(image NewProductImage, suffix string) NewProductImage {
	image.R2ObjectKey += "-" + suffix
	image.IsPrimary = true
	return image
}
