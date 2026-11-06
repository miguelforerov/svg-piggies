package productimages

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

const (
	MaxImageBytes = 10 << 20
	UploadTTL     = 10 * time.Minute
	maxTokenSize  = 4096
)

var (
	ErrInvalidUpload     = errors.New("invalid image upload")
	ErrExpiredUpload     = errors.New("image upload expired")
	ErrObjectNotFound    = errors.New("uploaded image not found")
	ErrObjectMismatch    = errors.New("uploaded image metadata does not match authorization")
	ErrInvalidImage      = errors.New("uploaded object is not a valid image")
	ErrStorageFailure    = errors.New("image storage operation failed")
	ErrImageNotFound     = errors.New("product image not found")
	ErrUploadUnavailable = errors.New("image uploads are temporarily unavailable")
)

type ProductLookup interface {
	GetProduct(context.Context, string) (products.Product, error)
}

type ImageRegistrar interface {
	AddProductImages(context.Context, AddProductImagesInput) ([]ProductImage, error)
}

type ImageLookup interface {
	GetByObjectKey(context.Context, string) (ProductImage, error)
	CheckUploadSchema(context.Context) error
}

type UploadStorage interface {
	PresignPut(context.Context, storage.Bucket, string, string, time.Duration) (storage.PresignedRequest, error)
	HeadObject(context.Context, storage.Bucket, string) (storage.ObjectMetadata, error)
	ReadObject(context.Context, storage.Bucket, string, string, int64) ([]byte, error)
}

type UploadService struct {
	products        ProductLookup
	images          ImageRegistrar
	lookup          ImageLookup
	storage         UploadStorage
	secret          [32]byte
	now             func() time.Time
	validationSlot  chan struct{}
	schemaCheckSlot chan struct{}
	schemaReady     atomic.Bool
}

func NewUploadService(products ProductLookup, images ImageRegistrar, lookup ImageLookup, store UploadStorage, signingSecret string) (*UploadService, error) {
	if products == nil || images == nil || lookup == nil || store == nil || signingSecret == "" {
		return nil, errors.New("product lookup, image registrar, image lookup, storage, and signing secret are required")
	}
	// Domain separation keeps upload tokens independent from R2 request signing.
	secret := sha256.Sum256([]byte("svg-piggies:image-upload-token:v1:\x00" + signingSecret))
	return &UploadService{products: products, images: images, lookup: lookup, storage: store,
		secret: secret, now: time.Now, validationSlot: make(chan struct{}, 1), schemaCheckSlot: make(chan struct{}, 1)}, nil
}

type RequestUploadInput struct {
	ProductID        string
	OriginalFilename string
	ContentType      string
	FileSizeBytes    int64
}

type UploadAuthorization struct {
	UploadToken string
	ObjectKey   string
	UploadURL   string
	Headers     http.Header
	ExpiresAt   time.Time
}

type ConfirmUploadInput struct {
	ProductID    string
	UploadToken  string
	AltText      *string
	DisplayOrder int
	IsPrimary    bool
}

type uploadClaims struct {
	Version          int    `json:"v"`
	ProductID        string `json:"p"`
	ObjectKey        string `json:"k"`
	OriginalFilename string `json:"n"`
	ContentType      string `json:"c"`
	FileSizeBytes    int64  `json:"s"`
	ExpiresAt        int64  `json:"e"`
}

