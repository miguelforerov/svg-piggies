package productfiles

import (
	"errors"
	"time"
)

const (
	MaxFileBytes int64 = 50_000_000
	UploadTTL          = 10 * time.Minute
)

var (
	ErrInvalidInput      = errors.New("invalid product file input")
	ErrInvalidUpload     = errors.New("invalid product file upload")
	ErrExpiredUpload     = errors.New("product file upload expired")
	ErrReferenceNotFound = errors.New("product not found")
	ErrFileNotFound      = errors.New("product file not found")
	ErrObjectNotFound    = errors.New("uploaded file not found")
	ErrObjectMismatch    = errors.New("uploaded file metadata does not match authorization")
	ErrConflict          = errors.New("product file conflicts with an existing record")
	ErrStorageFailure    = errors.New("file storage operation failed")
	ErrUploadUnavailable = errors.New("file uploads are unavailable")
)

type ProductFile struct {
	ID                      string
	ProductID               string
	FileKey                 string
	FileName                string
	DisplayName             string
	Format                  string
	ContentType             string
	FileSizeBytes           int64
	R2ETag                  *string
	AvailabilityStatus      string
	ConfirmationFingerprint string
	FingerprintKeyVersion   int
	DisplayOrder            int
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// ConfirmationSnapshot contains only fields returned by the admin confirmation API.
// It remains immutable after the original confirmation, including after deletion.
type ConfirmationSnapshot struct {
	ID                 string    `json:"id"`
	ProductID          string    `json:"productId"`
	FileName           string    `json:"fileName"`
	DisplayName        string    `json:"displayName"`
	Format             string    `json:"format"`
	ContentType        string    `json:"contentType"`
	FileSizeBytes      int64     `json:"fileSizeBytes"`
	AvailabilityStatus string    `json:"availabilityStatus"`
	DisplayOrder       int       `json:"displayOrder"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func snapshotOf(f ProductFile) ConfirmationSnapshot {
	return ConfirmationSnapshot{f.ID, f.ProductID, f.FileName, f.DisplayName, f.Format, f.ContentType,
		f.FileSizeBytes, f.AvailabilityStatus, f.DisplayOrder, f.CreatedAt, f.UpdatedAt}
}

func (s ConfirmationSnapshot) productFile() ProductFile {
	return ProductFile{ID: s.ID, ProductID: s.ProductID, FileName: s.FileName, DisplayName: s.DisplayName,
		Format: s.Format, ContentType: s.ContentType, FileSizeBytes: s.FileSizeBytes,
		AvailabilityStatus: s.AvailabilityStatus, DisplayOrder: s.DisplayOrder,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

type Consumption struct {
	FileKey                 string
	ProductID               string
	TokenDigest             string // Empty only for records imported from pre-consumption installations.
	ConfirmationFingerprint string
	FingerprintKeyVersion   int
	ExpiresAt               *time.Time // Unknown for imported records.
	Snapshot                ConfirmationSnapshot
}

func (c Consumption) productFile() ProductFile {
	f := c.Snapshot.productFile()
	f.FileKey = c.FileKey
	f.ConfirmationFingerprint = c.ConfirmationFingerprint
	f.FingerprintKeyVersion = c.FingerprintKeyVersion
	return f
}

type RequestUploadInput struct {
	ProductID     string
	FileName      string
	Format        string
	ContentType   string
	FileSizeBytes int64
}

type UploadAuthorization struct {
	UploadToken string
	UploadURL   string
	Headers     map[string][]string
	ExpiresAt   time.Time
}

type ConfirmUploadInput struct {
	ProductID    string
	UploadToken  string
	DisplayName  string
	DisplayOrder int
}

type UpdateMetadataInput struct {
	ProductID    string
	FileID       string
	DisplayName  string
	DisplayOrder int
}

type KeyPair struct {
	Signing     []byte
	Fingerprint []byte
}
