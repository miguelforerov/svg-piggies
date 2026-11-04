package orders

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	createFromCart func(context.Context, CreateCheckoutOrderInput) (Order, error)
	updateStatus   func(context.Context, string, Status) (Order, error)
}

func (r *repositoryStub) List(context.Context) ([]Order, error) { return nil, nil }
func (r *repositoryStub) Get(context.Context, string) (Order, error) {
	return Order{}, ErrNotFound
}
func (r *repositoryStub) CreateFromCart(
	ctx context.Context,
	input CreateCheckoutOrderInput,
) (Order, error) {
	return r.createFromCart(ctx, input)
}
func (r *repositoryStub) UpdateStatus(
	ctx context.Context,
	id string,
	status Status,
) (Order, error) {
	return r.updateStatus(ctx, id, status)
}

func TestCreateCheckoutOrderHashesTokenAndNormalizesCustomer(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{createFromCart: func(
		_ context.Context,
		input CreateCheckoutOrderInput,
	) (Order, error) {
		if len(input.CartTokenHash) != 64 || input.CartTokenHash == "abcdefghijklmnopqrstuvwxyz" {
			t.Fatalf("cart token hash = %q", input.CartTokenHash)
		}
		if input.Customer.Email != "buyer@example.com" {
			t.Fatalf("email = %q", input.Customer.Email)
		}
		if input.Customer.FirstName != "Grace" {
			t.Fatalf("first name = %q", input.Customer.FirstName)
		}
		return Order{Items: nil}, nil
	}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatal(err)
	}

	order, err := service.CreateCheckoutOrder(
		context.Background(),
		"abcdefghijklmnopqrstuvwxyz",
		CheckoutCustomerInput{Email: " Buyer@Example.com ", FirstName: " Grace "},
	)
	if err != nil {
		t.Fatal(err)
	}
	if order.Items == nil {
		t.Fatal("items should be normalized to an empty slice")
	}
}

func TestCreateCheckoutOrderRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{createFromCart: func(
		context.Context,
		CreateCheckoutOrderInput,
	) (Order, error) {
		t.Fatal("repository should not be called")
		return Order{}, nil
	}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.CreateCheckoutOrder(
		context.Background(), "short", CheckoutCustomerInput{Email: "buyer@example.com"},
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("short token error = %v, want ErrInvalidInput", err)
	}
	_, err = service.CreateCheckoutOrder(
		context.Background(), "abcdefghijklmnopqrstuvwxyz", CheckoutCustomerInput{Email: "invalid"},
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid email error = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateOrderRejectsPaymentLikeStatus(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{updateStatus: func(
		context.Context,
		string,
		Status,
	) (Order, error) {
		t.Fatal("repository should not be called")
		return Order{}, nil
	}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.UpdateOrder(
		context.Background(), "4b09ea69-8252-4259-87d6-281020cf05f1", Status("paid"),
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}
