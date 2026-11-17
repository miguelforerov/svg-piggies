package productfiles

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

const testProduct = "b6140bec-35f1-4e68-b54c-d4c8070bb987"

type productStub struct{ err error }

func (p productStub) GetProduct(_ context.Context, id string) (products.Product, error) {
	return products.Product{ID: id}, p.err
}

type repoStub struct {
	mu        sync.Mutex
	file      *ProductFile
	consumed  *Consumption
	onKeyMiss func()
	err       error
	createErr error
	deleteErr error
	creates   int
}

func (r *repoStub) List(context.Context, string) ([]ProductFile, error) {
	if r.file == nil {
		return []ProductFile{}, nil
	}
	return []ProductFile{*r.file}, nil
}
func (r *repoStub) GetConsumptionByKey(_ context.Context, key string) (Consumption, error) {
	r.mu.Lock()
	if r.err != nil {
		r.mu.Unlock()
		return Consumption{}, r.err
	}
	if r.consumed == nil || r.consumed.FileKey != key {
		hook := r.onKeyMiss
		r.mu.Unlock()
		if hook != nil {
			hook()
		}
		return Consumption{}, ErrFileNotFound
	}
	result := *r.consumed
	r.mu.Unlock()
	return result, nil
}
func (r *repoStub) GetConsumptionByDigest(_ context.Context, digest string) (Consumption, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return Consumption{}, r.err
	}
	if r.consumed == nil || r.consumed.TokenDigest != digest {
		return Consumption{}, ErrFileNotFound
	}
	return *r.consumed, nil
}
func (r *repoStub) GetByIDAndProduct(_ context.Context, _, _ string) (ProductFile, error) {
	if r.file == nil {
		return ProductFile{}, ErrFileNotFound
	}
	return *r.file, nil
}
func (r *repoStub) CreateConfirmed(_ context.Context, f ProductFile, c Consumption) (ProductFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.creates++
	if r.createErr != nil {
		return ProductFile{}, r.createErr
	}
	if r.file != nil || r.consumed != nil {
		return ProductFile{}, ErrConflict
	}
	f.ID = uuid.NewString()
	r.file = &f
	c.Snapshot = snapshotOf(f)
	r.consumed = &c
	return f, nil
}
func (r *repoStub) UpdateMetadata(_ context.Context, in UpdateMetadataInput) (ProductFile, error) {
	if r.file == nil {
		return ProductFile{}, ErrFileNotFound
	}
	r.file.DisplayName = in.DisplayName
	r.file.DisplayOrder = in.DisplayOrder
	return *r.file, nil
}
func (r *repoStub) DeleteByIDAndProduct(context.Context, string, string) error {
	if r.file == nil {
		return ErrFileNotFound
	}
	if r.consumed == nil || !consumptionMatchesFile(*r.consumed, *r.file) {
		return ErrConflict
	}
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.file = nil
	return nil
}
func (r *repoStub) Reorder(context.Context, string, []string) error { return nil }

type storeStub struct {
	mu         sync.Mutex
	onHead     func(int)
	onRead     func(int)
	heads      int
	reads      int
	deletes    int
	presignErr error
	headErr    error
	deleteErr  error
	metadata   storage.ObjectMetadata
	data       []byte
	readErr    error
}

