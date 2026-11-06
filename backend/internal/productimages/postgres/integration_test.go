//go:build !js

package postgres_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	bobpgx "github.com/stephenafamo/bob/drivers/pgx"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
	imagespostgres "github.com/zdenaforero/svg-piggies/backend/internal/productimages/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	productspostgres "github.com/zdenaforero/svg-piggies/backend/internal/products/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

// These tests deliberately do not read DATABASE_URL or run migrations. Migrate a
// fresh, dedicated test database with dbmate before running them.
func testDatabase(t *testing.T) (string, *database.PoolProvider) {
	t.Helper()
	connectionString := os.Getenv("TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("PostgreSQL integration skipped: TEST_DATABASE_URL is not configured")
	}
	u, err := url.Parse(connectionString)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("TEST_DATABASE_URL must be a PostgreSQL URL")
	}
	name := strings.TrimPrefix(u.Path, "/")
	if !regexp.MustCompile(`(^|[_-])test([_-]|$)`).MatchString(name) || name == "svg_piggies_development" {
		t.Fatalf("TEST_DATABASE_URL must name a dedicated test database; got %q", name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	provider, err := database.NewPoolProvider(ctx, connectionString)
	if err != nil {
		t.Fatalf("connect to test PostgreSQL: %v", err)
	}
	t.Cleanup(provider.Close)
	connection, err := provider.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var actual string
	err = connection.QueryRow(ctx, `SELECT current_database()`).Scan(&actual)
	_ = connection.Release(ctx)
	if err != nil || actual != name {
		t.Fatalf("test database identity mismatch: URL=%q actual=%q error=%v", name, actual, err)
	}
	return name, provider
}

func requireMigrations(t *testing.T, provider *database.PoolProvider) {
	t.Helper()
	ctx := context.Background()
	connection, err := provider.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release(ctx)
	rows, err := connection.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatalf("migrations are absent: %v", err)
	}
	versions := make([]string, 0, 8)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			t.Fatal(err)
		}
		versions = append(versions, version)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	want := []string{"0001", "0002", "0003", "0004", "0005", "0006", "0007", "0008"}
	if !reflect.DeepEqual(versions, want) {
		t.Fatalf("test database migrations = %v; want %v", versions, want)
	}
	var nullable string
	err = connection.QueryRow(ctx, `SELECT is_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'product_images' AND column_name = 'confirmation_fingerprint'`).Scan(&nullable)
	if err != nil || nullable != "YES" {
		t.Fatalf("test database needs current 0002 migration: confirmation_fingerprint nullable=%q error=%v", nullable, err)
	}
}

type fakeStorage struct {
	data               []byte
	mu                 sync.Mutex
	puts, heads, reads int
}

func (s *fakeStorage) PresignPut(_ context.Context, bucket storage.Bucket, key, contentType string, ttl time.Duration) (storage.PresignedRequest, error) {
	if bucket != storage.AssetsBucket || contentType != "image/webp" || ttl != productimages.UploadTTL || key == "" {
		return storage.PresignedRequest{}, errors.New("unexpected PUT arguments")
	}
	s.mu.Lock()
	s.puts++
	s.mu.Unlock()
	return storage.PresignedRequest{URL: "https://storage.test/put", Headers: http.Header{"Content-Type": {contentType}}}, nil
}

func (s *fakeStorage) HeadObject(_ context.Context, bucket storage.Bucket, _ string) (storage.ObjectMetadata, error) {
	if bucket != storage.AssetsBucket {
		return storage.ObjectMetadata{}, errors.New("unexpected bucket")
	}
	s.mu.Lock()
	s.heads++
	s.mu.Unlock()
	return storage.ObjectMetadata{ContentLength: int64(len(s.data)), ContentType: "image/webp", ETag: "integration-etag"}, nil
}

func (s *fakeStorage) ReadObject(_ context.Context, bucket storage.Bucket, _, etag string, limit int64) ([]byte, error) {
	if bucket != storage.AssetsBucket || etag != "integration-etag" || limit != productimages.MaxImageBytes {
		return nil, errors.New("unexpected GET arguments")
	}
	s.mu.Lock()
	s.reads++
	s.mu.Unlock()
	return s.data, nil
}

type fixture struct {
	provider  *database.PoolProvider
	product   *products.Service
	image     *imagespostgres.Repository
	upload    *productimages.UploadService
	store     *fakeStorage
	productID string
	data      []byte
}

type singleConnectionProvider struct{ conn *pgx.Conn }

