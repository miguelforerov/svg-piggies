package productimages

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

const uploadProductID = "b2f18e56-dff5-4b7a-ab89-3dc6df2fa7af"

type uploadProductStub struct{ err error }

func (s *uploadProductStub) GetProduct(_ context.Context, id string) (products.Product, error) {
	if s.err != nil {
		return products.Product{}, s.err
	}
	return products.Product{ID: id}, nil
}

type uploadImageStub struct {
	image  ProductImage
	err    error
	input  AddProductImagesInput
	called int
}

func (s *uploadImageStub) AddProductImages(_ context.Context, input AddProductImagesInput) ([]ProductImage, error) {
	s.called++
	s.input = input
	if s.err != nil {
		return nil, s.err
	}
	if s.image.ID == "" {
		s.image = ProductImage{ID: "image-id", ProductID: input.ProductID, R2ObjectKey: input.Images[0].R2ObjectKey,
			ConfirmationFingerprint: input.Images[0].ConfirmationFingerprint}
	}
	return []ProductImage{s.image}, nil
}

type uploadLookupStub struct {
	image       ProductImage
	err         error
	called      int
	missUntil   int
	schemaErr   error
	schemaCalls int
}

func (s *uploadLookupStub) CheckUploadSchema(context.Context) error {
	s.schemaCalls++
	return s.schemaErr
}

func (s *uploadLookupStub) GetByObjectKey(_ context.Context, key string) (ProductImage, error) {
	s.called++
	if s.called <= s.missUntil {
		return ProductImage{}, ErrImageNotFound
	}
	if s.err != nil {
		return ProductImage{}, s.err
	}
	if s.image.ID == "" {
		return ProductImage{}, ErrImageNotFound
	}
	if s.image.R2ObjectKey != key {
		return ProductImage{}, ErrImageNotFound
	}
	return s.image, nil
}

type uploadStorageStub struct {
	bucket                       storage.Bucket
	key, contentType             string
	duration                     time.Duration
	metadata                     storage.ObjectMetadata
	data                         []byte
	presignErr, headErr, readErr error
	readHook                     func()
	headCalled                   int
	readCalled                   int
	presignCalled                int
}

func (s *uploadStorageStub) ReadObject(_ context.Context, bucket storage.Bucket, key, _ string, _ int64) ([]byte, error) {
	s.readCalled++
	s.bucket, s.key = bucket, key
	if s.readHook != nil {
		s.readHook()
	}
	if s.readErr != nil {
		return nil, s.readErr
	}
	return s.data, nil
}

func TestConfirmUploadStopsAfterContextCancellation(t *testing.T) {
	s, _, images, _, store := newUploadTestService(t)
	a := requestTestUpload(t, s)
	store.metadata = storage.ObjectMetadata{ContentLength: int64(len(store.data)), ContentType: "image/webp"}
	ctx, cancel := context.WithCancel(context.Background())
	store.readHook = cancel
	if _, err := s.ConfirmUpload(ctx, ConfirmUploadInput{ProductID: uploadProductID, UploadToken: a.UploadToken}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled confirmation = %v", err)
	}
	if images.called != 0 {
		t.Fatal("canceled confirmation registered an image")
	}
}

func (s *uploadStorageStub) PresignPut(_ context.Context, bucket storage.Bucket, key, contentType string, duration time.Duration) (storage.PresignedRequest, error) {
	s.presignCalled++
	s.bucket, s.key, s.contentType, s.duration = bucket, key, contentType, duration
	if s.presignErr != nil {
		return storage.PresignedRequest{}, s.presignErr
	}
	return storage.PresignedRequest{URL: "https://example.test/upload", Headers: http.Header{"Content-Type": {contentType}, "If-None-Match": {"*"}}}, nil
}