func (s *storeStub) ReadObject(_ context.Context, b storage.Bucket, _, _ string, _ int64) ([]byte, error) {
	if b != storage.FilesBucket {
		panic("wrong bucket")
	}
	s.mu.Lock()
	s.reads++
	n, hook := s.reads, s.onRead
	s.mu.Unlock()
	if hook != nil {
		hook(n)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data, s.readErr
}

func (s *storeStub) PresignPut(_ context.Context, b storage.Bucket, _, _ string, _ time.Duration) (storage.PresignedRequest, error) {
	if b != storage.FilesBucket {
		panic("wrong bucket")
	}
	return storage.PresignedRequest{URL: "https://r2.test/upload", Headers: http.Header{"If-None-Match": []string{"*"}}}, s.presignErr
}
func (s *storeStub) HeadObject(_ context.Context, b storage.Bucket, _ string) (storage.ObjectMetadata, error) {
	if b != storage.FilesBucket {
		panic("wrong bucket")
	}
	s.mu.Lock()
	s.heads++
	n, hook := s.heads, s.onHead
	s.mu.Unlock()
	if hook != nil {
		hook(n)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.metadata, s.headErr
}
func (s *storeStub) DeleteObject(_ context.Context, b storage.Bucket, _ string) error {
	if b != storage.FilesBucket {
		panic("wrong bucket")
	}
	s.deletes++
	return s.deleteErr
}

func fixture(t *testing.T) (*Service, *repoStub, *storeStub) {
	t.Helper()
	r := &repoStub{}
	data := []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>")
	for len(data) < 123 {
		data = append(data, ' ')
	}
	st := &storeStub{metadata: storage.ObjectMetadata{ContentLength: 123, ContentType: "image/svg+xml", ETag: "etag"}, data: data}
	keys := map[int]KeyPair{1: {Signing: []byte("11111111111111111111111111111111"), Fingerprint: []byte("22222222222222222222222222222222")}, 2: {Signing: []byte("33333333333333333333333333333333"), Fingerprint: []byte("44444444444444444444444444444444")}}
	s, err := NewService(productStub{}, r, st, keys, 2)
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Unix(1_700_000_000, 0) }
	return s, r, st
}

func authorize(t *testing.T, s *Service, format, mime, name string, size int64) UploadAuthorization {
	t.Helper()
	a, err := s.RequestUpload(context.Background(), RequestUploadInput{ProductID: testProduct, FileName: name, Format: format, ContentType: mime, FileSizeBytes: size})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestRequestUploadFormatsAndValidation(t *testing.T) {
	for _, tc := range []struct{ format, mime, name string }{{"svg", "image/svg+xml", "pig.svg"}, {"png", "image/png", "pig.png"}, {"dxf", "image/vnd.dxf", "pig.dxf"}, {"pdf", "application/pdf", "pig.pdf"}} {
		s, _, _ := fixture(t)
		a := authorize(t, s, tc.format, tc.mime, tc.name, 123)
		if a.UploadToken == "" || a.UploadURL == "" || len(a.Headers["If-None-Match"]) != 1 || a.Headers["If-None-Match"][0] != "*" {
			t.Fatalf("bad authorization: %+v", a)
		}
	}
	for _, tc := range []RequestUploadInput{{ProductID: testProduct, FileName: "pig.exe", Format: "exe", ContentType: "application/octet-stream", FileSizeBytes: 1}, {ProductID: testProduct, FileName: "pig.svg", Format: "svg", ContentType: "application/pdf", FileSizeBytes: 1}, {ProductID: testProduct, FileName: "pig.svg", Format: "svg", ContentType: "image/svg+xml", FileSizeBytes: MaxFileBytes + 1}} {
		s, _, _ := fixture(t)
		if _, err := s.RequestUpload(context.Background(), tc); !errors.Is(err, ErrInvalidUpload) {
			t.Fatalf("want invalid upload, got %v", err)
		}
	}
	s, _, st := fixture(t)
	st.presignErr = errors.New("r2 unavailable")
	if _, err := s.RequestUpload(context.Background(), RequestUploadInput{ProductID: testProduct, FileName: "pig.svg", Format: "svg", ContentType: "image/svg+xml", FileSizeBytes: 123}); !errors.Is(err, ErrStorageFailure) {
		t.Fatal(err)
	}
	s.products = productStub{err: products.ErrNotFound}
	if _, err := s.RequestUpload(context.Background(), RequestUploadInput{ProductID: testProduct, FileName: "pig.svg", Format: "svg", ContentType: "image/svg+xml", FileSizeBytes: 123}); !errors.Is(err, ErrReferenceNotFound) {
		t.Fatal(err)
	}
}

func TestConfirmIdempotencyAndExpiry(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	in := ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}
	first, err := s.ConfirmUpload(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return a.ExpiresAt.Add(time.Second) }
	again, err := s.ConfirmUpload(context.Background(), in)
	if err != nil || again.ID != first.ID || r.creates != 1 || st.heads != 1 {
		t.Fatalf("idempotency failed: %v %+v", err, again)
	}
	other := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	claims1, payload, err := verifyClaims(other.UploadToken, s.keys)
	if err != nil {
		t.Fatal(err)
	}
	claims1.FileKey = first.FileKey
	other.UploadToken = signClaims(claims1, s.keys[int(claims1.KeyVersion)].Signing)
	_ = payload
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: other.UploadToken}); !errors.Is(err, ErrConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
}