func (s *UploadService) RequestUpload(ctx context.Context, input RequestUploadInput) (UploadAuthorization, error) {
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return UploadAuthorization{}, fmt.Errorf("%w: invalid product id", ErrInvalidUpload)
	}
	extension, err := storage.AssetExtension(input.ContentType)
	if err != nil {
		return UploadAuthorization{}, fmt.Errorf("%w: unsupported content type", ErrInvalidUpload)
	}
	if err := validateFilename(input.OriginalFilename, extension); err != nil {
		return UploadAuthorization{}, err
	}
	if input.FileSizeBytes <= 0 || input.FileSizeBytes > MaxImageBytes {
		return UploadAuthorization{}, fmt.Errorf("%w: image size must be between 1 and %d bytes", ErrInvalidUpload, MaxImageBytes)
	}
	if _, err := s.products.GetProduct(ctx, productID.String()); err != nil {
		if errors.Is(err, products.ErrNotFound) {
			return UploadAuthorization{}, ErrReferenceNotFound
		}
		return UploadAuthorization{}, fmt.Errorf("find product for image upload: %w", err)
	}
	if err := s.ensureUploadSchema(ctx); err != nil {
		return UploadAuthorization{}, err
	}
	uploadID, err := uuid.NewRandom()
	if err != nil {
		return UploadAuthorization{}, fmt.Errorf("create image upload id: %w", err)
	}
	key := fmt.Sprintf("products/%s/images/%s%s", productID, uploadID, extension)
	expiresAt := s.now().UTC().Add(UploadTTL)
	claims := uploadClaims{1, productID.String(), key, input.OriginalFilename, input.ContentType, input.FileSizeBytes, expiresAt.Unix()}
	token, err := s.sign(claims)
	if err != nil {
		return UploadAuthorization{}, err
	}
	request, err := s.storage.PresignPut(ctx, storage.AssetsBucket, key, input.ContentType, UploadTTL)
	if err != nil {
		return UploadAuthorization{}, fmt.Errorf("%w: presign image upload: %w", ErrStorageFailure, err)
	}
	return UploadAuthorization{UploadToken: token, ObjectKey: key, UploadURL: request.URL, Headers: request.Headers, ExpiresAt: expiresAt}, nil
}

// Cache only successful checks. A missing migration or transient database
// failure blocks presigning, but the next request can retry after recovery.
func (s *UploadService) ensureUploadSchema(ctx context.Context) error {
	if s.schemaReady.Load() {
		return nil
	}
	select {
	case s.schemaCheckSlot <- struct{}{}:
		defer func() { <-s.schemaCheckSlot }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if s.schemaReady.Load() {
		return nil
	}
	if err := s.lookup.CheckUploadSchema(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrUploadUnavailable, err)
	}
	s.schemaReady.Store(true)
	return nil
}

