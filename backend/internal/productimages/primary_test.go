package productimages

import (
	"context"
	"errors"
	"testing"
)

type primaryRepositoryStub struct {
	productID string
	imageID   string
	calls     int
	err       error
}

func (s *primaryRepositoryStub) SetPrimary(_ context.Context, productID, imageID string) error {
	s.productID, s.imageID = productID, imageID
	s.calls++
	return s.err
}

func TestSetPrimaryImage(t *testing.T) {
	repo := &primaryRepositoryStub{}
	service, err := NewPrimaryService(repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SetPrimaryImage(context.Background(), " product ", " image "); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 1 || repo.productID != "product" || repo.imageID != "image" {
		t.Fatalf("repository call = %+v", repo)
	}
	if err := service.SetPrimaryImage(context.Background(), "product", " "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid image id error = %v", err)
	}
	if repo.calls != 1 {
		t.Fatal("invalid input reached repository")
	}
}

func TestSetPrimaryImagePreservesRepositoryErrors(t *testing.T) {
	for _, want := range []error{ErrReferenceNotFound, ErrImageNotFound, errors.New("database unavailable")} {
		repo := &primaryRepositoryStub{err: want}
		service, err := NewPrimaryService(repo)
		if err != nil {
			t.Fatal(err)
		}
		if err := service.SetPrimaryImage(context.Background(), "product", "image"); !errors.Is(err, want) {
			t.Fatalf("error = %v, want %v", err, want)
		}
	}
}
