package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/zdenaforero/svg-piggies/backend/internal/auth"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
	"github.com/zdenaforero/svg-piggies/backend/internal/config"
)

const testCollectionID = "4b09ea69-8252-4259-87d6-281020cf05f1"

func TestHealth(t *testing.T) {
	t.Parallel()

	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{}, nil)
	if err != nil {
		t.Fatalf("newHandler() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4321" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestCollectionRoutes(t *testing.T) {
	service := &collectionServiceStub{
		list: func(context.Context) ([]collections.Collection, error) {
			return []collections.Collection{{
				ID:          testCollectionID,
				Name:        "Animals",
				Slug:        "animals",
				Description: "Animal illustrations",
			}}, nil
		},
		get: func(_ context.Context, id string) (collections.Collection, error) {
			if id != testCollectionID {
				t.Fatalf("GetCollection() id = %q", id)
			}
			return collections.Collection{ID: id, Name: "Animals", Slug: "animals"}, nil
		},
		create: func(
			_ context.Context,
			input collections.CreateCollectionInput,
		) (collections.Collection, error) {
			want := collections.CreateCollectionInput{
				Name:        "Animals",
				Slug:        "animals",
				Description: "Animal illustrations",
			}
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("CreateCollection() input = %#v, want %#v", input, want)
			}
			return collections.Collection{
				ID:          testCollectionID,
				Name:        input.Name,
				Slug:        input.Slug,
				Description: input.Description,
			}, nil
		},
		update: func(
			_ context.Context,
			id string,
			input collections.UpdateCollectionInput,
		) (collections.Collection, error) {
			if id != testCollectionID {
				t.Fatalf("UpdateCollection() id = %q", id)
			}
			return collections.Collection{
				ID:          id,
				Name:        input.Name,
				Slug:        input.Slug,
				Description: input.Description,
			}, nil
		},
		delete: func(_ context.Context, id string) error {
			if id != testCollectionID {
				t.Fatalf("DeleteCollection() id = %q", id)
			}
			return nil
		},
	}
	handler := newTestHandler(service)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "list", method: http.MethodGet, path: "/api/admin/collections", wantStatus: http.StatusOK},
		{name: "get", method: http.MethodGet, path: "/api/admin/collections/" + testCollectionID, wantStatus: http.StatusOK},
		{
			name:       "create",
			method:     http.MethodPost,
			path:       "/api/admin/collections",
			body:       `{"name":"Animals","slug":"animals","description":"Animal illustrations"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "update",
			method:     http.MethodPut,
			path:       "/api/admin/collections/" + testCollectionID,
			body:       `{"name":"Wildlife","slug":"wildlife"}`,
			wantStatus: http.StatusOK,
		},
		{name: "delete", method: http.MethodDelete, path: "/api/admin/collections/" + testCollectionID, wantStatus: http.StatusNoContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body)
			}
		})
	}
}

func TestCollectionRouteRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(&collectionServiceStub{})
	request := httptest.NewRequest(http.MethodGet, "/api/admin/collections/not-a-uuid", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_request")
}

func TestCollectionRouteMapsNotFound(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(&collectionServiceStub{
		get: func(context.Context, string) (collections.Collection, error) {
			return collections.Collection{}, collections.ErrNotFound
		},
	})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/collections/"+testCollectionID,
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	assertErrorCode(t, response, "not_found")
}

func newTestHandler(service *collectionServiceStub) http.Handler {
	handler, err := newHandler(config.Config{
		Environment:       "test",
		AuthMode:          auth.ModeDevelopment,
		CORSAllowedOrigin: "http://localhost:4321",
	}, dependencies{collections: service}, nil)
	if err != nil {
		panic(err)
	}
	return handler
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != want {
		t.Fatalf("error code = %q, want %q", body.Code, want)
	}
}

type collectionServiceStub struct {
	list   func(context.Context) ([]collections.Collection, error)
	get    func(context.Context, string) (collections.Collection, error)
	create func(context.Context, collections.CreateCollectionInput) (collections.Collection, error)
	update func(context.Context, string, collections.UpdateCollectionInput) (collections.Collection, error)
	delete func(context.Context, string) error
}

func (s *collectionServiceStub) GetCollections(ctx context.Context) ([]collections.Collection, error) {
	if s.list == nil {
		return nil, errors.New("unexpected GetCollections call")
	}
	return s.list(ctx)
}

func (s *collectionServiceStub) GetCollection(
	ctx context.Context,
	id string,
) (collections.Collection, error) {
	if s.get == nil {
		return collections.Collection{}, errors.New("unexpected GetCollection call")
	}
	return s.get(ctx, id)
}

func (s *collectionServiceStub) CreateCollection(
	ctx context.Context,
	input collections.CreateCollectionInput,
) (collections.Collection, error) {
	if s.create == nil {
		return collections.Collection{}, errors.New("unexpected CreateCollection call")
	}
	return s.create(ctx, input)
}

func (s *collectionServiceStub) UpdateCollection(
	ctx context.Context,
	id string,
	input collections.UpdateCollectionInput,
) (collections.Collection, error) {
	if s.update == nil {
		return collections.Collection{}, errors.New("unexpected UpdateCollection call")
	}
	return s.update(ctx, id, input)
}

func (s *collectionServiceStub) DeleteCollection(ctx context.Context, id string) error {
	if s.delete == nil {
		return errors.New("unexpected DeleteCollection call")
	}
	return s.delete(ctx, id)
}