func (s *UploadService) ConfirmUpload(ctx context.Context, input ConfirmUploadInput) (ProductImage, error) {
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return ProductImage{}, fmt.Errorf("%w: invalid product id", ErrInvalidUpload)
	}
	claims, err := s.verify(input.UploadToken)
	if err != nil {
		return ProductImage{}, err
	}
	if claims.ProductID != productID.String() {
		return ProductImage{}, ErrInvalidUpload
	}
	if input.DisplayOrder < 0 {
		return ProductImage{}, fmt.Errorf("%w: display order cannot be negative", ErrInvalidUpload)
	}
	// Decoding can allocate several pixel buffers. Keep at most one validation
	// in flight per process, including the database registration that follows it.
	select {
	case s.validationSlot <- struct{}{}:
		defer func() { <-s.validationSlot }()
	case <-ctx.Done():
		return ProductImage{}, ctx.Err()
	}
	if image, err := s.lookup.GetByObjectKey(ctx, claims.ObjectKey); err == nil {
		if image.ProductID != claims.ProductID {
			return ProductImage{}, ErrConflict
		}
		return confirmedImage(image, input)
	} else if !errors.Is(err, ErrImageNotFound) {
		return ProductImage{}, fmt.Errorf("find confirmed image: %w", err)
	}
	if !s.now().Before(time.Unix(claims.ExpiresAt, 0)) {
		return ProductImage{}, ErrExpiredUpload
	}

	metadata, err := s.storage.HeadObject(ctx, storage.AssetsBucket, claims.ObjectKey)
	if err != nil {
		var responseError interface{ HTTPStatusCode() int }
		if errors.As(err, &responseError) && responseError.HTTPStatusCode() == http.StatusNotFound {
			return ProductImage{}, ErrObjectNotFound
		}
		return ProductImage{}, fmt.Errorf("%w: head uploaded image: %w", ErrStorageFailure, err)
	}
	if metadata.ContentLength != claims.FileSizeBytes || metadata.ContentLength > MaxImageBytes || metadata.ContentType != claims.ContentType {
		return ProductImage{}, ErrObjectMismatch
	}
	data, err := s.storage.ReadObject(ctx, storage.AssetsBucket, claims.ObjectKey, metadata.ETag, MaxImageBytes)
	if err != nil {
		return ProductImage{}, fmt.Errorf("%w: read uploaded image: %w", ErrStorageFailure, err)
	}
	if int64(len(data)) != metadata.ContentLength {
		return ProductImage{}, ErrObjectMismatch
	}
	width, height, err := validateImageBytes(data, claims.ContentType)
	if err != nil {
		return ProductImage{}, fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}
	if err := ctx.Err(); err != nil {
		return ProductImage{}, err
	}
	if !s.now().Before(time.Unix(claims.ExpiresAt, 0)) {
		return ProductImage{}, ErrExpiredUpload
	}
	var etag *string
	if metadata.ETag != "" {
		etag = &metadata.ETag
	}
	fingerprint := confirmationFingerprint(input)
	images, err := s.images.AddProductImages(ctx, AddProductImagesInput{ProductID: claims.ProductID, Images: []NewProductImage{{
		R2ObjectKey: claims.ObjectKey, OriginalFilename: claims.OriginalFilename,
		ContentType: claims.ContentType, FileSizeBytes: metadata.ContentLength,
		WidthPX: &width, HeightPX: &height,
		AltText: input.AltText, DisplayOrder: input.DisplayOrder, IsPrimary: input.IsPrimary, R2ETag: etag,
		ConfirmationFingerprint: &fingerprint,
	}}})
	if errors.Is(err, ErrConflict) {
		// The unique object key makes concurrent confirmations safe.
		image, lookupErr := s.lookup.GetByObjectKey(ctx, claims.ObjectKey)
		if lookupErr == nil && image.ProductID == claims.ProductID {
			return confirmedImage(image, input)
		}
	}
	if err != nil {
		return ProductImage{}, fmt.Errorf("register uploaded image: %w", err)
	}
	if len(images) != 1 {
		return ProductImage{}, errors.New("register uploaded image: unexpected result count")
	}
	return images[0], nil
}

func confirmedImage(image ProductImage, input ConfirmUploadInput) (ProductImage, error) {
	if image.ConfirmationFingerprint == nil || *image.ConfirmationFingerprint != confirmationFingerprint(input) {
		return ProductImage{}, ErrConflict
	}
	return image, nil
}

func confirmationFingerprint(input ConfirmUploadInput) string {
	payload, _ := json.Marshal(struct {
		AltText      *string `json:"altText"`
		DisplayOrder int     `json:"displayOrder"`
		IsPrimary    bool    `json:"isPrimary"`
	}{input.AltText, input.DisplayOrder, input.IsPrimary})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func validateFilename(filename, canonicalExtension string) error {
	if filename == "" || len(filename) > 255 || strings.TrimSpace(filename) != filename || filename == "." || filename == ".." ||
		strings.ContainsAny(filename, "/\\") || strings.IndexFunc(filename, unicode.IsControl) >= 0 {
		return fmt.Errorf("%w: invalid original filename", ErrInvalidUpload)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != canonicalExtension && !(canonicalExtension == ".jpg" && ext == ".jpeg") {
		return fmt.Errorf("%w: filename extension does not match content type", ErrInvalidUpload)
	}
	return nil
}

func (s *UploadService) sign(claims uploadClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("encode image upload token: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret[:])
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (s *UploadService) verify(token string) (uploadClaims, error) {
	if len(token) == 0 || len(token) > maxTokenSize {
		return uploadClaims{}, ErrInvalidUpload
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return uploadClaims{}, ErrInvalidUpload
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(signature) != sha256.Size {
		return uploadClaims{}, ErrInvalidUpload
	}
	mac := hmac.New(sha256.New, s.secret[:])
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return uploadClaims{}, ErrInvalidUpload
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return uploadClaims{}, ErrInvalidUpload
	}
	var claims uploadClaims
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Version != 1 {
		return uploadClaims{}, ErrInvalidUpload
	}
	return claims, nil
}
