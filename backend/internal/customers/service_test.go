package customers

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	create func(context.Context, CreateCustomerInput) (Customer, error)
}

func (r *repositoryStub) List(context.Context) ([]Customer, error) { return nil, nil }
func (r *repositoryStub) Get(context.Context, string) (Customer, error) {
	return Customer{}, ErrNotFound
}
func (r *repositoryStub) Create(
	ctx context.Context,
	input CreateCustomerInput,
) (Customer, error) {
	return r.create(ctx, input)
}
func (r *repositoryStub) Update(
	context.Context,
	string,
	UpdateCustomerInput,
) (Customer, error) {
	return Customer{}, ErrNotFound
}
func (r *repositoryStub) Delete(context.Context, string) error { return ErrNotFound }

func TestCreateCustomerNormalizesInput(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{create: func(
		_ context.Context,
		input CreateCustomerInput,
	) (Customer, error) {
		if input.Email != "person@example.com" {
			t.Fatalf("email = %q", input.Email)
		}
		if input.FirstName != "Ada" || input.LastName != "Lovelace" {
			t.Fatalf("name was not trimmed: %#v", input)
		}
		return Customer{Email: input.Email}, nil
	}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.CreateCustomer(context.Background(), CreateCustomerInput{
		Email: " Person@Example.com ", FirstName: " Ada ", LastName: " Lovelace ",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateCustomerRejectsInvalidEmail(t *testing.T) {
	t.Parallel()

	service, err := NewService(&repositoryStub{create: func(
		context.Context,
		CreateCustomerInput,
	) (Customer, error) {
		t.Fatal("repository should not be called")
		return Customer{}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.CreateCustomer(context.Background(), CreateCustomerInput{Email: "invalid"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}
