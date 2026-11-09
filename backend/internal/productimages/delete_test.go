package productimages

import (
	"context"
	"errors"
	"testing"

	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

type deleteImageRepoStub struct {
	image     ProductImage
	lookupErr error
	deleteErr error
	lookups   int
	deletes   int
}

func (s *deleteImageRepoStub) GetByIDAndProduct(_ context.Context, productID, imageID string) (ProductImage, error) {
	s.lookups++
	if s.lookupErr != nil {
		return ProductImage{}, s.lookupErr
	}
	if s.image.ProductID != productID || s.image.ID != imageID {
		return ProductImage{}, ErrImageNotFound
	}
	return s.image, nil
}

func (s *deleteImageRepoStub) DeleteByIDAndProduct(_ context.Context, productID, imageID string) error {
	s.deletes++
	if productID != s.image.ProductID || imageID != s.image.ID {
		return ErrImageNotFound
	}
	return s.deleteErr
}

type deleteObjectStub struct {
	keys    []string
	buckets []storage.Bucket
	err     error
}

func (s *deleteObjectStub) DeleteObject(_ context.Context, bucket storage.Bucket, key string) error {
	s.buckets = append(s.buckets, bucket)
	s.keys = append(s.keys, key)
	return s.err
}

func TestDeleteProductImage(t *testing.T) {
	const productID, imageID, key = "product-1", "image-1", "products/product-1/images/server-key.webp"
	for _, tc := range []struct {
		name       string
		productErr error
		image      ProductImage
		lookupErr  error
		storageErr error
		deleteErr  error
		wantErr    error
		wantStore  int
		wantDelete int
	}{
		{name: "success", image: ProductImage{ID: imageID, ProductID: productID, R2ObjectKey: key}, wantStore: 1, wantDelete: 1},
		{name: "missing product", productErr: products.ErrNotFound, wantErr: ErrReferenceNotFound},
		{name: "missing image", lookupErr: ErrImageNotFound, wantErr: ErrImageNotFound},
		{name: "other product image", image: ProductImage{ID: imageID, ProductID: "another-product", R2ObjectKey: key}, wantErr: ErrImageNotFound},
		{name: "R2 failure", image: ProductImage{ID: imageID, ProductID: productID, R2ObjectKey: key}, storageErr: errors.New("R2 unavailable"), wantErr: ErrStorageFailure, wantStore: 1},
		{name: "object already absent", image: ProductImage{ID: imageID, ProductID: productID, R2ObjectKey: key}, wantStore: 1, wantDelete: 1},
		{name: "database fails after R2", image: ProductImage{ID: imageID, ProductID: productID, R2ObjectKey: key}, deleteErr: errors.New("database unavailable"), wantStore: 1, wantDelete: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &deleteImageRepoStub{image: tc.image, lookupErr: tc.lookupErr, deleteErr: tc.deleteErr}
			store := &deleteObjectStub{err: tc.storageErr}
			service, err := NewDeleteService(&uploadProductStub{err: tc.productErr}, repo, store)
			if err != nil {
				t.Fatal(err)
			}
			err = service.DeleteProductImage(context.Background(), productID, imageID)
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.deleteErr != nil && !errors.Is(err, tc.deleteErr) {
				t.Fatalf("database error = %v", err)
			}
			if tc.wantErr == nil && tc.deleteErr == nil && err != nil {
				t.Fatal(err)
			}
			if len(store.keys) != tc.wantStore || repo.deletes != tc.wantDelete {
				t.Fatalf("R2 calls = %d; DB deletes = %d", len(store.keys), repo.deletes)
			}
			if tc.wantStore == 1 && (store.keys[0] != key || store.buckets[0] != storage.AssetsBucket) {
				t.Fatalf("deleted object = %q in bucket %q", store.keys[0], store.buckets[0])
			}
		})
	}
}

func TestDeleteProductImageCanRetryAfterDatabaseFailure(t *testing.T) {
	repo := &deleteImageRepoStub{image: ProductImage{ID: "image", ProductID: "product", R2ObjectKey: "server-key"}, deleteErr: errors.New("database unavailable")}
	store := &deleteObjectStub{}
	service, err := NewDeleteService(&uploadProductStub{}, repo, store)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteProductImage(context.Background(), "product", "image"); err == nil {
		t.Fatal("expected database error")
	}
	repo.deleteErr = nil
	if err := service.DeleteProductImage(context.Background(), "product", "image"); err != nil {
		t.Fatal(err)
	}
	if len(store.keys) != 2 || repo.deletes != 2 {
		t.Fatalf("retry did not repeat deletion: R2=%d DB=%d", len(store.keys), repo.deletes)
	}
}
