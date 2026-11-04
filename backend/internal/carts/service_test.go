package carts

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

const (
	testCartToken = "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFG"
	testProductID = "4b09ea69-8252-4259-87d6-281020cf05f1"
)

func TestNewServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	if _, err := NewService(nil); err == nil {
		t.Fatal("NewService(nil) error = nil, want error")
	}
}

func TestCreateCartReturnsTokenButStoresOnlyHash(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{
		createGuest: func(_ context.Context, tokenHash string) (Cart, error) {
			if len(tokenHash) != 64 {
				t.Fatalf("token hash length = %d, want 64", len(tokenHash))
			}
			return Cart{ID: "cart-id", Status: StatusActive, Currency: "aud"}, nil
		},
	}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := service.CreateCart(context.Background())
	if err != nil {
		t.Fatalf("CreateCart() error = %v", err)
	}
	if len(created.Token) < 20 {
		t.Fatalf("token length = %d, want opaque token", len(created.Token))
	}
	if created.Cart.Items == nil {
		t.Fatal("CreateCart() items = nil, want empty slice")
	}
}

func TestAddItemValidatesAndHashesInput(t *testing.T) {
	t.Parallel()

	want := Cart{
		ID:       "cart-id",
		Status:   StatusActive,
		Currency: "aud",
		Items: []CartItem{{
			ID:        "item-id",
			ProductID: testProductID,
			Quantity:  2,
		}},
	}
	repository := &repositoryStub{
		addItem: func(
			_ context.Context,
			tokenHash string,
			productID string,
			quantity int,
		) (Cart, error) {
			if tokenHash != hashToken(testCartToken) {
				t.Fatalf("token hash = %q", tokenHash)
			}
			if productID != testProductID || quantity != 2 {
				t.Fatalf("AddItem() input = %q, %d", productID, quantity)
			}
			return want, nil
		},
	}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got, err := service.AddItem(context.Background(), testCartToken, AddItemInput{
		ProductID: testProductID,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AddItem() = %#v, want %#v", got, want)
	}
}

func TestCartInputValidation(t *testing.T) {
	t.Parallel()

	service, err := NewService(&repositoryStub{})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "missing token",
			run: func() error {
				_, err := service.GetCart(context.Background(), "")
				return err
			},
		},
		{
			name: "invalid product id",
			run: func() error {
				_, err := service.AddItem(context.Background(), testCartToken, AddItemInput{
					ProductID: "not-a-uuid",
					Quantity:  1,
				})
				return err
			},
		},
		{
			name: "zero quantity",
			run: func() error {
				_, err := service.AddItem(context.Background(), testCartToken, AddItemInput{
					ProductID: testProductID,
					Quantity:  0,
				})
				return err
			},
		},
		{
			name: "quantity too large",
			run: func() error {
				_, err := service.UpdateItem(
					context.Background(),
					testCartToken,
					testProductID,
					MaxQuantity+1,
				)
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.run(); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

type repositoryStub struct {
	createGuest func(context.Context, string) (Cart, error)
	get         func(context.Context, string) (Cart, error)
	addItem     func(context.Context, string, string, int) (Cart, error)
	updateItem  func(context.Context, string, string, int) (Cart, error)
	deleteItem  func(context.Context, string, string) (Cart, error)
	clear       func(context.Context, string) error
}

func (r *repositoryStub) CreateGuest(ctx context.Context, tokenHash string) (Cart, error) {
	if r.createGuest == nil {
		return Cart{}, errors.New("unexpected CreateGuest call")
	}
	return r.createGuest(ctx, tokenHash)
}

func (r *repositoryStub) GetByTokenHash(ctx context.Context, tokenHash string) (Cart, error) {
	if r.get == nil {
		return Cart{}, errors.New("unexpected GetByTokenHash call")
	}
	return r.get(ctx, tokenHash)
}

func (r *repositoryStub) AddItem(
	ctx context.Context,
	tokenHash string,
	productID string,
	quantity int,
) (Cart, error) {
	if r.addItem == nil {
		return Cart{}, errors.New("unexpected AddItem call")
	}
	return r.addItem(ctx, tokenHash, productID, quantity)
}

func (r *repositoryStub) UpdateItem(
	ctx context.Context,
	tokenHash string,
	productID string,
	quantity int,
) (Cart, error) {
	if r.updateItem == nil {
		return Cart{}, errors.New("unexpected UpdateItem call")
	}
	return r.updateItem(ctx, tokenHash, productID, quantity)
}

func (r *repositoryStub) DeleteItem(
	ctx context.Context,
	tokenHash string,
	productID string,
) (Cart, error) {
	if r.deleteItem == nil {
		return Cart{}, errors.New("unexpected DeleteItem call")
	}
	return r.deleteItem(ctx, tokenHash, productID)
}

func (r *repositoryStub) Clear(ctx context.Context, tokenHash string) error {
	if r.clear == nil {
		return errors.New("unexpected Clear call")
	}
	return r.clear(ctx, tokenHash)
}
