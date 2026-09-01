package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
)

func (s *Server) GetCollections(
	ctx context.Context,
	_ generated.GetCollectionsRequestObject,
) (generated.GetCollectionsResponseObject, error) {
	result, err := s.collections.GetCollections(ctx)
	if err != nil {
		return nil, err
	}

	response := make(generated.GetCollections200JSONResponse, 0, len(result))
	for _, collection := range result {
		apiCollection, err := toAPICollection(collection)
		if err != nil {
			return nil, err
		}
		response = append(response, apiCollection)
	}
	return response, nil
}

func (s *Server) GetCollection(
	ctx context.Context,
	request generated.GetCollectionRequestObject,
) (generated.GetCollectionResponseObject, error) {
	collection, err := s.collections.GetCollection(ctx, request.CollectionId.String())
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.GetCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrNotFound):
			return generated.GetCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("collection")),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.GetCollection200JSONResponse(response), nil
}

func (s *Server) GetCollectionBySlug(
	ctx context.Context,
	request generated.GetCollectionBySlugRequestObject,
) (generated.GetCollectionBySlugResponseObject, error) {
	collection, err := s.collections.GetCollectionBySlug(ctx, request.Slug)
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrNotFound):
			return generated.GetCollectionBySlug404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("collection")),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.GetCollectionBySlug200JSONResponse(response), nil
}

func (s *Server) CreateCollection(
	ctx context.Context,
	request generated.CreateCollectionRequestObject,
) (generated.CreateCollectionResponseObject, error) {
	if request.Body == nil {
		return generated.CreateCollection400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	collection, err := s.collections.CreateCollection(ctx, collections.CreateCollectionInput{
		Name:        request.Body.Name,
		Slug:        request.Body.Slug,
		Description: optionalString(request.Body.Description),
	})
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.CreateCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrConflict):
			return generated.CreateCollection409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(conflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.CreateCollection201JSONResponse(response), nil
}

func (s *Server) UpdateCollection(
	ctx context.Context,
	request generated.UpdateCollectionRequestObject,
) (generated.UpdateCollectionResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateCollection400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	collection, err := s.collections.UpdateCollection(
		ctx,
		request.CollectionId.String(),
		collections.UpdateCollectionInput{
			Name:        request.Body.Name,
			Slug:        request.Body.Slug,
			Description: optionalString(request.Body.Description),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.UpdateCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrNotFound):
			return generated.UpdateCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("collection")),
			}, nil
		case errors.Is(err, collections.ErrConflict):
			return generated.UpdateCollection409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(conflictError()),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICollection(collection)
	if err != nil {
		return nil, err
	}
	return generated.UpdateCollection200JSONResponse(response), nil
}

func (s *Server) DeleteCollection(
	ctx context.Context,
	request generated.DeleteCollectionRequestObject,
) (generated.DeleteCollectionResponseObject, error) {
	err := s.collections.DeleteCollection(ctx, request.CollectionId.String())
	if err != nil {
		switch {
		case errors.Is(err, collections.ErrInvalidInput):
			return generated.DeleteCollection400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, collections.ErrNotFound):
			return generated.DeleteCollection404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("collection")),
			}, nil
		default:
			return nil, err
		}
	}

	return generated.DeleteCollection204Response{}, nil
}

func toAPICollection(collection collections.Collection) (generated.Collection, error) {
	id, err := uuid.Parse(collection.ID)
	if err != nil {
		return generated.Collection{}, fmt.Errorf("parse collection id: %w", err)
	}
	return generated.Collection{
		Id:          id,
		Name:        collection.Name,
		Slug:        collection.Slug,
		Description: collection.Description,
	}, nil
}
