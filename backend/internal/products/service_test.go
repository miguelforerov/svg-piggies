package products

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

func TestGetProductsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	got, err := service.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("GetProducts() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetProducts() = %#v, want empty non-nil slice", got)
	}
}

func TestGetProductTrimsID(t *testing.T) {
	t.Parallel()

	want := Product{ID: "product-id", Title: "Party Pig", Slug: "party-pig"}
	repository := &repositoryStub{
		get: func(_ context.Context, id string) (Product, error) {
			if id != want.ID {
				t.Fatalf("Get() id = %q, want %q", id, want.ID)
			}
			return want, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.GetProduct(context.Background(), " product-id ")
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProduct() = %#v, want %#v", got, want)
	}
}

func TestCreateProductValidatesAndDefaultsStatus(t *testing.T) {
	t.Parallel()

	wantInput := CreateProductInput{
		Title:       "Party Pig",
		Slug:        "party-pig",
		Description: "Printable party pig",
		Price:       "19.95",
		Status:      StatusDraft,
	}
	repository := &repositoryStub{
		create: func(_ context.Context, input CreateProductInput) (Product, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create() input = %#v, want %#v", input, wantInput)
			}
			return Product{
				ID:     "product-id",
				Title:  input.Title,
				Slug:   input.Slug,
				Price:  input.Price,
				Status: input.Status,
			}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.CreateProduct(context.Background(), CreateProductInput{
		Title:       " Party Pig ",
		Slug:        " party-pig ",
		Description: " Printable party pig ",
		Price:       " 19.95 ",
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	if got.Status != StatusDraft {
		t.Errorf("CreateProduct() status = %q, want %q", got.Status, StatusDraft)
	}
}

func TestUpdateProduct(t *testing.T) {
	t.Parallel()

	wantInput := UpdateProductInput{
		Title:  "Celebration Pig",
		Slug:   "celebration-pig",
		Price:  "24.50",
		Status: StatusActive,
	}
	repository := &repositoryStub{
		update: func(_ context.Context, id string, input UpdateProductInput) (Product, error) {
			if id != "product-id" {
				t.Fatalf("Update() id = %q", id)
			}
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Update() input = %#v, want %#v", input, wantInput)
			}
			return Product{ID: id, Title: input.Title, Status: input.Status}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.UpdateProduct(context.Background(), " product-id ", UpdateProductInput{
		Title:  " Celebration Pig ",
		Slug:   " celebration-pig ",
		Price:  " 24.50 ",
		Status: StatusActive,
	})
	if err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}
	if got.Status != StatusActive {
		t.Errorf("UpdateProduct() status = %q, want %q", got.Status, StatusActive)
	}
}

func TestDeleteProduct(t *testing.T) {
	t.Parallel()

	called := false
	repository := &repositoryStub{
		delete: func(_ context.Context, id string) error {
			called = true
			if id != "product-id" {
				t.Fatalf("Delete() id = %q", id)
			}
			return nil
		},
	}
	service := newTestService(t, repository)

	if err := service.DeleteProduct(context.Background(), " product-id "); err != nil {
		t.Fatalf("DeleteProduct() error = %v", err)
	}
	if !called {
		t.Fatal("Delete() was not called")
	}
}

func TestProductValidationErrors(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	validCreate := CreateProductInput{
		Title:  "Party Pig",
		Slug:   "party-pig",
		Price:  "19.95",
		Status: StatusDraft,
	}
	validUpdate := UpdateProductInput{
		Title:  "Party Pig",
		Slug:   "party-pig",
		Price:  "19.95",
		Status: StatusDraft,
	}

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "missing id",
			run: func() error {
				_, err := service.GetProduct(context.Background(), " ")
				return err
			},
		},
		{
			name: "missing title",
			run: func() error {
				input := validCreate
				input.Title = " "
				_, err := service.CreateProduct(context.Background(), input)
				return err
			},
		},
		{
			name: "missing slug",
			run: func() error {
				input := validCreate
				input.Slug = ""
				_, err := service.CreateProduct(context.Background(), input)
				return err
			},
		},
		{
			name: "negative price",
			run: func() error {
				input := validCreate
				input.Price = "-1.00"
				_, err := service.CreateProduct(context.Background(), input)
				return err
			},
		},
		{
			name: "too many decimal places",
			run: func() error {
				input := validCreate
				input.Price = "1.999"
				_, err := service.CreateProduct(context.Background(), input)
				return err
			},
		},
		{
			name: "price exceeds database precision",
			run: func() error {
				input := validCreate
				input.Price = "10000000000.00"
				_, err := service.CreateProduct(context.Background(), input)
				return err
			},
		},
		{
			name: "invalid create status",
			run: func() error {
				input := validCreate
				input.Status = "published"
				_, err := service.CreateProduct(context.Background(), input)
				return err
			},
		},
		{
			name: "missing update status",
			run: func() error {
				input := validUpdate
				input.Status = ""
				_, err := service.UpdateProduct(context.Background(), "product-id", input)
				return err
			},
		},
		{
			name: "missing delete id",
			run: func() error {
				return service.DeleteProduct(context.Background(), "")
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
		get: func(context.Context, string) (Product, error) {
			return Product{}, ErrNotFound
		},
	}
	service := newTestService(t, repository)

	_, err := service.GetProduct(context.Background(), "product-id")
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
	list   func(context.Context) ([]Product, error)
	get    func(context.Context, string) (Product, error)
	create func(context.Context, CreateProductInput) (Product, error)
	update func(context.Context, string, UpdateProductInput) (Product, error)
	delete func(context.Context, string) error
}

func (r *repositoryStub) List(ctx context.Context) ([]Product, error) {
	if r.list == nil {
		return nil, nil
	}
	return r.list(ctx)
}

func (r *repositoryStub) Get(ctx context.Context, id string) (Product, error) {
	if r.get == nil {
		return Product{}, nil
	}
	return r.get(ctx, id)
}

func (r *repositoryStub) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	if r.create == nil {
		return Product{}, nil
	}
	return r.create(ctx, input)
}

func (r *repositoryStub) Update(
	ctx context.Context,
	id string,
	input UpdateProductInput,
) (Product, error) {
	if r.update == nil {
		return Product{}, nil
	}
	return r.update(ctx, id, input)
}

func (r *repositoryStub) Delete(ctx context.Context, id string) error {
	if r.delete == nil {
		return nil
	}
	return r.delete(ctx, id)
}
