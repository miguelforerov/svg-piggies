package productcollections

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestNewServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	if _, err := NewService(nil); err == nil {
		t.Fatal("NewService(nil) error = nil, want error")
	}
}

func TestGetProductCollections(t *testing.T) {
	t.Parallel()

	want := []ProductCollection{{ProductID: "product-id", CollectionID: "collection-id"}}
	repository := &repositoryStub{
		list: func(_ context.Context, productID string) ([]ProductCollection, error) {
			if productID != "product-id" {
				t.Fatalf("ListByProduct() productID = %q", productID)
			}
			return want, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.GetProductCollections(context.Background(), " product-id ")
	if err != nil {
		t.Fatalf("GetProductCollections() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProductCollections() = %#v, want %#v", got, want)
	}
}

func TestGetProductCollectionsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	got, err := service.GetProductCollections(context.Background(), "product-id")
	if err != nil {
		t.Fatalf("GetProductCollections() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetProductCollections() = %#v, want empty non-nil slice", got)
	}
}

func TestCreateProductCollectionTrimsIDs(t *testing.T) {
	t.Parallel()

	want := CreateProductCollectionInput{
		ProductID:    "product-id",
		CollectionID: "collection-id",
	}
	repository := &repositoryStub{
		create: func(_ context.Context, input CreateProductCollectionInput) (ProductCollection, error) {
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("Create() input = %#v, want %#v", input, want)
			}
			return ProductCollection(input), nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.CreateProductCollection(context.Background(), CreateProductCollectionInput{
		ProductID:    " product-id ",
		CollectionID: " collection-id ",
	})
	if err != nil {
		t.Fatalf("CreateProductCollection() error = %v", err)
	}
	if got.ProductID != want.ProductID || got.CollectionID != want.CollectionID {
		t.Fatalf("CreateProductCollection() = %#v, want %#v", got, want)
	}
}

func TestDeleteProductCollectionTrimsIDs(t *testing.T) {
	t.Parallel()

	called := false
	repository := &repositoryStub{
		delete: func(_ context.Context, productID string, collectionID string) error {
			called = true
			if productID != "product-id" || collectionID != "collection-id" {
				t.Fatalf("Delete() IDs = %q, %q", productID, collectionID)
			}
			return nil
		},
	}
	service := newTestService(t, repository)

	err := service.DeleteProductCollection(
		context.Background(),
		" product-id ",
		" collection-id ",
	)
	if err != nil {
		t.Fatalf("DeleteProductCollection() error = %v", err)
	}
	if !called {
		t.Fatal("Delete() was not called")
	}
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "list without product id",
			run: func() error {
				_, err := service.GetProductCollections(context.Background(), " ")
				return err
			},
		},
		{
			name: "create without product id",
			run: func() error {
				_, err := service.CreateProductCollection(
					context.Background(),
					CreateProductCollectionInput{CollectionID: "collection-id"},
				)
				return err
			},
		},
		{
			name: "create without collection id",
			run: func() error {
				_, err := service.CreateProductCollection(
					context.Background(),
					CreateProductCollectionInput{ProductID: "product-id"},
				)
				return err
			},
		},
		{
			name: "delete without collection id",
			run: func() error {
				return service.DeleteProductCollection(context.Background(), "product-id", "")
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.run(); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestRepositoryErrorsRemainInspectable(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{
		create: func(context.Context, CreateProductCollectionInput) (ProductCollection, error) {
			return ProductCollection{}, ErrConflict
		},
	}
	service := newTestService(t, repository)

	_, err := service.CreateProductCollection(context.Background(), CreateProductCollectionInput{
		ProductID:    "product-id",
		CollectionID: "collection-id",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func newTestService(t *testing.T, repository Repository) *Service {
	t.Helper()

	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

type repositoryStub struct {
	list   func(context.Context, string) ([]ProductCollection, error)
	create func(context.Context, CreateProductCollectionInput) (ProductCollection, error)
	delete func(context.Context, string, string) error
}

func (r *repositoryStub) ListByProduct(
	ctx context.Context,
	productID string,
) ([]ProductCollection, error) {
	if r.list == nil {
		return nil, nil
	}
	return r.list(ctx, productID)
}

func (r *repositoryStub) Create(
	ctx context.Context,
	input CreateProductCollectionInput,
) (ProductCollection, error) {
	if r.create == nil {
		return ProductCollection{}, nil
	}
	return r.create(ctx, input)
}

func (r *repositoryStub) Delete(
	ctx context.Context,
	productID string,
	collectionID string,
) error {
	if r.delete == nil {
		return nil
	}
	return r.delete(ctx, productID, collectionID)
}
