package productimages

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zdenaforero/svg-piggies/backend/internal/products"
	"github.com/zdenaforero/svg-piggies/backend/internal/storage"
)

type imageReaderStub struct {
	images []ProductImage
	err    error
	called bool
}

func (s *imageReaderStub) GetByProductID(_ context.Context, _ string) ([]ProductImage, error) {
	s.called = true
	return s.images, s.err
}

type imageSignerStub struct {
	keys    []string
	buckets []storage.Bucket
	ttls    []time.Duration
	err     error
}

func (s *imageSignerStub) PresignGet(_ context.Context, bucket storage.Bucket, key string, ttl time.Duration) (storage.PresignedRequest, error) {
	s.keys = append(s.keys, key)
	s.buckets = append(s.buckets, bucket)
	s.ttls = append(s.ttls, ttl)
	return storage.PresignedRequest{URL: "https://example.test/" + key}, s.err
}

func TestGetProductImages(t *testing.T) {
	images := &imageReaderStub{images: []ProductImage{
		{ID: "first", R2ObjectKey: "first.webp", IsPrimary: true},
		{ID: "second", R2ObjectKey: "second.webp"},
	}}
	signer := &imageSignerStub{}
	service, err := NewReadService(&uploadProductStub{}, images, signer)
	if err != nil {
		t.Fatal(err)
	}
	got, err := service.GetProductImages(context.Background(), uploadProductID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "first" || got[1].ID != "second" || got[0].ImageURL != "https://example.test/first.webp" {
		t.Fatalf("images = %+v", got)
	}
	for i, key := range []string{"first.webp", "second.webp"} {
		if signer.keys[i] != key || signer.buckets[i] != storage.AssetsBucket || signer.ttls[i] != 5*time.Minute {
			t.Fatalf("sign call %d: key=%q bucket=%q ttl=%s", i, signer.keys[i], signer.buckets[i], signer.ttls[i])
		}
	}
}

func TestGetProductImagesEmptyAndErrors(t *testing.T) {
	storageErr := errors.New("secret storage detail")
	for _, tt := range []struct {
		name       string
		productErr error
		readerErr  error
		signerErr  error
		want       error
	}{
		{name: "empty"},
		{name: "missing product", productErr: products.ErrNotFound, want: ErrReferenceNotFound},
		{name: "reader", readerErr: errors.New("database failure")},
		{name: "storage", signerErr: storageErr, want: ErrStorageFailure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reader := &imageReaderStub{err: tt.readerErr}
			if tt.signerErr != nil {
				reader.images = []ProductImage{{R2ObjectKey: "key"}}
			}
			signer := &imageSignerStub{err: tt.signerErr}
			service, err := NewReadService(&uploadProductStub{err: tt.productErr}, reader, signer)
			if err != nil {
				t.Fatal(err)
			}
			got, err := service.GetProductImages(context.Background(), uploadProductID)
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if tt.readerErr != nil && !errors.Is(err, tt.readerErr) {
				t.Fatalf("error = %v", err)
			}
			if tt.want == nil && tt.readerErr == nil && (err != nil || got == nil || len(got) != 0) {
				t.Fatalf("images = %+v, error = %v", got, err)
			}
			if tt.productErr != nil && reader.called {
				t.Fatal("reader called for missing product")
			}
		})
	}
}

func TestNewReadServiceRequiresDependencies(t *testing.T) {
	product := &uploadProductStub{}
	reader := &imageReaderStub{}
	signer := &imageSignerStub{}
	for _, deps := range []struct {
		product ProductLookup
		reader  ProductImageReader
		signer  ImageURLSigner
	}{
		{reader: reader, signer: signer},
		{product: product, signer: signer},
		{product: product, reader: reader},
	} {
		if _, err := NewReadService(deps.product, deps.reader, deps.signer); err == nil {
			t.Fatal("expected missing dependency error")
		}
	}
}

func TestGetProductImagesRejectsEmptyProductID(t *testing.T) {
	reader := &imageReaderStub{}
	service, err := NewReadService(&uploadProductStub{}, reader, &imageSignerStub{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.GetProductImages(context.Background(), "  ")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
	if reader.called {
		t.Fatal("reader called for empty product ID")
	}
}