func (p singleConnectionProvider) Acquire(context.Context) (*database.Connection, error) {
	return database.NewConnection(p.conn, bobpgx.NewConn(p.conn), func(context.Context) error { return nil })
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	_, provider := testDatabase(t)
	requireMigrations(t, provider)
	productRepo, err := productspostgres.NewRepository(provider)
	if err != nil {
		t.Fatal(err)
	}
	productService, err := products.NewService(productRepo)
	if err != nil {
		t.Fatal(err)
	}
	imageRepo, err := imagespostgres.NewRepository(provider)
	if err != nil {
		t.Fatal(err)
	}
	imageService, err := productimages.NewService(imageRepo)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../testdata/valid.webp")
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStorage{data: data}
	upload, err := productimages.NewUploadService(productService, imageService, imageRepo, store, "integration-signing-secret")
	if err != nil {
		t.Fatal(err)
	}
	product, err := productService.CreateProduct(context.Background(), products.CreateProductInput{
		Title: "Integration fixture", Slug: "integration-" + uuid.NewString(), Price: "1.00", Status: products.StatusDraft,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := productService.DeleteProduct(ctx, product.ID); err != nil {
			t.Errorf("delete own fixture product: %v", err)
		}
	})
	return &fixture{provider, productService, imageRepo, upload, store, product.ID, data}
}

