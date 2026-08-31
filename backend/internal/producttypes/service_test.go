package producttypes

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

func TestGetProductTypesReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	got, err := service.GetProductTypes(context.Background())
	if err != nil {
		t.Fatalf("GetProductTypes() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetProductTypes() = %#v, want empty non-nil slice", got)
	}
}

func TestGetProductType(t *testing.T) {
	t.Parallel()

	want := ProductType{ID: "product-type-id", Name: "Animals", Slug: "animals"}
	repository := &repositoryStub{
		get: func(_ context.Context, id string) (ProductType, error) {
			if id != want.ID {
				t.Fatalf("Get() id = %q, want %q", id, want.ID)
			}
			return want, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.GetProductType(context.Background(), "  product-type-id  ")
	if err != nil {
		t.Fatalf("GetProductType() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProductType() = %#v, want %#v", got, want)
	}
}

func TestCreateProductTypeValidatesAndTrimsInput(t *testing.T) {
	t.Parallel()

	wantInput := CreateProductTypeInput{
		Name:        "Animals",
		Slug:        "animals",
		Description: "Animal illustrations",
	}
	repository := &repositoryStub{
		create: func(_ context.Context, input CreateProductTypeInput) (ProductType, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create() input = %#v, want %#v", input, wantInput)
			}
			return ProductType{ID: "product-type-id", Name: input.Name, Slug: input.Slug}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.CreateProductType(context.Background(), CreateProductTypeInput{
		Name:        " Animals ",
		Slug:        " animals ",
		Description: " Animal illustrations ",
	})
	if err != nil {
		t.Fatalf("CreateProductType() error = %v", err)
	}
	if got.ID != "product-type-id" {
		t.Errorf("CreateProductType() ID = %q", got.ID)
	}
}

func TestUpdateProductType(t *testing.T) {
	t.Parallel()

	wantInput := UpdateProductTypeInput{Name: "Wildlife", Slug: "wildlife"}
	repository := &repositoryStub{
		update: func(_ context.Context, id string, input UpdateProductTypeInput) (ProductType, error) {
			if id != "product-type-id" {
				t.Fatalf("Update() id = %q", id)
			}
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Update() input = %#v, want %#v", input, wantInput)
			}
			return ProductType{ID: id, Name: input.Name, Slug: input.Slug}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.UpdateProductType(
		context.Background(),
		" product-type-id ",
		UpdateProductTypeInput{Name: " Wildlife ", Slug: " wildlife "},
	)
	if err != nil {
		t.Fatalf("UpdateProductType() error = %v", err)
	}
	if got.Name != "Wildlife" {
		t.Errorf("UpdateProductType() Name = %q", got.Name)
	}
}

func TestDeleteProductType(t *testing.T) {
	t.Parallel()

	called := false
	repository := &repositoryStub{
		delete: func(_ context.Context, id string) error {
			called = true
			if id != "product-type-id" {
				t.Fatalf("Delete() id = %q", id)
			}
			return nil
		},
	}
	service := newTestService(t, repository)

	if err := service.DeleteProductType(context.Background(), " product-type-id "); err != nil {
		t.Fatalf("DeleteProductType() error = %v", err)
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
			name: "missing id",
			run: func() error {
				_, err := service.GetProductType(context.Background(), " ")
				return err
			},
		},
		{
			name: "missing create name",
			run: func() error {
				_, err := service.CreateProductType(
					context.Background(),
					CreateProductTypeInput{Slug: "animals"},
				)
				return err
			},
		},
		{
			name: "missing update slug",
			run: func() error {
				_, err := service.UpdateProductType(
					context.Background(),
					"product-type-id",
					UpdateProductTypeInput{Name: "Animals"},
				)
				return err
			},
		},
		{
			name: "missing delete id",
			run: func() error {
				return service.DeleteProductType(context.Background(), "")
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
		get: func(context.Context, string) (ProductType, error) {
			return ProductType{}, ErrNotFound
		},
	}
	service := newTestService(t, repository)

	_, err := service.GetProductType(context.Background(), "product-type-id")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
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
	list   func(context.Context) ([]ProductType, error)
	get    func(context.Context, string) (ProductType, error)
	create func(context.Context, CreateProductTypeInput) (ProductType, error)
	update func(context.Context, string, UpdateProductTypeInput) (ProductType, error)
	delete func(context.Context, string) error
}

func (r *repositoryStub) List(ctx context.Context) ([]ProductType, error) {
	if r.list == nil {
		return nil, nil
	}
	return r.list(ctx)
}

func (r *repositoryStub) Get(ctx context.Context, id string) (ProductType, error) {
	if r.get == nil {
		return ProductType{}, nil
	}
	return r.get(ctx, id)
}

func (r *repositoryStub) Create(
	ctx context.Context,
	input CreateProductTypeInput,
) (ProductType, error) {
	if r.create == nil {
		return ProductType{}, nil
	}
	return r.create(ctx, input)
}

func (r *repositoryStub) Update(
	ctx context.Context,
	id string,
	input UpdateProductTypeInput,
) (ProductType, error) {
	if r.update == nil {
		return ProductType{}, nil
	}
	return r.update(ctx, id, input)
}

func (r *repositoryStub) Delete(ctx context.Context, id string) error {
	if r.delete == nil {
		return nil
	}
	return r.delete(ctx, id)
}
