package productimages

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type orderRepoStub struct {
	calls   int
	product string
	ids     []string
	err     error
}

func (s *orderRepoStub) Reorder(_ context.Context, product string, ids []string) error {
	s.calls++
	s.product = product
	s.ids = ids
	return s.err
}

func TestOrderServiceValidationAndDelegation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		product string
		ids     []string
		valid   bool
	}{
		{"complete", " product ", []string{"first", "second"}, true},
		{"empty gallery", "product", []string{}, true},
		{"missing list", "product", nil, false},
		{"missing product", "", []string{}, false},
		{"duplicate", "product", []string{"first", "first"}, false},
		{"empty id", "product", []string{""}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &orderRepoStub{}
			service, _ := NewOrderService(repo)
			err := service.ReorderImages(context.Background(), tc.product, tc.ids)
			if tc.valid {
				if err != nil || repo.calls != 1 || repo.product != "product" || !reflect.DeepEqual(repo.ids, tc.ids) {
					t.Fatalf("err=%v repo=%+v", err, repo)
				}
			} else if !errors.Is(err, ErrInvalidInput) || repo.calls != 0 {
				t.Fatalf("err=%v calls=%d", err, repo.calls)
			}
		})
	}
}

func TestOrderServicePropagatesRepositoryFailure(t *testing.T) {
	repo := &orderRepoStub{err: ErrReferenceNotFound}
	service, _ := NewOrderService(repo)
	if err := service.ReorderImages(context.Background(), "product", []string{}); !errors.Is(err, ErrReferenceNotFound) {
		t.Fatalf("err=%v", err)
	}
}
