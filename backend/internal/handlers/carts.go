package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/carts"
)

func (s *Server) CreateCart(
	ctx context.Context,
	_ generated.CreateCartRequestObject,
) (generated.CreateCartResponseObject, error) {
	created, err := s.carts.CreateCart(ctx)
	if err != nil {
		return nil, err
	}

	response, err := toAPICreatedCart(created)
	if err != nil {
		return nil, err
	}
	return generated.CreateCart201JSONResponse(response), nil
}

func (s *Server) GetCart(
	ctx context.Context,
	request generated.GetCartRequestObject,
) (generated.GetCartResponseObject, error) {
	cart, err := s.carts.GetCart(ctx, string(request.Params.XCartToken))
	if err != nil {
		switch {
		case errors.Is(err, carts.ErrInvalidInput):
			return generated.GetCart400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, carts.ErrNotFound):
			return generated.GetCart404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("cart")),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICart(cart)
	if err != nil {
		return nil, err
	}
	return generated.GetCart200JSONResponse(response), nil
}

func (s *Server) AddCartItem(
	ctx context.Context,
	request generated.AddCartItemRequestObject,
) (generated.AddCartItemResponseObject, error) {
	if request.Body == nil {
		return generated.AddCartItem400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	quantity := 1
	if request.Body.Quantity != nil {
		quantity = *request.Body.Quantity
	}
	cart, err := s.carts.AddItem(
		ctx,
		string(request.Params.XCartToken),
		carts.AddItemInput{
			ProductID: request.Body.ProductId.String(),
			Quantity:  quantity,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, carts.ErrInvalidInput):
			return generated.AddCartItem400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, carts.ErrNotFound), errors.Is(err, carts.ErrProductUnavailable):
			return generated.AddCartItem404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("cart or product")),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICart(cart)
	if err != nil {
		return nil, err
	}
	return generated.AddCartItem200JSONResponse(response), nil
}

func (s *Server) UpdateCartItem(
	ctx context.Context,
	request generated.UpdateCartItemRequestObject,
) (generated.UpdateCartItemResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateCartItem400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}

	cart, err := s.carts.UpdateItem(
		ctx,
		string(request.Params.XCartToken),
		request.ProductId.String(),
		request.Body.Quantity,
	)
	if err != nil {
		switch {
		case errors.Is(err, carts.ErrInvalidInput):
			return generated.UpdateCartItem400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, carts.ErrNotFound), errors.Is(err, carts.ErrProductUnavailable):
			return generated.UpdateCartItem404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("cart item")),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICart(cart)
	if err != nil {
		return nil, err
	}
	return generated.UpdateCartItem200JSONResponse(response), nil
}

func (s *Server) DeleteCartItem(
	ctx context.Context,
	request generated.DeleteCartItemRequestObject,
) (generated.DeleteCartItemResponseObject, error) {
	cart, err := s.carts.DeleteItem(
		ctx,
		string(request.Params.XCartToken),
		request.ProductId.String(),
	)
	if err != nil {
		switch {
		case errors.Is(err, carts.ErrInvalidInput):
			return generated.DeleteCartItem400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, carts.ErrNotFound):
			return generated.DeleteCartItem404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("cart item")),
			}, nil
		default:
			return nil, err
		}
	}

	response, err := toAPICart(cart)
	if err != nil {
		return nil, err
	}
	return generated.DeleteCartItem200JSONResponse(response), nil
}

func (s *Server) ClearCart(
	ctx context.Context,
	request generated.ClearCartRequestObject,
) (generated.ClearCartResponseObject, error) {
	err := s.carts.ClearCart(ctx, string(request.Params.XCartToken))
	if err != nil {
		switch {
		case errors.Is(err, carts.ErrInvalidInput):
			return generated.ClearCart400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, carts.ErrNotFound):
			return generated.ClearCart404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("cart")),
			}, nil
		default:
			return nil, err
		}
	}
	return generated.ClearCart204Response{}, nil
}

func toAPICreatedCart(created carts.CreatedCart) (generated.CreatedCart, error) {
	cart, err := toAPICart(created.Cart)
	if err != nil {
		return generated.CreatedCart{}, err
	}
	token := created.Token
	return generated.CreatedCart{Cart: cart, Token: &token}, nil
}

func toAPICart(cart carts.Cart) (generated.Cart, error) {
	id, err := uuid.Parse(cart.ID)
	if err != nil {
		return generated.Cart{}, fmt.Errorf("parse cart id: %w", err)
	}

	items := make([]generated.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		responseItem, err := toAPICartItem(item)
		if err != nil {
			return generated.Cart{}, err
		}
		items = append(items, responseItem)
	}

	createdAt := cart.CreatedAt
	updatedAt := cart.UpdatedAt
	expiresAt := cart.ExpiresAt
	return generated.Cart{
		Id:        id,
		Status:    generated.CartStatus(cart.Status),
		Currency:  cart.Currency,
		Items:     items,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		ExpiresAt: &expiresAt,
	}, nil
}

func toAPICartItem(item carts.CartItem) (generated.CartItem, error) {
	id, err := uuid.Parse(item.ID)
	if err != nil {
		return generated.CartItem{}, fmt.Errorf("parse cart item id: %w", err)
	}
	productID, err := uuid.Parse(item.ProductID)
	if err != nil {
		return generated.CartItem{}, fmt.Errorf("parse cart product id: %w", err)
	}
	createdAt := item.CreatedAt
	updatedAt := item.UpdatedAt
	return generated.CartItem{
		Id:        id,
		ProductId: productID,
		Name:      item.Name,
		Slug:      item.Slug,
		Price:     item.Price,
		Quantity:  item.Quantity,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}, nil
}
