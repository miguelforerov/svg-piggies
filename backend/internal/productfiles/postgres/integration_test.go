//go:build !js

package postgres_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/productfiles"
	filespostgres "github.com/zdenaforero/svg-piggies/backend/internal/productfiles/postgres"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	productspostgres "github.com/zdenaforero/svg-piggies/backend/internal/products/postgres"
)

func TestProductFilesPostgresIntegration(t *testing.T) {
	connectionString := os.Getenv("TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}
	u, err := url.Parse(connectionString)
	if err != nil || strings.TrimPrefix(u.Path, "/") != "svg_piggies_test" {
		t.Fatal("integration requires svg_piggies_test")
	}
	ctx := context.Background()
	provider, err := database.NewPoolProvider(ctx, connectionString)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(provider.Close)
	connection, err := provider.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var databaseName string
	if err := connection.QueryRow(ctx, `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatal(err)
	}
	_ = connection.Release(ctx)
	if databaseName != "svg_piggies_test" {
		t.Fatalf("wrong database %q", databaseName)
	}
	productRepo, err := productspostgres.NewRepository(provider)
	if err != nil {
		t.Fatal(err)
	}
	productService, err := products.NewService(productRepo)
	if err != nil {
		t.Fatal(err)
	}
	product, err := productService.CreateProduct(ctx, products.CreateProductInput{Title: "File integration fixture", Slug: "file-integration-" + uuid.NewString(), Price: "1.00", Status: products.StatusDraft})
	if err != nil {
		t.Fatal(err)
	}
	deletedProduct := false
	t.Cleanup(func() {
		if !deletedProduct {
			if err := productService.DeleteProduct(context.Background(), product.ID); err != nil {
				t.Errorf("cleanup product: %v", err)
			}
		}
		c, err := provider.Acquire(context.Background())
		if err != nil {
			t.Errorf("cleanup acquire: %v", err)
			return
		}
		defer c.Release(context.Background())
		if _, err := c.Exec(context.Background(), `DELETE FROM product_file_consumptions WHERE product_id=$1`, product.ID); err != nil {
			t.Errorf("cleanup consumptions: %v", err)
		}
	})
	repo, err := filespostgres.NewRepository(provider)
	if err != nil {
		t.Fatal(err)
	}
	key := "products/" + product.ID + "/files/" + uuid.NewString() + ".svg"
	expires := time.Now().Add(time.Minute)
	file, err := repo.CreateConfirmed(ctx, productfiles.ProductFile{ProductID: product.ID, FileKey: key, FileName: "pig.svg", DisplayName: "Pig", Format: "svg", ContentType: "image/svg+xml", FileSizeBytes: 123, AvailabilityStatus: "active", ConfirmationFingerprint: strings.Repeat("a", 64), FingerprintKeyVersion: 1}, productfiles.Consumption{FileKey: key, ProductID: product.ID, TokenDigest: strings.Repeat("c", 64), ConfirmationFingerprint: strings.Repeat("a", 64), FingerprintKeyVersion: 1, ExpiresAt: &expires})
	if err != nil {
		t.Fatal(err)
	}
	secondKey := "products/" + product.ID + "/files/" + uuid.NewString() + ".pdf"
	second, err := repo.CreateConfirmed(ctx, productfiles.ProductFile{ProductID: product.ID, FileKey: secondKey, FileName: "pig.pdf", DisplayName: "PDF", Format: "pdf", ContentType: "application/pdf", FileSizeBytes: 12, AvailabilityStatus: "active", ConfirmationFingerprint: strings.Repeat("b", 64), FingerprintKeyVersion: 1, DisplayOrder: 0}, productfiles.Consumption{FileKey: secondKey, ProductID: product.ID, TokenDigest: strings.Repeat("d", 64), ConfirmationFingerprint: strings.Repeat("b", 64), FingerprintKeyVersion: 1, ExpiresAt: &expires})
	if err != nil {
		t.Fatal(err)
	}
	rollbackKey := "products/" + product.ID + "/files/" + uuid.NewString() + ".svg"
	_, err = repo.CreateConfirmed(ctx, productfiles.ProductFile{ProductID: product.ID, FileKey: rollbackKey,
		FileName: "rollback.svg", DisplayName: "Rollback", Format: "svg", ContentType: "image/svg+xml",
		FileSizeBytes: 123, AvailabilityStatus: "active", ConfirmationFingerprint: strings.Repeat("e", 64),
		FingerprintKeyVersion: 1, DisplayOrder: 0}, productfiles.Consumption{FileKey: rollbackKey,
		ProductID: product.ID, TokenDigest: strings.Repeat("c", 64),
		ConfirmationFingerprint: strings.Repeat("e", 64), FingerprintKeyVersion: 1, ExpiresAt: &expires})
	if !errors.Is(err, productfiles.ErrConflict) {
		t.Fatalf("duplicate token digest did not roll back: %v", err)
	}
	if _, err := repo.GetByKey(ctx, rollbackKey); !errors.Is(err, productfiles.ErrFileNotFound) {
		t.Fatalf("rolled-back file remained visible: %v", err)
	}
	untrackedKey := "products/" + product.ID + "/files/" + uuid.NewString() + ".svg"
	c, err := provider.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var untrackedID string
	err = c.QueryRow(ctx, `INSERT INTO product_files
	 (product_id,file_key,file_name,display_name,format,content_type,file_size_bytes,
	  confirmation_fingerprint,fingerprint_key_version,display_order)
	 VALUES ($1,$2,'untracked.svg','Untracked','svg','image/svg+xml',123,$3,1,2)
	 RETURNING id::text`, product.ID, untrackedKey, strings.Repeat("f", 64)).Scan(&untrackedID)
	_ = c.Release(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteByIDAndProduct(ctx, product.ID, untrackedID); !errors.Is(err, productfiles.ErrConflict) {
		t.Fatalf("delete removed a file without a consumption: %v", err)
	}
	if _, err := repo.GetByKey(ctx, untrackedKey); err != nil {
		t.Fatalf("untracked file was removed: %v", err)
	}
	c, err = provider.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Exec(ctx, `DELETE FROM product_files WHERE id=$1`, untrackedID)
	_ = c.Release(ctx)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := repo.List(ctx, product.ID)
	if err != nil || len(listed) != 2 || listed[0].ID != second.ID || listed[1].ID != file.ID || listed[1].DisplayOrder != 1 {
		t.Fatalf("list: %v %+v", err, listed)
	}
	byKey, err := repo.GetByKey(ctx, key)
	if err != nil || byKey.ID != file.ID {
		t.Fatalf("lookup: %v %+v", err, byKey)
	}
	updated, err := repo.UpdateMetadata(ctx, productfiles.UpdateMetadataInput{ProductID: product.ID, FileID: file.ID, DisplayName: "New name", DisplayOrder: 2})
	if err != nil || updated.DisplayName != "New name" || updated.DisplayOrder != 1 {
		t.Fatalf("update: %v %+v", err, updated)
	}
	updated, err = repo.UpdateMetadata(ctx, productfiles.UpdateMetadataInput{ProductID: product.ID, FileID: file.ID, DisplayName: "New name", DisplayOrder: 0})
	if err != nil || updated.DisplayOrder != 0 {
		t.Fatalf("move: %v %+v", err, updated)
	}
	if err := repo.Reorder(ctx, product.ID, []string{second.ID, file.ID}); err != nil {
		t.Fatal(err)
	}
	listed, err = repo.List(ctx, product.ID)
	if err != nil || listed[0].ID != second.ID || listed[0].DisplayOrder != 0 || listed[1].DisplayOrder != 1 {
		t.Fatalf("order: %v %+v", err, listed)
	}
	if err := repo.Reorder(ctx, product.ID, []string{}); !errors.Is(err, productfiles.ErrInvalidInput) {
		t.Fatalf("invalid order: %v", err)
	}
	if _, err := repo.GetByIDAndProduct(ctx, uuid.NewString(), file.ID); !errors.Is(err, productfiles.ErrFileNotFound) {
		t.Fatalf("wrong product lookup: %v", err)
	}
	if err := repo.DeleteByIDAndProduct(ctx, product.ID, file.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByKey(ctx, key); !errors.Is(err, productfiles.ErrFileNotFound) {
		t.Fatalf("deleted lookup: %v", err)
	}
	if _, err := repo.GetConsumptionByKey(ctx, key); err != nil {
		t.Fatalf("consumption disappeared after delete: %v", err)
	}
	if _, err := repo.CreateConfirmed(ctx, productfiles.ProductFile{ProductID: product.ID, FileKey: key,
		FileName: "pig.svg", DisplayName: "Pig", Format: "svg", ContentType: "image/svg+xml",
		FileSizeBytes: 123, AvailabilityStatus: "active", ConfirmationFingerprint: strings.Repeat("a", 64),
		FingerprintKeyVersion: 1}, productfiles.Consumption{FileKey: key, ProductID: product.ID,
		TokenDigest: strings.Repeat("e", 64), ConfirmationFingerprint: strings.Repeat("a", 64),
		FingerprintKeyVersion: 1, ExpiresAt: &expires}); !errors.Is(err, productfiles.ErrConflict) {
		t.Fatalf("consumed key was reused: %v", err)
	}
	listed, err = repo.List(ctx, product.ID)
	if err != nil || len(listed) != 1 || listed[0].ID != second.ID || listed[0].DisplayOrder != 0 {
		t.Fatalf("order after delete: %v %+v", err, listed)
	}
	if err := productService.DeleteProduct(ctx, product.ID); err != nil {
		t.Fatal(err)
	}
	deletedProduct = true
	if _, err := repo.GetConsumptionByKey(ctx, secondKey); err != nil {
		t.Fatalf("consumption disappeared after product deletion: %v", err)
	}
}
