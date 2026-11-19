package productfiles

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

type ProductLookup interface {
	GetProduct(context.Context, string) (products.Product, error)
}
type Repository interface {
	List(context.Context, string) ([]ProductFile, error)
	GetConsumptionByDigest(context.Context, string) (Consumption, error)
	GetConsumptionByKey(context.Context, string) (Consumption, error)
	GetByIDAndProduct(context.Context, string, string) (ProductFile, error)
	CreateConfirmed(context.Context, ProductFile, Consumption) (ProductFile, error)
	UpdateMetadata(context.Context, UpdateMetadataInput) (ProductFile, error)
	DeleteByIDAndProduct(context.Context, string, string) error
	Reorder(context.Context, string, []string) error
}
type Storage interface {
	PresignPut(context.Context, storage.Bucket, string, string, time.Duration) (storage.PresignedRequest, error)
	HeadObject(context.Context, storage.Bucket, string) (storage.ObjectMetadata, error)
	ReadObject(context.Context, storage.Bucket, string, string, int64) ([]byte, error)
	DeleteObject(context.Context, storage.Bucket, string) error
}

type Service struct {
	products      ProductLookup
	repo          Repository
	storage       Storage
	keys          map[int]KeyPair
	activeVersion int
	now           func() time.Time
}

func NewService(products ProductLookup, repo Repository, store Storage, keys map[int]KeyPair, activeVersion int) (*Service, error) {
	if products == nil || repo == nil || store == nil {
		return nil, errors.New("product lookup, file repository, and storage are required")
	}
	if formatConfigurationError != nil {
		return nil, fmt.Errorf("load product file formats: %w", formatConfigurationError)
	}
	if len(keys) > 0 {
		if err := validateKeys(keys, activeVersion); err != nil {
			return nil, err
		}
	}
	copyKeys := make(map[int]KeyPair, len(keys))
	for version, key := range keys {
		copyKeys[version] = KeyPair{Signing: append([]byte(nil), key.Signing...), Fingerprint: append([]byte(nil), key.Fingerprint...)}
	}
	return &Service{products: products, repo: repo, storage: store, keys: copyKeys, activeVersion: activeVersion, now: time.Now}, nil
}

