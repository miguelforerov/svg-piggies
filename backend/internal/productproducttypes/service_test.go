package productproducttypes

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

func TestGetProductProductTypes(t *testing.T) {
	t.Parallel()

	want := []ProductProductType{{ProductID: "product-id", ProductTypeID: "product-type-id"}}
	repository := &repositoryStub{
		list: func(_ context.Context, productID string) ([]ProductProductType, error) {
			if productID != "product-id" {
				t.Fatalf("ListByProduct() productID = %q", productID)
			}
			return want, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.GetProductProductTypes(context.Background(), " product-id ")
	if err != nil {
		t.Fatalf("GetProductProductTypes() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProductProductTypes() = %#v, want %#v", got, want)
	}
}

func TestGetProductProductTypesReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	got, err := service.GetProductProductTypes(context.Background(), "product-id")
	if err != nil {
		t.Fatalf("GetProductProductTypes() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetProductProductTypes() = %#v, want empty non-nil slice", got)
	}
}

func TestCreateProductProductTypeTrimsIDs(t *testing.T) {
	t.Parallel()

	want := CreateProductProductTypeInput{
		ProductID:     "product-id",
		ProductTypeID: "product-type-id",
	}
	repository := &repositoryStub{
		create: func(
			_ context.Context,
			input CreateProductProductTypeInput,
		) (ProductProductType, error) {
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("Create() input = %#v, want %#v", input, want)
			}
			return ProductProductType(input), nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.CreateProductProductType(context.Background(), CreateProductProductTypeInput{
		ProductID:     " product-id ",
		ProductTypeID: " product-type-id ",
	})
	if err != nil {
		t.Fatalf("CreateProductProductType() error = %v", err)
	}
	if got.ProductID != want.ProductID || got.ProductTypeID != want.ProductTypeID {
		t.Fatalf("CreateProductProductType() = %#v, want %#v", got, want)
	}
}

func TestDeleteProductProductTypeTrimsIDs(t *testing.T) {
	t.Parallel()

	called := false
	repository := &repositoryStub{
		delete: func(_ context.Context, productID string, productTypeID string) error {
			called = true
			if productID != "product-id" || productTypeID != "product-type-id" {
				t.Fatalf("Delete() IDs = %q, %q", productID, productTypeID)
			}
			return nil
		},
	}
	service := newTestService(t, repository)

	err := service.DeleteProductProductType(
		context.Background(),
		" product-id ",
		" product-type-id ",
	)
	if err != nil {
		t.Fatalf("DeleteProductProductType() error = %v", err)
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
				_, err := service.GetProductProductTypes(context.Background(), " ")
				return err
			},
		},
		{
			name: "create without product id",
			run: func() error {
				_, err := service.CreateProductProductType(
					context.Background(),
					CreateProductProductTypeInput{ProductTypeID: "product-type-id"},
				)
				return err
			},
		},
		{
			name: "create without product type id",
			run: func() error {
				_, err := service.CreateProductProductType(
					context.Background(),
					CreateProductProductTypeInput{ProductID: "product-id"},
				)
				return err
			},
		},
		{
			name: "delete without product type id",
			run: func() error {
				return service.DeleteProductProductType(context.Background(), "product-id", "")
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
		create: func(
			context.Context,
			CreateProductProductTypeInput,
		) (ProductProductType, error) {
			return ProductProductType{}, ErrConflict
		},
	}
	service := newTestService(t, repository)

	_, err := service.CreateProductProductType(context.Background(), CreateProductProductTypeInput{
		ProductID:     "product-id",
		ProductTypeID: "product-type-id",
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
	list   func(context.Context, string) ([]ProductProductType, error)
	create func(context.Context, CreateProductProductTypeInput) (ProductProductType, error)
	delete func(context.Context, string, string) error
}

func (r *repositoryStub) ListByProduct(
	ctx context.Context,
	productID string,
) ([]ProductProductType, error) {
	if r.list == nil {
		return nil, nil
	}
	return r.list(ctx, productID)
}

func (r *repositoryStub) Create(
	ctx context.Context,
	input CreateProductProductTypeInput,
) (ProductProductType, error) {
	if r.create == nil {
		return ProductProductType{}, nil
	}
	return r.create(ctx, input)
}

func (r *repositoryStub) Delete(
	ctx context.Context,
	productID string,
	productTypeID string,
) error {
	if r.delete == nil {
		return nil
	}
	return r.delete(ctx, productID, productTypeID)
}
