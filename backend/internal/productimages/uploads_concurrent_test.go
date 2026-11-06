package productimages

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

// This double models the UNIQUE object-key constraint; PostgreSQL concurrency
// itself still needs an integration test against a test database.
type concurrentImageRepository struct {
	mu      sync.Mutex
	image   *ProductImage
	lookups int
	gate    chan struct{}
}

func (*concurrentImageRepository) CheckUploadSchema(context.Context) error { return nil }

func (r *concurrentImageRepository) GetByObjectKey(_ context.Context, key string) (ProductImage, error) {
	r.mu.Lock()
	if r.image != nil {
		image := *r.image
		r.mu.Unlock()
		return image, nil
	}
	r.lookups++
	if r.lookups == 2 {
		close(r.gate)
	}
	r.mu.Unlock()
	<-r.gate
	return ProductImage{}, ErrImageNotFound
}

func (r *concurrentImageRepository) AddProductImages(_ context.Context, input AddProductImagesInput) ([]ProductImage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.image != nil {
		return nil, ErrConflict
	}
	newImage := input.Images[0]
	image := ProductImage{ID: "committed", ProductID: input.ProductID,
		R2ObjectKey: newImage.R2ObjectKey, AltText: newImage.AltText,
		DisplayOrder: newImage.DisplayOrder, IsPrimary: newImage.IsPrimary,
		ConfirmationFingerprint: newImage.ConfirmationFingerprint}
	r.image = &image
	return []ProductImage{image}, nil
}

type concurrentUploadStorage struct{ data []byte }

func (s concurrentUploadStorage) PresignPut(context.Context, storage.Bucket, string, string, time.Duration) (storage.PresignedRequest, error) {
	return storage.PresignedRequest{URL: "https://example.test/upload", Headers: http.Header{}}, nil
}
func (s concurrentUploadStorage) HeadObject(context.Context, storage.Bucket, string) (storage.ObjectMetadata, error) {
	return storage.ObjectMetadata{ContentLength: int64(len(s.data)), ContentType: "image/webp", ETag: "etag"}, nil
}
func (s concurrentUploadStorage) ReadObject(context.Context, storage.Bucket, string, string, int64) ([]byte, error) {
	return s.data, nil
}

func TestConcurrentConfirmations(t *testing.T) {
	data, err := os.ReadFile("testdata/valid.webp")
	if err != nil {
		t.Fatal(err)
	}
	for _, different := range []bool{false, true} {
		name := "same metadata"
		if different {
			name = "different metadata"
		}
		t.Run(name, func(t *testing.T) {
			repo := &concurrentImageRepository{gate: make(chan struct{})}
			service, err := NewUploadService(&uploadProductStub{}, repo, repo, concurrentUploadStorage{data}, "test-secret")
			if err != nil {
				t.Fatal(err)
			}
			otherService, err := NewUploadService(&uploadProductStub{}, repo, repo, concurrentUploadStorage{data}, "test-secret")
			if err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
			service.now = func() time.Time { return now }
			otherService.now = func() time.Time { return now }
			token, err := service.sign(uploadClaims{Version: 1, ProductID: uploadProductID,
				ObjectKey:   "products/" + uploadProductID + "/images/a.webp",
				ContentType: "image/webp", FileSizeBytes: int64(len(data)), ExpiresAt: now.Add(time.Minute).Unix()})
			if err != nil {
				t.Fatal(err)
			}
			inputs := []ConfirmUploadInput{{ProductID: uploadProductID, UploadToken: token},
				{ProductID: uploadProductID, UploadToken: token}}
			if different {
				inputs[1].AltText = stringPointer("different")
			}
			var wg sync.WaitGroup
			results := make([]ProductImage, 2)
			errorsFound := make([]error, 2)
			for index := range inputs {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					instance := service
					if index == 1 {
						instance = otherService
					}
					results[index], errorsFound[index] = instance.ConfirmUpload(context.Background(), inputs[index])
				}(index)
			}
			wg.Wait()
			if repo.image == nil {
				t.Fatal("no image committed")
			}
			if !different && (errorsFound[0] != nil || errorsFound[1] != nil ||
				results[0].ID != "committed" || results[1].ID != "committed") {
				t.Fatalf("same metadata results = %+v, errors = %v", results, errorsFound)
			}
			if different {
				successes, conflicts := 0, 0
				for _, err := range errorsFound {
					if err == nil {
						successes++
					} else if errors.Is(err, ErrConflict) {
						conflicts++
					}
				}
				if successes != 1 || conflicts != 1 {
					t.Fatalf("different metadata errors = %v", errorsFound)
				}
			}
		})
	}
}