func validateFile(name, format, contentType string, size int64) error {
	definition, supported := formats[format]
	validContentType := false
	if supported {
		for _, allowed := range definition.ContentTypes {
			if contentType == allowed {
				validContentType = true
				break
			}
		}
	}
	if len(name) == 0 || len(name) > 255 || name != strings.TrimSpace(name) ||
		strings.ContainsAny(name, `/\`) || filepath.Base(name) != name ||
		strings.IndexFunc(name, func(r rune) bool { return unicode.IsControl(r) }) >= 0 ||
		!supported || strings.ToLower(filepath.Ext(name)) != definition.Extension || !validContentType ||
		size <= 0 || size > MaxFileBytes {
		return ErrInvalidUpload
	}
	return nil
}

func (s *Service) checkProduct(ctx context.Context, productID string) (uuid.UUID, error) {
	id, err := uuid.Parse(productID)
	if err != nil {
		return uuid.Nil, ErrInvalidInput
	}
	if _, err := s.products.GetProduct(ctx, id.String()); err != nil {
		if errors.Is(err, products.ErrNotFound) {
			return uuid.Nil, ErrReferenceNotFound
		}
		return uuid.Nil, fmt.Errorf("find product for file operation: %w", err)
	}
	return id, nil
}

func (s *Service) RequestUpload(ctx context.Context, input RequestUploadInput) (UploadAuthorization, error) {
	if len(s.keys) == 0 {
		return UploadAuthorization{}, ErrUploadUnavailable
	}
	id, err := uuid.Parse(input.ProductID)
	if err != nil {
		return UploadAuthorization{}, ErrInvalidUpload
	}
	if err := validateFile(input.FileName, input.Format, input.ContentType, input.FileSizeBytes); err != nil {
		return UploadAuthorization{}, err
	}
	if _, err := s.checkProduct(ctx, id.String()); err != nil {
		return UploadAuthorization{}, err
	}
	uploadID, err := uuid.NewRandom()
	if err != nil {
		return UploadAuthorization{}, fmt.Errorf("generate file key: %w", err)
	}
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return UploadAuthorization{}, fmt.Errorf("generate upload nonce: %w", err)
	}
	key := fmt.Sprintf("products/%s/files/%s.%s", id, uploadID, input.Format)
	expires := s.now().UTC().Add(UploadTTL).Truncate(time.Second)
	c := claims{uint32(s.activeVersion), id, nonce, input.FileSizeBytes, expires.Unix(), key, input.FileName, input.Format, input.ContentType}
	token := signClaims(c, s.keys[s.activeVersion].Signing)
	request, err := s.storage.PresignPut(ctx, storage.FilesBucket, key, input.ContentType, UploadTTL)
	if err != nil {
		return UploadAuthorization{}, fmt.Errorf("%w: presign file upload: %w", ErrStorageFailure, err)
	}
	return UploadAuthorization{UploadToken: token, UploadURL: request.URL, Headers: request.Headers, ExpiresAt: expires}, nil
}

func (s *Service) ConfirmUpload(ctx context.Context, input ConfirmUploadInput) (ProductFile, error) {
	id, err := uuid.Parse(input.ProductID)
	if err != nil {
		return ProductFile{}, ErrInvalidUpload
	}
	digest := tokenDigest(input.UploadToken)
	consumed, err := s.repo.GetConsumptionByDigest(ctx, digest)
	if err == nil {
		if consumed.ProductID != id.String() {
			return ProductFile{}, ErrInvalidUpload
		}
		return consumed.productFile(), nil
	}
	if !errors.Is(err, ErrFileNotFound) {
		return ProductFile{}, fmt.Errorf("find consumed upload: %w", err)
	}
	if len(s.keys) == 0 {
		return ProductFile{}, ErrUploadUnavailable
	}
	c, payload, err := verifyClaims(input.UploadToken, s.keys)
	if err != nil {
		return ProductFile{}, err
	}
	if c.ProductID != id || c.FileKey == "" || c.FileSizeBytes <= 0 || c.FileSizeBytes > MaxFileBytes ||
		validateFile(c.FileName, c.Format, c.ContentType, c.FileSizeBytes) != nil ||
		!strings.HasPrefix(c.FileKey, "products/"+id.String()+"/files/") {
		return ProductFile{}, ErrInvalidUpload
	}
	fp := fingerprint(payload, s.keys[int(c.KeyVersion)].Fingerprint)
	existing, err := s.repo.GetConsumptionByKey(ctx, c.FileKey)
	if err == nil {
		return sameConsumption(existing, id.String(), digest, fp, int(c.KeyVersion))
	}
	if !errors.Is(err, ErrFileNotFound) {
		return ProductFile{}, fmt.Errorf("find confirmed file: %w", err)
	}
	if !s.now().Before(time.Unix(c.ExpiresAt, 0)) {
		return s.recoverConfirmed(ctx, id.String(), digest, ErrExpiredUpload)
	}
	if _, err := s.checkProduct(ctx, id.String()); err != nil {
		if errors.Is(err, ErrReferenceNotFound) {
			return s.recoverConfirmed(ctx, id.String(), digest, err)
		}
		return ProductFile{}, err
	}
	metadata, err := s.storage.HeadObject(ctx, storage.FilesBucket, c.FileKey)
	if err != nil {
		var responseError interface{ HTTPStatusCode() int }
		if errors.As(err, &responseError) && responseError.HTTPStatusCode() == http.StatusNotFound {
			return s.recoverConfirmed(ctx, id.String(), digest, ErrObjectNotFound)
		}
		return s.recoverConfirmed(ctx, id.String(), digest, fmt.Errorf("%w: head uploaded file: %w", ErrStorageFailure, err))
	}
	if metadata.ContentLength != c.FileSizeBytes || metadata.ContentType != c.ContentType {
		return s.recoverConfirmed(ctx, id.String(), digest, ErrObjectMismatch)
	}
	data, err := s.storage.ReadObject(ctx, storage.FilesBucket, c.FileKey, metadata.ETag, MaxFileBytes)
	if err != nil {
		return s.recoverConfirmed(ctx, id.String(), digest, fmt.Errorf("%w: read uploaded file: %w", ErrStorageFailure, err))
	}
	if int64(len(data)) != metadata.ContentLength || !validFileBytes(c.Format, data) {
		return s.recoverConfirmed(ctx, id.String(), digest, ErrObjectMismatch)
	}
	if !s.now().Before(time.Unix(c.ExpiresAt, 0)) {
		return s.recoverConfirmed(ctx, id.String(), digest, ErrExpiredUpload)
	}
	displayName := input.DisplayName
	if displayName == "" {
		displayName = c.FileName
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || utf8.RuneCountInString(displayName) > 500 || input.DisplayOrder < 0 {
		return ProductFile{}, ErrInvalidInput
	}
	var etag *string
	if metadata.ETag != "" {
		etag = &metadata.ETag
	}
	file, err := s.repo.CreateConfirmed(ctx, ProductFile{ProductID: id.String(), FileKey: c.FileKey, FileName: c.FileName,
		DisplayName: displayName, Format: c.Format, ContentType: c.ContentType, FileSizeBytes: metadata.ContentLength,
		R2ETag: etag, AvailabilityStatus: "active", ConfirmationFingerprint: fp,
		FingerprintKeyVersion: int(c.KeyVersion), DisplayOrder: input.DisplayOrder},
		Consumption{FileKey: c.FileKey, ProductID: id.String(), TokenDigest: digest,
			ConfirmationFingerprint: fp, FingerprintKeyVersion: int(c.KeyVersion),
			ExpiresAt: timePtr(time.Unix(c.ExpiresAt, 0).UTC())})
	if errors.Is(err, ErrConflict) {
		winner, lookupErr := s.repo.GetConsumptionByKey(ctx, c.FileKey)
		if lookupErr == nil {
			return sameConsumption(winner, id.String(), digest, fp, int(c.KeyVersion))
		}
		if !errors.Is(lookupErr, ErrFileNotFound) {
			return ProductFile{}, lookupErr
		}
		return s.recoverConfirmed(ctx, id.String(), digest, err)
	}
	if errors.Is(err, ErrReferenceNotFound) {
		return s.recoverConfirmed(ctx, id.String(), digest, err)
	}
	if err != nil {
		return ProductFile{}, err
	}
	return file, nil
}

func validFileBytes(format string, data []byte) bool {
	switch format {
	case "png":
		return bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n"))
	case "pdf":
		return bytes.HasPrefix(data, []byte("%PDF-"))
	case "svg":
		head := bytes.TrimSpace(data)
		if bytes.HasPrefix(head, []byte("<?xml")) {
			end := bytes.Index(head, []byte("?>"))
			if end < 0 {
				return false
			}
			head = bytes.TrimSpace(head[end+2:])
		}
		return bytes.HasPrefix(head, []byte("<svg>")) || bytes.HasPrefix(head, []byte("<svg ")) || bytes.HasPrefix(head, []byte("<svg\n"))
	case "dxf":
		return bytes.HasPrefix(data, []byte("AutoCAD Binary DXF")) || bytes.HasPrefix(bytes.TrimSpace(data), []byte("0\nSECTION")) || bytes.HasPrefix(bytes.TrimSpace(data), []byte("0\r\nSECTION"))
	case "zip":
		return bytes.HasPrefix(data, []byte("PK\x03\x04")) || bytes.HasPrefix(data, []byte("PK\x05\x06")) || bytes.HasPrefix(data, []byte("PK\x07\x08"))
	default:
		return false
	}
}

func timePtr(t time.Time) *time.Time { return &t }

// Recheck only after a failure that could race with another confirmation.
// A digest hit identifies the exact token previously verified at commit time.
func (s *Service) recoverConfirmed(ctx context.Context, productID, digest string, originalErr error) (ProductFile, error) {
	record, err := s.repo.GetConsumptionByDigest(ctx, digest)
	if err == nil {
		if record.ProductID != productID {
			return ProductFile{}, ErrInvalidUpload
		}
		return record.productFile(), nil
	}
	if errors.Is(err, ErrFileNotFound) {
		return ProductFile{}, originalErr
	}
	return ProductFile{}, fmt.Errorf("recheck consumed upload: %w", err)
}

func sameConsumption(record Consumption, productID, digest, fp string, version int) (ProductFile, error) {
	if record.ProductID != productID || record.FingerprintKeyVersion != version || record.ConfirmationFingerprint != fp ||
		(record.TokenDigest != "" && record.TokenDigest != digest) {
		return ProductFile{}, ErrConflict
	}
	return record.productFile(), nil
}

func (s *Service) List(ctx context.Context, productID string) ([]ProductFile, error) {
	id, err := s.checkProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	return s.repo.List(ctx, id.String())
}

func (s *Service) UpdateMetadata(ctx context.Context, input UpdateMetadataInput) (ProductFile, error) {
	if _, err := uuid.Parse(input.FileID); err != nil {
		return ProductFile{}, ErrInvalidInput
	}
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.DisplayName == "" || utf8.RuneCountInString(input.DisplayName) > 500 || input.DisplayOrder < 0 {
		return ProductFile{}, ErrInvalidInput
	}
	id, err := s.checkProduct(ctx, input.ProductID)
	if err != nil {
		return ProductFile{}, err
	}
	input.ProductID = id.String()
	return s.repo.UpdateMetadata(ctx, input)
}

func (s *Service) Delete(ctx context.Context, productID, fileID string) error {
	if _, err := uuid.Parse(fileID); err != nil {
		return ErrInvalidInput
	}
	id, err := s.checkProduct(ctx, productID)
	if err != nil {
		return err
	}
	file, err := s.repo.GetByIDAndProduct(ctx, id.String(), fileID)
	if err != nil {
		return err
	}
	consumed, err := s.repo.GetConsumptionByKey(ctx, file.FileKey)
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			return ErrConflict
		}
		return err
	}
	if !consumptionMatchesFile(consumed, file) {
		return ErrConflict
	}
	// The row retains the key if R2 or PostgreSQL deletion fails, allowing retry.
	if err := s.storage.DeleteObject(ctx, storage.FilesBucket, file.FileKey); err != nil {
		return fmt.Errorf("%w: delete R2 file: %w", ErrStorageFailure, err)
	}
	return s.repo.DeleteByIDAndProduct(ctx, id.String(), fileID)
}

func consumptionMatchesFile(record Consumption, file ProductFile) bool {
	return record.FileKey == file.FileKey && record.ProductID == file.ProductID &&
		record.Snapshot.ID == file.ID && record.Snapshot.ProductID == file.ProductID &&
		record.Snapshot.FileName == file.FileName && record.Snapshot.Format == file.Format &&
		record.Snapshot.ContentType == file.ContentType && record.Snapshot.FileSizeBytes == file.FileSizeBytes &&
		record.Snapshot.CreatedAt.Equal(file.CreatedAt) &&
		record.ConfirmationFingerprint == file.ConfirmationFingerprint &&
		record.FingerprintKeyVersion == file.FingerprintKeyVersion
}

func (s *Service) Reorder(ctx context.Context, productID string, ids []string) error {
	id, err := s.checkProduct(ctx, productID)
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(ids))
	for _, value := range ids {
		parsed, err := uuid.Parse(value)
		if err != nil || seen[parsed.String()] {
			return ErrInvalidInput
		}
		seen[parsed.String()] = true
	}
	return s.repo.Reorder(ctx, id.String(), ids)
}