func TestRequestUploadRequiresConfirmationSchemaBeforePresigning(t *testing.T) {
	s, _, _, lookup, store := newUploadTestService(t)
	lookup.schemaErr = &pgconn.PgError{Code: "42703", Message: "column confirmation_fingerprint does not exist"}
	input := RequestUploadInput{ProductID: uploadProductID, OriginalFilename: "party.webp", ContentType: "image/webp", FileSizeBytes: 442}
	if authorization, err := s.RequestUpload(context.Background(), input); !errors.Is(err, ErrUploadUnavailable) || authorization.UploadURL != "" || store.presignCalled != 0 {
		t.Fatalf("missing schema: authorization=%+v, err=%v, presigns=%d", authorization, err, store.presignCalled)
	}
	lookup.schemaErr = nil
	if _, err := s.RequestUpload(context.Background(), input); err != nil || store.presignCalled != 1 || lookup.schemaCalls != 2 {
		t.Fatalf("schema recovered: err=%v, presigns=%d, checks=%d", err, store.presignCalled, lookup.schemaCalls)
	}
	if _, err := s.RequestUpload(context.Background(), input); err != nil || store.presignCalled != 2 || lookup.schemaCalls != 2 {
		t.Fatalf("successful schema check not cached: err=%v, presigns=%d, checks=%d", err, store.presignCalled, lookup.schemaCalls)
	}
}
func (s *uploadStorageStub) HeadObject(_ context.Context, bucket storage.Bucket, key string) (storage.ObjectMetadata, error) {
	s.headCalled++
	s.bucket, s.key = bucket, key
	if s.headErr != nil {
		return storage.ObjectMetadata{}, s.headErr
	}
	return s.metadata, nil
}

func newUploadTestService(t *testing.T) (*UploadService, *uploadProductStub, *uploadImageStub, *uploadLookupStub, *uploadStorageStub) {
	t.Helper()
	p, i, l, r := &uploadProductStub{}, &uploadImageStub{}, &uploadLookupStub{}, &uploadStorageStub{}
	r.data, _ = os.ReadFile("testdata/valid.webp")
	if len(r.data) == 0 {
		t.Fatal("missing valid WebP test fixture")
	}
	s, err := NewUploadService(p, i, l, r, "test-signing-secret")
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC) }
	return s, p, i, l, r
}