func TestConsumedTokenSurvivesDeleteProductRemovalAndKeyRetirement(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	in := ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}
	first, err := s.ConfirmUpload(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateMetadata(context.Background(), UpdateMetadataInput{ProductID: testProduct, FileID: first.ID,
		DisplayName: "Renamed", DisplayOrder: 0}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(context.Background(), testProduct, first.ID); err != nil {
		t.Fatal(err)
	}
	if r.file != nil || r.consumed == nil {
		t.Fatal("deletion did not retain only the consumption")
	}
	s.now = func() time.Time { return a.ExpiresAt.Add(time.Hour) }
	s.products = productStub{err: products.ErrNotFound}
	s.keys[3] = KeyPair{Signing: []byte("55555555555555555555555555555555"), Fingerprint: []byte("66666666666666666666666666666666")}
	s.activeVersion = 3
	delete(s.keys, 2)
	again, err := s.ConfirmUpload(context.Background(), in)
	if err != nil || again.ID != first.ID || again.DisplayName != first.DisplayName || r.creates != 1 || st.heads != 1 || st.reads != 1 {
		t.Fatalf("historical confirmation failed: %v %+v", err, again)
	}
	in.UploadToken += "x"
	if _, err := s.ConfirmUpload(context.Background(), in); !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("altered token recovered a consumption: %v", err)
	}
}

func TestDifferentAuthorizationCannotReuseDeletedKey(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	first, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(context.Background(), testProduct, first.ID); err != nil {
		t.Fatal(err)
	}
	other := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	c, _, err := verifyClaims(other.UploadToken, s.keys)
	if err != nil {
		t.Fatal(err)
	}
	c.FileKey = r.consumed.FileKey
	other.UploadToken = signClaims(c, s.keys[int(c.KeyVersion)].Signing)
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: other.UploadToken}); !errors.Is(err, ErrConflict) {
		t.Fatalf("different authorization reused deleted key: %v", err)
	}
	if r.creates != 1 || st.heads != 1 {
		t.Fatal("conflicting authorization reached storage or created a new file")
	}
}

func TestConfirmRecoversConsumptionAfterConcurrentExpiryOrProductRemoval(t *testing.T) {
	for _, scenario := range []string{"expired", "product removed"} {
		t.Run(scenario, func(t *testing.T) {
			s, r, st := fixture(t)
			a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
			entered, release := make(chan struct{}), make(chan struct{})
			var misses atomic.Int32
			r.onKeyMiss = func() {
				if misses.Add(1) == 1 {
					close(entered)
					<-release
				}
			}
			type outcome struct {
				file ProductFile
				err  error
			}
			finished := make(chan outcome, 1)
			in := ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}
			go func() {
				file, err := s.ConfirmUpload(context.Background(), in)
				finished <- outcome{file, err}
			}()
			<-entered
			first, err := s.ConfirmUpload(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "expired" {
				s.now = func() time.Time { return a.ExpiresAt.Add(time.Second) }
			} else {
				s.products = productStub{err: products.ErrNotFound}
			}
			close(release)
			result := <-finished
			if result.err != nil || result.file.ID != first.ID || st.heads != 1 {
				t.Fatalf("concurrent historical recovery: %+v heads=%d", result, st.heads)
			}
		})
	}
}

func TestConfirmRecoversConsumptionAfterConcurrentR2Failure(t *testing.T) {
	for _, scenario := range []string{"missing", "head failure", "read failure"} {
		t.Run(scenario, func(t *testing.T) {
			s, _, st := fixture(t)
			a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
			entered, release := make(chan struct{}), make(chan struct{})
			gate := func(n int) {
				if n == 1 {
					close(entered)
					<-release
				}
			}
			if scenario == "read failure" {
				st.onRead = gate
			} else {
				st.onHead = gate
			}
			type outcome struct {
				file ProductFile
				err  error
			}
			finished := make(chan outcome, 1)
			in := ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}
			go func() {
				file, err := s.ConfirmUpload(context.Background(), in)
				finished <- outcome{file, err}
			}()
			<-entered
			first, err := s.ConfirmUpload(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "missing" {
				if err := s.Delete(context.Background(), testProduct, first.ID); err != nil {
					t.Fatal(err)
				}
			}
			st.mu.Lock()
			switch scenario {
			case "missing":
				st.headErr = notFoundStorageError{}
			case "head failure":
				st.headErr = errors.New("R2 unavailable")
			case "read failure":
				st.readErr = errors.New("R2 read unavailable")
			}
			st.mu.Unlock()
			close(release)
			result := <-finished
			if result.err != nil || result.file.ID != first.ID {
				t.Fatalf("concurrent historical recovery: %+v", result)
			}
		})
	}
}

