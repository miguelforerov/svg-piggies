package productimages

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type altTextRepositoryStub struct {
	input AltTextUpdate
	calls int
	err   error
}

func (s *altTextRepositoryStub) UpdateAltText(_ context.Context, input AltTextUpdate) error {
	s.input = input
	s.calls++
	return s.err
}

func TestUpdateAltTextValues(t *testing.T) {
	for _, tc := range []struct {
		name    string
		present bool
		input   *string
		want    *string
	}{
		{name: "omitted"},
		{name: "null", present: true},
		{name: "empty", present: true, input: altTextPtr(""), want: altTextPtr("")},
		{name: "spaces", present: true, input: altTextPtr("  \t "), want: altTextPtr("")},
		{name: "text", present: true, input: altTextPtr("  Pink pig  "), want: altTextPtr("Pink pig")},
		{name: "500 Unicode characters", present: true, input: altTextPtr(strings.Repeat("🐷", 500)), want: altTextPtr(strings.Repeat("🐷", 500))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &altTextRepositoryStub{}
			service, err := NewAltTextService(repo)
			if err != nil {
				t.Fatal(err)
			}
			if err := service.UpdateAltText(context.Background(), AltTextUpdate{ProductID: " product ", ImageID: " image ", Present: tc.present, AltText: tc.input}); err != nil {
				t.Fatal(err)
			}
			if repo.calls != 1 || repo.input.ProductID != "product" || repo.input.ImageID != "image" || repo.input.Present != tc.present {
				t.Fatalf("repository input = %+v", repo.input)
			}
			if (repo.input.AltText == nil) != (tc.want == nil) {
				t.Fatalf("alt text = %v, want %v", repo.input.AltText, tc.want)
			}
			if tc.want != nil && *repo.input.AltText != *tc.want {
				t.Fatalf("alt text = %q, want %q", *repo.input.AltText, *tc.want)
			}
		})
	}
}

func TestUpdateAltTextRejectsTooLongAndPreservesDatabaseErrors(t *testing.T) {
	repo := &altTextRepositoryStub{}
	service, err := NewAltTextService(repo)
	if err != nil {
		t.Fatal(err)
	}
	longText := strings.Repeat("🐷", 501)
	err = service.UpdateAltText(context.Background(), AltTextUpdate{ProductID: "product", ImageID: "image", Present: true, AltText: &longText})
	if !errors.Is(err, ErrInvalidInput) || repo.calls != 0 {
		t.Fatalf("long text error = %v; calls = %d", err, repo.calls)
	}
	repo.err = errors.New("database failed")
	err = service.UpdateAltText(context.Background(), AltTextUpdate{ProductID: "product", ImageID: "image", Present: true, AltText: altTextPtr("valid")})
	if !errors.Is(err, repo.err) {
		t.Fatalf("database error = %v", err)
	}
}

func altTextPtr(value string) *string { return &value }