func requestTestUpload(t *testing.T, s *UploadService) UploadAuthorization {
	t.Helper()
	a, err := s.RequestUpload(context.Background(), RequestUploadInput{ProductID: uploadProductID, OriginalFilename: "party.webp", ContentType: "image/webp", FileSizeBytes: int64(len(s.storage.(*uploadStorageStub).data))})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestRequestUploadCreatesTrustedKeyAndToken(t *testing.T) {
	s, _, _, _, store := newUploadTestService(t)
	a := requestTestUpload(t, s)
	if store.bucket != storage.AssetsBucket || store.key != a.ObjectKey || store.contentType != "image/webp" || store.duration != UploadTTL {
		t.Fatalf("presign call = %+v", store)
	}
	if !strings.HasPrefix(a.ObjectKey, "products/"+uploadProductID+"/images/") || !strings.HasSuffix(a.ObjectKey, ".webp") {
		t.Fatalf("key = %q", a.ObjectKey)
	}
	if a.UploadURL == "" || a.Headers.Get("Content-Type") != "image/webp" || a.Headers.Get("If-None-Match") != "*" || a.UploadToken == "" {
		t.Fatalf("authorization = %+v", a)
	}
	if b := requestTestUpload(t, s); b.ObjectKey == a.ObjectKey {
		t.Fatal("two uploads received the same key")
	}
	if strings.Contains(a.UploadToken, "test-signing-secret") {
		t.Fatal("token exposes signing secret")
	}
}

func TestRequestUploadRejectsInvalidInputAndMissingProduct(t *testing.T) {
	s, product, _, _, store := newUploadTestService(t)
	base := RequestUploadInput{ProductID: uploadProductID, OriginalFilename: "party.webp", ContentType: "image/webp", FileSizeBytes: 100}
	tests := []struct {
		name   string
		change func(*RequestUploadInput)
	}{
		{"filename path", func(i *RequestUploadInput) { i.OriginalFilename = "../party.webp" }},
		{"filename extension", func(i *RequestUploadInput) { i.OriginalFilename = "party.png" }},
		{"svg", func(i *RequestUploadInput) { i.ContentType = "image/svg+xml" }},
		{"avif", func(i *RequestUploadInput) { i.ContentType = "image/avif"; i.OriginalFilename = "party.avif" }},
		{"zero size", func(i *RequestUploadInput) { i.FileSizeBytes = 0 }},
		{"too large", func(i *RequestUploadInput) { i.FileSizeBytes = MaxImageBytes + 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := base
			tt.change(&input)
			if _, err := s.RequestUpload(context.Background(), input); !errors.Is(err, ErrInvalidUpload) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	product.err = products.ErrNotFound
	if _, err := s.RequestUpload(context.Background(), base); !errors.Is(err, ErrReferenceNotFound) {
		t.Fatalf("missing product = %v", err)
	}
	product.err = nil
	store.presignErr = errors.New("R2 unavailable")
	if _, err := s.RequestUpload(context.Background(), base); !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("R2 error = %v", err)
	}
}

func TestConfirmUploadUsesVerifiedObjectMetadata(t *testing.T) {
	s, _, images, _, store := newUploadTestService(t)
	a := requestTestUpload(t, s)
	store.metadata = storage.ObjectMetadata{ContentLength: int64(len(store.data)), ContentType: "image/webp", ETag: "etag"}
	alt := "Party pig"
	result, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: uploadProductID, UploadToken: a.UploadToken, AltText: &alt, DisplayOrder: 2, IsPrimary: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "image-id" || store.bucket != storage.AssetsBucket || store.key != a.ObjectKey {
		t.Fatalf("result = %+v; storage = %+v", result, store)
	}
	if images.called != 1 || len(images.input.Images) != 1 {
		t.Fatalf("registration = %+v", images.input)
	}
	img := images.input.Images[0]
	if img.R2ObjectKey != a.ObjectKey || img.FileSizeBytes != int64(len(store.data)) || img.ContentType != "image/webp" || img.R2ETag == nil || *img.R2ETag != "etag" || img.WidthPX == nil || *img.WidthPX <= 0 || img.HeightPX == nil || *img.HeightPX <= 0 || !img.IsPrimary || img.DisplayOrder != 2 {
		t.Fatalf("registered image = %+v", img)
	}
}

func TestConfirmUploadRejectsInvalidOrMismatchedObjects(t *testing.T) {
	s, _, images, _, store := newUploadTestService(t)
	a := requestTestUpload(t, s)
	input := ConfirmUploadInput{ProductID: uploadProductID, UploadToken: a.UploadToken}
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: uploadProductID, UploadToken: a.UploadToken + "x"}); !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("tampered token = %v", err)
	}
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: "6f7310e0-3d7c-481a-bffc-cf8d19377e70", UploadToken: a.UploadToken}); !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("wrong product = %v", err)
	}
	store.headErr = errors.New("R2 unavailable")
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("R2 error = %v", err)
	}
	store.headErr = &smithyhttp.ResponseError{Response: &smithyhttp.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, Err: errors.New("not found")}
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("missing object = %v", err)
	}
	store.headErr = nil
	store.metadata = storage.ObjectMetadata{ContentLength: 1, ContentType: "image/webp"}
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrObjectMismatch) {
		t.Fatalf("size mismatch = %v", err)
	}
	store.metadata = storage.ObjectMetadata{ContentLength: MaxImageBytes + 1, ContentType: "image/webp"}
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrObjectMismatch) || store.readCalled != 0 {
		t.Fatalf("oversized object was read or accepted: %v, reads=%d", err, store.readCalled)
	}
	store.metadata = storage.ObjectMetadata{ContentLength: int64(len(store.data)), ContentType: "image/png"}
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrObjectMismatch) {
		t.Fatalf("MIME mismatch = %v", err)
	}
	if images.called != 0 {
		t.Fatal("invalid object was registered")
	}
	store.metadata = storage.ObjectMetadata{ContentLength: int64(len(store.data)), ContentType: "image/webp"}
	store.data = make([]byte, len(store.data))
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("invalid image bytes = %v", err)
	}
	if images.called != 0 {
		t.Fatal("invalid image bytes were registered")
	}
	s.now = func() time.Time { return a.ExpiresAt.Add(time.Second) }
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrExpiredUpload) {
		t.Fatalf("expired token = %v", err)
	}
}