func TestDeleteRequiresMatchingConsumptionBeforeR2(t *testing.T) {
	for _, scenario := range []string{"missing", "wrong product", "wrong fingerprint", "wrong file id", "wrong file name"} {
		t.Run(scenario, func(t *testing.T) {
			s, r, st := fixture(t)
			r.file = &ProductFile{ID: uuid.NewString(), ProductID: testProduct, FileKey: "products/key",
				ConfirmationFingerprint: strings.Repeat("a", 64), FingerprintKeyVersion: 1}
			if scenario != "missing" {
				r.consumed = &Consumption{FileKey: r.file.FileKey, ProductID: testProduct,
					ConfirmationFingerprint: r.file.ConfirmationFingerprint, FingerprintKeyVersion: 1,
					Snapshot: snapshotOf(*r.file)}
				switch scenario {
				case "wrong product":
					r.consumed.ProductID = uuid.NewString()
				case "wrong fingerprint":
					r.consumed.ConfirmationFingerprint = strings.Repeat("b", 64)
				case "wrong file id":
					r.consumed.Snapshot.ID = uuid.NewString()
				case "wrong file name":
					r.consumed.Snapshot.FileName = "other.svg"
				}
			}
			if err := s.Delete(context.Background(), testProduct, r.file.ID); !errors.Is(err, ErrConflict) || st.deletes != 0 || r.file == nil {
				t.Fatalf("invalid consumption allowed R2 delete: %v deletes=%d", err, st.deletes)
			}
		})
	}
}

func TestConfirmRejectsInvalidAndExpired(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	in := ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}
	in.UploadToken += "x"
	if _, err := s.ConfirmUpload(context.Background(), in); !errors.Is(err, ErrInvalidUpload) {
		t.Fatal(err)
	}
	in.UploadToken = a.UploadToken
	in.ProductID = uuid.NewString()
	if _, err := s.ConfirmUpload(context.Background(), in); !errors.Is(err, ErrInvalidUpload) {
		t.Fatal(err)
	}
	in.ProductID = testProduct
	s.now = func() time.Time { return a.ExpiresAt }
	if _, err := s.ConfirmUpload(context.Background(), in); !errors.Is(err, ErrExpiredUpload) {
		t.Fatal(err)
	}
	if r.creates != 0 || st.heads != 0 {
		t.Fatal("expired upload touched storage or database")
	}
}

func TestConfirmObjectErrors(t *testing.T) {
	for _, tc := range []struct {
		metadata storage.ObjectMetadata
		err      error
		want     error
	}{{storage.ObjectMetadata{ContentLength: 1, ContentType: "image/svg+xml"}, nil, ErrObjectMismatch}, {storage.ObjectMetadata{}, errors.New("storage unavailable"), ErrStorageFailure}} {
		s, _, st := fixture(t)
		st.metadata = tc.metadata
		st.headErr = tc.err
		a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
		if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}); !errors.Is(err, tc.want) {
			t.Fatalf("want %v, got %v", tc.want, err)
		}
	}
}

func TestDeleteFailureKeepsMetadata(t *testing.T) {
	s, r, st := fixture(t)
	r.file = &ProductFile{ID: uuid.NewString(), ProductID: testProduct, FileKey: "products/key"}
	r.consumed = &Consumption{FileKey: r.file.FileKey, ProductID: testProduct, Snapshot: snapshotOf(*r.file)}
	st.deleteErr = errors.New("r2 unavailable")
	if err := s.Delete(context.Background(), testProduct, r.file.ID); !errors.Is(err, ErrStorageFailure) || r.file == nil {
		t.Fatalf("metadata lost: %v", err)
	}
}

func TestConcurrentConfirmationsShareOneRecord(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	var wg sync.WaitGroup
	results := make([]ProductFile, 2)
	errs := make([]error, 2)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken})
		}(i)
	}
	wg.Wait()
	if errs[0] != nil || errs[1] != nil || results[0].ID != results[1].ID || r.creates != 2 && r.creates != 1 || st.heads < 1 {
		t.Fatalf("concurrent confirmation: results=%+v errors=%v creates=%d", results, errs, r.creates)
	}
}