func (f *fixture) authorize(t *testing.T) productimages.UploadAuthorization {
	t.Helper()
	a, err := f.upload.RequestUpload(context.Background(), productimages.RequestUploadInput{
		ProductID: f.productID, OriginalFilename: "valid.webp", ContentType: "image/webp", FileSizeBytes: int64(len(f.data)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func (f *fixture) count(t *testing.T, key string) int {
	t.Helper()
	ctx := context.Background()
	connection, err := f.provider.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release(ctx)
	var count int
	if err := connection.QueryRow(ctx, `SELECT count(*) FROM product_images WHERE r2_object_key = $1`, key).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestPostgresIntegrationSchema(t *testing.T) {
	_, provider := testDatabase(t)
	requireMigrations(t, provider)
	repo, err := imagespostgres.NewRepository(provider)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CheckUploadSchema(context.Background()); err != nil {
		t.Fatalf("real schema check: %v", err)
	}
	if _, err := repo.GetByObjectKey(context.Background(), "integration/missing-"+uuid.NewString()); !errors.Is(err, productimages.ErrImageNotFound) {
		t.Fatalf("real lookup query: %v", err)
	}
}

func TestPostgresIntegrationMissingSchemaBlocksPut(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	// This temporary relation shadows only this session's public table. It
	// disappears when the connection closes; public schema stays untouched.
	if _, err := conn.Exec(ctx, `CREATE TEMP TABLE product_images (id integer)`); err != nil {
		t.Fatal(err)
	}
	provider := singleConnectionProvider{conn}
	repo, err := imagespostgres.NewRepository(provider)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CheckUploadSchema(ctx); err == nil {
		t.Fatal("schema check accepted missing column")
	}
	productRepo, err := productspostgres.NewRepository(provider)
	if err != nil {
		t.Fatal(err)
	}
	productService, err := products.NewService(productRepo)
	if err != nil {
		t.Fatal(err)
	}
	upload, err := productimages.NewUploadService(productService, mustImageService(t, repo), repo, f.store, "integration-signing-secret")
	if err != nil {
		t.Fatal(err)
	}
	a, err := upload.RequestUpload(ctx, productimages.RequestUploadInput{ProductID: f.productID, OriginalFilename: "valid.webp", ContentType: "image/webp", FileSizeBytes: int64(len(f.data))})
	if !errors.Is(err, productimages.ErrUploadUnavailable) || a.UploadURL != "" || a.UploadToken != "" || f.store.puts != 0 {
		t.Fatalf("missing schema granted PUT: %+v %v", a, err)
	}
}

func TestPostgresIntegrationFingerprintConstraint(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	for _, fingerprint := range []string{"", strings.Repeat("a", 63), strings.Repeat("A", 64), strings.Repeat("g", 64)} {
		tx, err := pgx.Connect(ctx, os.Getenv("TEST_DATABASE_URL"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, `INSERT INTO product_images(product_id,r2_object_key,original_filename,content_type,file_size_bytes,confirmation_fingerprint) VALUES($1,$2,'x.webp','image/webp',1,$3)`, f.productID, "integration/invalid-"+uuid.NewString(), fingerprint)
		_ = tx.Close(ctx)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
			t.Fatalf("fingerprint %q: expected CHECK violation, got %v", fingerprint, err)
		}
	}
	connection, err := f.provider.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release(ctx)
	tx, err := connection.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO product_images(product_id,r2_object_key,original_filename,content_type,file_size_bytes,confirmation_fingerprint) VALUES($1,$2,'x.webp','image/webp',1,NULL)`, f.productID, "integration/nullable-"+uuid.NewString())
	if err != nil {
		t.Fatalf("NULL fingerprint rejected: %v", err)
	}
}

func TestPostgresIntegrationUploadAndIdempotency(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	missing := uuid.NewString()
	if a, err := f.upload.RequestUpload(ctx, productimages.RequestUploadInput{ProductID: missing, OriginalFilename: "valid.webp", ContentType: "image/webp", FileSizeBytes: int64(len(f.data))}); !errors.Is(err, productimages.ErrReferenceNotFound) || a.UploadURL != "" {
		t.Fatalf("missing product: %+v %v", a, err)
	}
	for _, input := range []productimages.RequestUploadInput{
		{ProductID: f.productID, OriginalFilename: "bad.png", ContentType: "image/webp", FileSizeBytes: 1},
		{ProductID: f.productID, OriginalFilename: "valid.webp", ContentType: "image/svg+xml", FileSizeBytes: 1},
		{ProductID: f.productID, OriginalFilename: "valid.webp", ContentType: "image/webp", FileSizeBytes: 0},
		{ProductID: f.productID, OriginalFilename: "valid.webp", ContentType: "image/webp", FileSizeBytes: productimages.MaxImageBytes + 1},
	} {
		if a, err := f.upload.RequestUpload(ctx, input); !errors.Is(err, productimages.ErrInvalidUpload) || a.UploadURL != "" {
			t.Fatalf("invalid input: %+v %v", a, err)
		}
	}
	a := f.authorize(t)
	if !strings.HasPrefix(a.ObjectKey, "products/"+f.productID+"/images/") || !strings.HasSuffix(a.ObjectKey, ".webp") || a.UploadURL == "" || a.UploadToken == "" {
		t.Fatalf("authorization: %+v", a)
	}
	input := productimages.ConfirmUploadInput{ProductID: f.productID, UploadToken: a.UploadToken, AltText: ptr("fixture alt"), DisplayOrder: 2, IsPrimary: true}
	if _, err := f.upload.ConfirmUpload(ctx, productimages.ConfirmUploadInput{ProductID: f.productID, UploadToken: a.UploadToken + "x"}); !errors.Is(err, productimages.ErrInvalidUpload) {
		t.Fatalf("invalid token: %v", err)
	}
	created, err := f.upload.ConfirmUpload(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := f.image.GetByObjectKey(ctx, a.ObjectKey)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != stored.ID || stored.ProductID != f.productID || stored.R2ObjectKey != a.ObjectKey || stored.OriginalFilename != "valid.webp" || stored.ContentType != "image/webp" || stored.FileSizeBytes != int64(len(f.data)) || stored.WidthPX == nil || *stored.WidthPX <= 0 || stored.HeightPX == nil || *stored.HeightPX <= 0 || stored.AltText == nil || *stored.AltText != "fixture alt" || stored.DisplayOrder != 2 || !stored.IsPrimary || stored.R2ETag == nil || *stored.R2ETag != "integration-etag" || stored.ConfirmationFingerprint == nil || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(*stored.ConfirmationFingerprint) || stored.CreatedAt.IsZero() || stored.UpdatedAt.IsZero() {
		t.Fatalf("persisted metadata: %+v", stored)
	}
	repeated, err := f.upload.ConfirmUpload(ctx, input)
	if err != nil || repeated.ID != created.ID || f.count(t, a.ObjectKey) != 1 {
		t.Fatalf("identical repeat: %+v %v", repeated, err)
	}
	expiredToken := expireToken(t, a.UploadToken)
	expiredRepeat := input
	expiredRepeat.UploadToken = expiredToken
	repeated, err = f.upload.ConfirmUpload(ctx, expiredRepeat)
	if err != nil || repeated.ID != created.ID || f.count(t, a.ObjectKey) != 1 {
		t.Fatalf("expired identical repeat: %+v %v", repeated, err)
	}
	changed := input
	changed.AltText = ptr("different")
	if _, err := f.upload.ConfirmUpload(ctx, changed); !errors.Is(err, productimages.ErrConflict) {
		t.Fatalf("changed repeat: %v", err)
	}
	after, err := f.image.GetByObjectKey(ctx, a.ObjectKey)
	if err != nil || !reflect.DeepEqual(after, stored) {
		t.Fatalf("original changed: %+v %v", after, err)
	}
	service, err := productimages.NewService(f.image)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.AddProductImages(ctx, productimages.AddProductImagesInput{ProductID: f.productID, Images: []productimages.NewProductImage{{R2ObjectKey: a.ObjectKey, OriginalFilename: "valid.webp", ContentType: "image/webp", FileSizeBytes: int64(len(f.data))}}})
	if !errors.Is(err, productimages.ErrConflict) || f.count(t, a.ObjectKey) != 1 {
		t.Fatalf("UNIQUE object key: %v", err)
	}
	if f.store.puts != 1 || f.store.heads != 1 || f.store.reads != 1 {
		t.Fatalf("storage calls: %+v", f.store)
	}
}

func TestPostgresIntegrationExpiredTokenRejectsNewImage(t *testing.T) {
	f := newFixture(t)
	a := f.authorize(t)
	input := productimages.ConfirmUploadInput{ProductID: f.productID, UploadToken: expireToken(t, a.UploadToken)}
	if _, err := f.upload.ConfirmUpload(context.Background(), input); !errors.Is(err, productimages.ErrExpiredUpload) || f.count(t, a.ObjectKey) != 0 {
		t.Fatalf("expired new image: %v", err)
	}
}

func expireToken(t *testing.T, token string) string {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatal("malformed authorization token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	claims["e"] = time.Now().Add(-time.Minute).Unix()
	payload, err = json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	secret := sha256.Sum256([]byte("svg-piggies:image-upload-token:v1:\x00" + "integration-signing-secret"))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestPostgresIntegrationConcurrentConfirmations(t *testing.T) {
	for _, different := range []bool{false, true} {
		name := "identical"
		if different {
			name = "different"
		}
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			a := f.authorize(t)
			const n = 8
			results := make([]productimages.ProductImage, n)
			errs := make([]error, n)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					repo, err := imagespostgres.NewRepository(f.provider)
					if err != nil {
						errs[i] = err
						return
					}
					registrar, err := productimages.NewService(repo)
					if err != nil {
						errs[i] = err
						return
					}
					upload, err := productimages.NewUploadService(f.product, registrar, repo, f.store, "integration-signing-secret")
					if err != nil {
						errs[i] = err
						return
					}
					input := productimages.ConfirmUploadInput{ProductID: f.productID, UploadToken: a.UploadToken}
					if different && i%2 == 1 {
						input.AltText = ptr("different")
					}
					<-start
					results[i], errs[i] = upload.ConfirmUpload(context.Background(), input)
				}(i)
			}
			close(start)
			wg.Wait()
			if f.count(t, a.ObjectKey) != 1 {
				t.Fatal("concurrent confirmation created duplicate rows")
			}
			stored, err := f.image.GetByObjectKey(context.Background(), a.ObjectKey)
			if err != nil {
				t.Fatal(err)
			}
			for i, err := range errs {
				wantSuccess := !different || (i%2 == 1) == (stored.AltText != nil)
				if wantSuccess && (err != nil || results[i].ID != stored.ID) {
					t.Fatalf("confirmation %d: %+v %v", i, results[i], err)
				}
				if !wantSuccess && !errors.Is(err, productimages.ErrConflict) {
					t.Fatalf("confirmation %d: expected conflict, got %v", i, err)
				}
			}
		})
	}
}

func TestPostgresIntegrationDatabaseFailureBlocksAuthorization(t *testing.T) {
	f := newFixture(t)
	_, broken := testDatabase(t)
	broken.Close()
	productRepo, err := productspostgres.NewRepository(broken)
	if err != nil {
		t.Fatal(err)
	}
	productService, err := products.NewService(productRepo)
	if err != nil {
		t.Fatal(err)
	}
	upload, err := productimages.NewUploadService(productService, mustImageService(t, f.image), f.image, f.store, "integration-signing-secret")
	if err != nil {
		t.Fatal(err)
	}
	a, err := upload.RequestUpload(context.Background(), productimages.RequestUploadInput{ProductID: f.productID, OriginalFilename: "valid.webp", ContentType: "image/webp", FileSizeBytes: int64(len(f.data))})
	if err == nil || a.UploadURL != "" || f.store.puts != 0 {
		t.Fatalf("database failure granted PUT: %+v %v", a, err)
	}
}

func mustImageService(t *testing.T, repo *imagespostgres.Repository) *productimages.Service {
	t.Helper()
	s, err := productimages.NewService(repo)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func ptr(s string) *string { return &s }