func TestConfirmUploadRepeatedAndDatabaseErrors(t *testing.T) {
	s, _, images, lookup, store := newUploadTestService(t)
	a := requestTestUpload(t, s)
	input := ConfirmUploadInput{ProductID: uploadProductID, UploadToken: a.UploadToken}
	fingerprint := confirmationFingerprint(input)
	lookup.image = ProductImage{ID: "existing", ProductID: uploadProductID, R2ObjectKey: a.ObjectKey,
		ConfirmationFingerprint: &fingerprint}
	result, err := s.ConfirmUpload(context.Background(), input)
	if err != nil || result.ID != "existing" || store.headCalled != 0 || images.called != 0 {
		t.Fatalf("repeat = %+v, %v", result, err)
	}
	s.now = func() time.Time { return a.ExpiresAt.Add(time.Hour) }
	if result, err := s.ConfirmUpload(context.Background(), input); err != nil || result.ID != "existing" {
		t.Fatalf("repeat after expiry = %+v, %v", result, err)
	}
	for _, different := range []ConfirmUploadInput{
		{ProductID: uploadProductID, UploadToken: a.UploadToken, AltText: stringPointer("changed")},
		{ProductID: uploadProductID, UploadToken: a.UploadToken, DisplayOrder: 1},
		{ProductID: uploadProductID, UploadToken: a.UploadToken, IsPrimary: true},
	} {
		if _, err := s.ConfirmUpload(context.Background(), different); !errors.Is(err, ErrConflict) {
			t.Fatalf("conflicting repeat = %v", err)
		}
	}
	lookup.image = ProductImage{}
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrExpiredUpload) {
		t.Fatalf("new image after expiry = %v", err)
	}
	s.now = func() time.Time { return a.ExpiresAt.Add(-time.Minute) }
	lookup.image = ProductImage{ID: "existing", ProductID: uploadProductID, R2ObjectKey: a.ObjectKey,
		ConfirmationFingerprint: &fingerprint}
	lookup.image.ProductID = "6f7310e0-3d7c-481a-bffc-cf8d19377e70"
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, ErrConflict) {
		t.Fatalf("existing key owned by another product = %v", err)
	}
	lookup.image = ProductImage{}
	lookup.err = errors.New("database unavailable")
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, lookup.err) {
		t.Fatalf("lookup error = %v", err)
	}
	lookup.err = nil
	store.metadata = storage.ObjectMetadata{ContentLength: int64(len(store.data)), ContentType: "image/webp"}
	images.err = errors.New("database unavailable")
	if _, err := s.ConfirmUpload(context.Background(), input); !errors.Is(err, images.err) {
		t.Fatalf("insert error = %v", err)
	}
	images.err = ErrConflict
	lookup.image = ProductImage{ID: "existing", ProductID: uploadProductID, R2ObjectKey: a.ObjectKey,
		ConfirmationFingerprint: &fingerprint}
	lookup.called = 0
	lookup.missUntil = 1
	if image, err := s.ConfirmUpload(context.Background(), input); err != nil || image.ID != "existing" {
		t.Fatalf("concurrent confirmation = %v", err)
	}
}