func TestConfirmDatabaseAndReadErrors(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	r.err = errors.New("database unavailable")
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}); err == nil || st.heads != 0 {
		t.Fatalf("database error: %v heads=%d", err, st.heads)
	}
	r.err = nil
	st.readErr = errors.New("R2 read failed")
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}); !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("read error: %v", err)
	}
	st.readErr = nil
	st.data = []byte("not SVG")
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}); !errors.Is(err, ErrObjectMismatch) {
		t.Fatalf("content mismatch: %v", err)
	}
}

func TestDeleteDatabaseFailureKeepsRetriableRow(t *testing.T) {
	s, r, st := fixture(t)
	r.file = &ProductFile{ID: uuid.NewString(), ProductID: testProduct, FileKey: "products/key"}
	r.consumed = &Consumption{FileKey: r.file.FileKey, ProductID: testProduct, Snapshot: snapshotOf(*r.file)}
	r.deleteErr = errors.New("database unavailable")
	if err := s.Delete(context.Background(), testProduct, r.file.ID); err == nil || st.deletes != 1 || r.file == nil {
		t.Fatalf("partial deletion: %v", err)
	}
}

func TestKeyRotationKeepsPriorVersionValid(t *testing.T) {
	s, _, _ := fixture(t)
	s.activeVersion = 1
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	s.activeVersion = 2
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}); err != nil {
		t.Fatalf("old version rejected: %v", err)
	}
}

func TestFileSignatures(t *testing.T) {
	for _, tc := range []struct {
		format string
		valid  []byte
	}{{"svg", []byte("<svg></svg>")}, {"png", []byte("\x89PNG\r\n\x1a\n")}, {"dxf", []byte("0\nSECTION")}, {"pdf", []byte("%PDF-1.7")}} {
		if !validFileBytes(tc.format, tc.valid) || validFileBytes(tc.format, []byte("wrong")) {
			t.Fatalf("signature validation for %s", tc.format)
		}
	}
}

type notFoundStorageError struct{}

func (notFoundStorageError) Error() string       { return "not found" }
func (notFoundStorageError) HTTPStatusCode() int { return http.StatusNotFound }

func TestConfirmMissingObjectAndCreateFailure(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	in := ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken}
	st.headErr = notFoundStorageError{}
	if _, err := s.ConfirmUpload(context.Background(), in); !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("missing object: %v", err)
	}
	st.headErr = nil
	r.createErr = errors.New("database insert failed")
	if _, err := s.ConfirmUpload(context.Background(), in); err == nil || r.file != nil {
		t.Fatalf("database failure: %v", err)
	}
}

func TestInvalidSignatureNeverRetrievesExistingFile(t *testing.T) {
	s, r, st := fixture(t)
	a := authorize(t, s, "svg", "image/svg+xml", "pig.svg", 123)
	_, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: a.UploadToken})
	if err != nil {
		t.Fatal(err)
	}
	parts := []byte(a.UploadToken)
	for i := range parts {
		if parts[i] == '.' {
			if parts[i+1] == 'A' {
				parts[i+1] = 'B'
			} else {
				parts[i+1] = 'A'
			}
			break
		}
	}
	if _, err := s.ConfirmUpload(context.Background(), ConfirmUploadInput{ProductID: testProduct, UploadToken: string(parts)}); !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("tampered token: %v", err)
	}
	if r.creates != 1 || st.heads != 1 {
		t.Fatalf("tampered token reached confirmation path: creates=%d heads=%d", r.creates, st.heads)
	}
}

func TestDisplayNameValidation(t *testing.T) {
	s, _, _ := fixture(t)
	if _, err := s.UpdateMetadata(context.Background(), UpdateMetadataInput{ProductID: testProduct, FileID: uuid.NewString(), DisplayName: "   ", DisplayOrder: 0}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("blank name: %v", err)
	}
	if _, err := s.UpdateMetadata(context.Background(), UpdateMetadataInput{ProductID: testProduct, FileID: uuid.NewString(), DisplayName: strings.Repeat("x", 501), DisplayOrder: 0}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("oversized name: %v", err)
	}
}
