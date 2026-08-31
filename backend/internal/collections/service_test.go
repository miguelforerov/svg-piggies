package collections

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

func TestGetCollectionsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &repositoryStub{})
	got, err := service.GetCollections(context.Background())
	if err != nil {
		t.Fatalf("GetCollections() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("GetCollections() = %#v, want empty non-nil slice", got)
	}
}

func TestGetCollection(t *testing.T) {
	t.Parallel()

	want := Collection{ID: "collection-id", Name: "Animals", Slug: "animals"}
	repository := &repositoryStub{
		get: func(_ context.Context, id string) (Collection, error) {
			if id != want.ID {
				t.Fatalf("Get() id = %q, want %q", id, want.ID)
			}
			return want, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.GetCollection(context.Background(), "  collection-id  ")
	if err != nil {
		t.Fatalf("GetCollection() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetCollection() = %#v, want %#v", got, want)
	}
}

func TestCreateCollectionValidatesAndTrimsInput(t *testing.T) {
	t.Parallel()

	wantInput := CreateCollectionInput{
		Name:        "Animals",
		Slug:        "animals",
		Description: "Animal illustrations",
	}
	repository := &repositoryStub{
		create: func(_ context.Context, input CreateCollectionInput) (Collection, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create() input = %#v, want %#v", input, wantInput)
			}
			return Collection{ID: "collection-id", Name: input.Name, Slug: input.Slug}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.CreateCollection(context.Background(), CreateCollectionInput{
		Name:        " Animals ",
		Slug:        " animals ",
		Description: " Animal illustrations ",
	})
	if err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	if got.ID != "collection-id" {
		t.Errorf("CreateCollection() ID = %q", got.ID)
	}
}

func TestUpdateCollection(t *testing.T) {
	t.Parallel()

	wantInput := UpdateCollectionInput{Name: "Wildlife", Slug: "wildlife"}
	repository := &repositoryStub{
		update: func(_ context.Context, id string, input UpdateCollectionInput) (Collection, error) {
			if id != "collection-id" {
				t.Fatalf("Update() id = %q", id)
			}
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Update() input = %#v, want %#v", input, wantInput)
			}
			return Collection{ID: id, Name: input.Name, Slug: input.Slug}, nil
		},
	}
	service := newTestService(t, repository)

	got, err := service.UpdateCollection(
		context.Background(),
		" collection-id ",
		UpdateCollectionInput{Name: " Wildlife ", Slug: " wildlife "},
	)
	if err != nil {
		t.Fatalf("UpdateCollection() error = %v", err)
	}
	if got.Name != "Wildlife" {
		t.Errorf("UpdateCollection() Name = %q", got.Name)
	}
}

func TestDeleteCollection(t *testing.T) {
	t.Parallel()

	called := false
	repository := &repositoryStub{
		delete: func(_ context.Context, id string) error {
			called = true
			if id != "collection-id" {
				t.Fatalf("Delete() id = %q", id)
			}
			return nil
		},
	}
	service := newTestService(t, repository)

	if err := service.DeleteCollection(context.Background(), " collection-id "); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
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
				_, err := service.GetCollection(context.Background(), " ")
				return err
			},
		},
		{
			name: "missing create name",
			run: func() error {
				_, err := service.CreateCollection(context.Background(), CreateCollectionInput{Slug: "animals"})
				return err
			},
		},
		{
			name: "missing update slug",
			run: func() error {
				_, err := service.UpdateCollection(
					context.Background(),
					"collection-id",
					UpdateCollectionInput{Name: "Animals"},
				)
				return err
			},
		},
		{
			name: "missing delete id",
			run: func() error {
				return service.DeleteCollection(context.Background(), "")
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
		get: func(context.Context, string) (Collection, error) {
			return Collection{}, ErrNotFound
		},
	}
	service := newTestService(t, repository)

	_, err := service.GetCollection(context.Background(), "collection-id")
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
	list   func(context.Context) ([]Collection, error)
	get    func(context.Context, string) (Collection, error)
	create func(context.Context, CreateCollectionInput) (Collection, error)
	update func(context.Context, string, UpdateCollectionInput) (Collection, error)
	delete func(context.Context, string) error
}

func (r *repositoryStub) List(ctx context.Context) ([]Collection, error) {
	if r.list == nil {
		return nil, nil
	}
	return r.list(ctx)
}

func (r *repositoryStub) Get(ctx context.Context, id string) (Collection, error) {
	if r.get == nil {
		return Collection{}, nil
	}
	return r.get(ctx, id)
}

func (r *repositoryStub) Create(
	ctx context.Context,
	input CreateCollectionInput,
) (Collection, error) {
	if r.create == nil {
		return Collection{}, nil
	}
	return r.create(ctx, input)
}

func (r *repositoryStub) Update(
	ctx context.Context,
	id string,
	input UpdateCollectionInput,
) (Collection, error) {
	if r.update == nil {
		return Collection{}, nil
	}
	return r.update(ctx, id, input)
}

func (r *repositoryStub) Delete(ctx context.Context, id string) error {
	if r.delete == nil {
		return nil
	}
	return r.delete(ctx, id)
}
