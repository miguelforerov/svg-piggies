package productrelationships

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

func TestGetProductRelationships(t *testing.T) {
	t.Parallel()

	want := []ProductRelationship{{
		ID:               "relationship-id",
		ProductID:        "product-id",
		RelatedProductID: "related-product-id",
		DisplayOrder:     2,
	}}
	repository := &repositoryStub{
		list: func(_ context.Context, productID string) ([]ProductRelationship, error) {
			if productID != "product-id" {
				t.Fatalf("ListByProduct() productID = %q", productID)
			}
			return want, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.GetProductRelationships(context.Background(), " product-id ")
	if err != nil {
		t.Fatalf("GetProductRelationships() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProductRelationships() = %#v, want %#v", got, want)
	}
}

func TestGetProductRelationshipsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	got, err := service.GetProductRelationships(context.Background(), "product-id")
	if err != nil {
		t.Fatalf("GetProductRelationships() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetProductRelationships() = %#v, want empty non-nil slice", got)
	}
}

func TestGetProductRelationshipTrimsIDs(t *testing.T) {
	t.Parallel()

	want := ProductRelationship{ID: "relationship-id", ProductID: "product-id"}
	repository := &repositoryStub{
		get: func(_ context.Context, productID string, relationshipID string) (ProductRelationship, error) {
			if productID != "product-id" || relationshipID != "relationship-id" {
				t.Fatalf("Get() IDs = %q, %q", productID, relationshipID)
			}
			return want, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.GetProductRelationship(
		context.Background(),
		" product-id ",
		" relationship-id ",
	)
	if err != nil {
		t.Fatalf("GetProductRelationship() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProductRelationship() = %#v, want %#v", got, want)
	}
}

func TestCreateProductRelationshipValidatesAndTrimsInput(t *testing.T) {
	t.Parallel()

	wantInput := CreateProductRelationshipInput{
		ProductID:        "product-id",
		RelatedProductID: "related-product-id",
		DisplayOrder:     3,
	}
	repository := &repositoryStub{
		create: func(_ context.Context, input CreateProductRelationshipInput) (ProductRelationship, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create() input = %#v, want %#v", input, wantInput)
			}
			return ProductRelationship{
				ID: "relationship-id", ProductID: input.ProductID,
				RelatedProductID: input.RelatedProductID, DisplayOrder: input.DisplayOrder,
			}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.CreateProductRelationship(context.Background(), CreateProductRelationshipInput{
		ProductID:        " product-id ",
		RelatedProductID: " related-product-id ",
		DisplayOrder:     3,
	})
	if err != nil {
		t.Fatalf("CreateProductRelationship() error = %v", err)
	}
	if got.ID != "relationship-id" {
		t.Errorf("CreateProductRelationship() ID = %q", got.ID)
	}
}

func TestUpdateProductRelationship(t *testing.T) {
	t.Parallel()

	wantInput := UpdateProductRelationshipInput{
		RelatedProductID: "other-product-id",
		DisplayOrder:     4,
	}
	repository := &repositoryStub{
		update: func(
			_ context.Context,
			productID string,
			relationshipID string,
			input UpdateProductRelationshipInput,
		) (ProductRelationship, error) {
			if productID != "product-id" || relationshipID != "relationship-id" {
				t.Fatalf("Update() IDs = %q, %q", productID, relationshipID)
			}
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Update() input = %#v, want %#v", input, wantInput)
			}
			return ProductRelationship{ID: relationshipID, ProductID: productID}, nil
		},
	}
	service := newTestService(t, repository)

	_, err := service.UpdateProductRelationship(
		context.Background(),
		" product-id ",
		" relationship-id ",
		UpdateProductRelationshipInput{
			RelatedProductID: " other-product-id ",
			DisplayOrder:     4,
		},
	)
	if err != nil {
		t.Fatalf("UpdateProductRelationship() error = %v", err)
	}
}

func TestDeleteProductRelationship(t *testing.T) {
	t.Parallel()

	called := false
	repository := &repositoryStub{
		delete: func(_ context.Context, productID string, relationshipID string) error {
			called = true
			if productID != "product-id" || relationshipID != "relationship-id" {
				t.Fatalf("Delete() IDs = %q, %q", productID, relationshipID)
			}
			return nil
		},
	}
	service := newTestService(t, repository)

	err := service.DeleteProductRelationship(
		context.Background(),
		" product-id ",
		" relationship-id ",
	)
	if err != nil {
		t.Fatalf("DeleteProductRelationship() error = %v", err)
	}
	if !called {
		t.Fatal("Delete() was not called")
	}
}

func TestReplaceProductRelationshipsValidatesAndPreservesOrder(t *testing.T) {
	t.Parallel()

	wantInput := ReplaceProductRelationshipsInput{
		ProductID:         "product-id",
		RelatedProductIDs: []string{"related-product-2", "related-product-1"},
	}
	repository := &repositoryStub{
		replace: func(
			_ context.Context,
			input ReplaceProductRelationshipsInput,
		) (ProductWithRelationships, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Replace() input = %#v, want %#v", input, wantInput)
			}
			return ProductWithRelationships{}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.ReplaceProductRelationships(
		context.Background(),
		ReplaceProductRelationshipsInput{
			ProductID:         " product-id ",
			RelatedProductIDs: []string{" related-product-2 ", " related-product-1 "},
		},
	)
	if err != nil {
		t.Fatalf("ReplaceProductRelationships() error = %v", err)
	}
	if got.Relationships == nil {
		t.Fatal("ReplaceProductRelationships() relationships = nil, want empty slice")
	}
}

func TestCreateProductRelationshipsIgnoresRepeatedInputIDs(t *testing.T) {
	t.Parallel()

	wantInput := CreateProductRelationshipsInput{
		ProductID:         "product-id",
		RelatedProductIDs: []string{"related-product-1", "related-product-2"},
	}
	repository := &repositoryStub{
		createMany: func(
			_ context.Context,
			input CreateProductRelationshipsInput,
		) (ProductWithRelationships, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("CreateMany() input = %#v, want %#v", input, wantInput)
			}
			return ProductWithRelationships{}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.CreateProductRelationships(
		context.Background(),
		CreateProductRelationshipsInput{
			ProductID: " product-id ",
			RelatedProductIDs: []string{
				" related-product-1 ",
				"related-product-1",
				" related-product-2 ",
			},
		},
	)
	if err != nil {
		t.Fatalf("CreateProductRelationships() error = %v", err)
	}
	if got.Relationships == nil {
		t.Fatal("CreateProductRelationships() relationships = nil, want empty slice")
	}
}

func TestProductRelationshipValidationErrors(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "missing product id",
			run: func() error {
				_, err := service.GetProductRelationships(context.Background(), " ")
				return err
			},
		},
		{
			name: "missing relationship id",
			run: func() error {
				_, err := service.GetProductRelationship(context.Background(), "product-id", "")
				return err
			},
		},
		{
			name: "self relationship",
			run: func() error {
				_, err := service.CreateProductRelationship(
					context.Background(),
					CreateProductRelationshipInput{
						ProductID:        "product-id",
						RelatedProductID: "product-id",
					},
				)
				return err
			},
		},
		{
			name: "negative display order",
			run: func() error {
				_, err := service.CreateProductRelationship(
					context.Background(),
					CreateProductRelationshipInput{
						ProductID:        "product-id",
						RelatedProductID: "related-product-id",
						DisplayOrder:     -1,
					},
				)
				return err
			},
		},
		{
			name: "missing related product on update",
			run: func() error {
				_, err := service.UpdateProductRelationship(
					context.Background(),
					"product-id",
					"relationship-id",
					UpdateProductRelationshipInput{},
				)
				return err
			},
		},
		{
			name: "bulk self relationship",
			run: func() error {
				_, err := service.ReplaceProductRelationships(
					context.Background(),
					ReplaceProductRelationshipsInput{
						ProductID:         "product-id",
						RelatedProductIDs: []string{"product-id"},
					},
				)
				return err
			},
		},
		{
			name: "additive bulk requires a relationship",
			run: func() error {
				_, err := service.CreateProductRelationships(
					context.Background(),
					CreateProductRelationshipsInput{ProductID: "product-id"},
				)
				return err
			},
		},
		{
			name: "duplicate bulk relationship",
			run: func() error {
				_, err := service.ReplaceProductRelationships(
					context.Background(),
					ReplaceProductRelationshipsInput{
						ProductID: "product-id",
						RelatedProductIDs: []string{
							"related-product-id",
							"related-product-id",
						},
					},
				)
				return err
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
		create: func(context.Context, CreateProductRelationshipInput) (ProductRelationship, error) {
			return ProductRelationship{}, ErrConflict
		},
	}
	service := newTestService(t, repository)

	_, err := service.CreateProductRelationship(context.Background(), CreateProductRelationshipInput{
		ProductID:        "product-id",
		RelatedProductID: "related-product-id",
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
	list       func(context.Context, string) ([]ProductRelationship, error)
	get        func(context.Context, string, string) (ProductRelationship, error)
	create     func(context.Context, CreateProductRelationshipInput) (ProductRelationship, error)
	createMany func(context.Context, CreateProductRelationshipsInput) (ProductWithRelationships, error)
	update     func(context.Context, string, string, UpdateProductRelationshipInput) (ProductRelationship, error)
	replace    func(context.Context, ReplaceProductRelationshipsInput) (ProductWithRelationships, error)
	delete     func(context.Context, string, string) error
}

func (r *repositoryStub) ListByProduct(
	ctx context.Context,
	productID string,
) ([]ProductRelationship, error) {
	if r.list == nil {
		return nil, nil
	}
	return r.list(ctx, productID)
}

func (r *repositoryStub) Get(
	ctx context.Context,
	productID string,
	relationshipID string,
) (ProductRelationship, error) {
	if r.get == nil {
		return ProductRelationship{}, nil
	}
	return r.get(ctx, productID, relationshipID)
}

func (r *repositoryStub) Create(
	ctx context.Context,
	input CreateProductRelationshipInput,
) (ProductRelationship, error) {
	if r.create == nil {
		return ProductRelationship{}, nil
	}
	return r.create(ctx, input)
}

func (r *repositoryStub) CreateMany(
	ctx context.Context,
	input CreateProductRelationshipsInput,
) (ProductWithRelationships, error) {
	if r.createMany == nil {
		return ProductWithRelationships{}, nil
	}
	return r.createMany(ctx, input)
}

func (r *repositoryStub) Update(
	ctx context.Context,
	productID string,
	relationshipID string,
	input UpdateProductRelationshipInput,
) (ProductRelationship, error) {
	if r.update == nil {
		return ProductRelationship{}, nil
	}
	return r.update(ctx, productID, relationshipID, input)
}

func (r *repositoryStub) Replace(
	ctx context.Context,
	input ReplaceProductRelationshipsInput,
) (ProductWithRelationships, error) {
	if r.replace == nil {
		return ProductWithRelationships{}, nil
	}
	return r.replace(ctx, input)
}

func (r *repositoryStub) Delete(
	ctx context.Context,
	productID string,
	relationshipID string,
) error {
	if r.delete == nil {
		return nil
	}
	return r.delete(ctx, productID, relationshipID)
}
