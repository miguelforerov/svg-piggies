package collections

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("collections repository is required")
	}
	return &Service{repository: repository}, nil
}

func (s *Service) GetCollections(ctx context.Context) ([]Collection, error) {
	collections, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get collections: %w", err)
	}
	if collections == nil {
		return []Collection{}, nil
	}
	return collections, nil
}

func (s *Service) GetCollection(ctx context.Context, id string) (Collection, error) {
	id, err := validateID(id)
	if err != nil {
		return Collection{}, err
	}

	collection, err := s.repository.Get(ctx, id)
	if err != nil {
		return Collection{}, fmt.Errorf("get collection %s: %w", id, err)
	}
	return collection, nil
}

func (s *Service) CreateCollection(
	ctx context.Context,
	input CreateCollectionInput,
) (Collection, error) {
	input, err := validateCreateInput(input)
	if err != nil {
		return Collection{}, err
	}

	collection, err := s.repository.Create(ctx, input)
	if err != nil {
		return Collection{}, fmt.Errorf("create collection: %w", err)
	}
	return collection, nil
}

func (s *Service) UpdateCollection(
	ctx context.Context,
	id string,
	input UpdateCollectionInput,
) (Collection, error) {
	id, err := validateID(id)
	if err != nil {
		return Collection{}, err
	}
	input, err = validateUpdateInput(input)
	if err != nil {
		return Collection{}, err
	}

	collection, err := s.repository.Update(ctx, id, input)
	if err != nil {
		return Collection{}, fmt.Errorf("update collection %s: %w", id, err)
	}
	return collection, nil
}

func (s *Service) DeleteCollection(ctx context.Context, id string) error {
	id, err := validateID(id)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete collection %s: %w", id, err)
	}
	return nil
}

func validateID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return id, nil
}

func validateCreateInput(input CreateCollectionInput) (CreateCollectionInput, error) {
	name, slug, description, err := validateFields(input.Name, input.Slug, input.Description)
	if err != nil {
		return CreateCollectionInput{}, err
	}
	return CreateCollectionInput{
		Name:        name,
		Slug:        slug,
		Description: description,
	}, nil
}

func validateUpdateInput(input UpdateCollectionInput) (UpdateCollectionInput, error) {
	name, slug, description, err := validateFields(input.Name, input.Slug, input.Description)
	if err != nil {
		return UpdateCollectionInput{}, err
	}
	return UpdateCollectionInput{
		Name:        name,
		Slug:        slug,
		Description: description,
	}, nil
}

func validateFields(name, slug, description string) (string, string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", "", fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", "", "", fmt.Errorf("%w: slug is required", ErrInvalidInput)
	}

	return name, slug, strings.TrimSpace(description), nil
}
