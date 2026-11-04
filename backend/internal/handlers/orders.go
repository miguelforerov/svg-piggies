package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/orders"
)

func (s *Server) CreateCheckoutOrder(
	ctx context.Context,
	request generated.CreateCheckoutOrderRequestObject,
) (generated.CreateCheckoutOrderResponseObject, error) {
	if request.Body == nil {
		return generated.CreateCheckoutOrder400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}
	order, err := s.orders.CreateCheckoutOrder(
		ctx,
		string(request.Params.XCartToken),
		orders.CheckoutCustomerInput{
			Email:     string(request.Body.Email),
			FirstName: optionalString(request.Body.FirstName),
			LastName:  optionalString(request.Body.LastName),
			Phone:     optionalString(request.Body.Phone),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, orders.ErrInvalidInput), errors.Is(err, orders.ErrEmptyCart):
			return generated.CreateCheckoutOrder400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, orders.ErrNotFound), errors.Is(err, orders.ErrProductUnavailable):
			return generated.CreateCheckoutOrder404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("cart or product")),
			}, nil
		case errors.Is(err, orders.ErrConflict):
			return generated.CreateCheckoutOrder409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(generated.Error{
					Code: generated.ErrorCodeConflict, Message: "the cart has already been converted",
				}),
			}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPIOrder(order)
	if err != nil {
		return nil, err
	}
	return generated.CreateCheckoutOrder201JSONResponse(response), nil
}

func (s *Server) GetOrders(
	ctx context.Context,
	_ generated.GetOrdersRequestObject,
) (generated.GetOrdersResponseObject, error) {
	result, err := s.orders.GetOrders(ctx)
	if err != nil {
		return nil, err
	}
	response := make(generated.GetOrders200JSONResponse, 0, len(result))
	for _, order := range result {
		apiOrder, err := toAPIOrder(order)
		if err != nil {
			return nil, err
		}
		response = append(response, apiOrder)
	}
	return response, nil
}

func (s *Server) GetOrder(
	ctx context.Context,
	request generated.GetOrderRequestObject,
) (generated.GetOrderResponseObject, error) {
	order, err := s.orders.GetOrder(ctx, request.OrderId.String())
	if err != nil {
		if errors.Is(err, orders.ErrNotFound) {
			return generated.GetOrder404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("order")),
			}, nil
		}
		return nil, err
	}
	response, err := toAPIOrder(order)
	if err != nil {
		return nil, err
	}
	return generated.GetOrder200JSONResponse(response), nil
}

func (s *Server) UpdateOrder(
	ctx context.Context,
	request generated.UpdateOrderRequestObject,
) (generated.UpdateOrderResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateOrder400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}
	order, err := s.orders.UpdateOrder(
		ctx,
		request.OrderId.String(),
		orders.Status(request.Body.OrderStatus),
	)
	if err != nil {
		switch {
		case errors.Is(err, orders.ErrInvalidInput):
			return generated.UpdateOrder400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, orders.ErrNotFound):
			return generated.UpdateOrder404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("order")),
			}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPIOrder(order)
	if err != nil {
		return nil, err
	}
	return generated.UpdateOrder200JSONResponse(response), nil
}

func toAPIOrder(order orders.Order) (generated.Order, error) {
	id, err := uuid.Parse(order.ID)
	if err != nil {
		return generated.Order{}, fmt.Errorf("parse order id: %w", err)
	}
	var customerID *openapi_types.UUID
	if order.CustomerID != nil {
		parsed, err := uuid.Parse(*order.CustomerID)
		if err != nil {
			return generated.Order{}, fmt.Errorf("parse order customer id: %w", err)
		}
		customerID = &parsed
	}
	items := make([]generated.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		apiItem, err := toAPIOrderItem(item)
		if err != nil {
			return generated.Order{}, err
		}
		items = append(items, apiItem)
	}
	createdAt := order.CreatedAt
	updatedAt := order.UpdatedAt
	return generated.Order{
		Id:                      id,
		CustomerId:              customerID,
		CustomerEmail:           openapi_types.Email(order.CustomerEmail),
		OrderStatus:             generated.OrderStatus(order.OrderStatus),
		PaymentStatus:           generated.PaymentStatus(order.PaymentStatus),
		Currency:                order.Currency,
		SubtotalAmount:          order.SubtotalAmount,
		DiscountAmount:          order.DiscountAmount,
		TaxAmount:               order.TaxAmount,
		TotalAmount:             order.TotalAmount,
		AmountRefunded:          order.AmountRefunded,
		StripeCheckoutSessionId: order.StripeCheckoutSessionID,
		StripePaymentIntentId:   order.StripePaymentIntentID,
		Items:                   items,
		CreatedAt:               &createdAt,
		UpdatedAt:               &updatedAt,
		PaidAt:                  order.PaidAt,
		FulfilledAt:             order.FulfilledAt,
		CancelledAt:             order.CancelledAt,
	}, nil
}

func toAPIOrderItem(item orders.OrderItem) (generated.OrderItem, error) {
	id, err := uuid.Parse(item.ID)
	if err != nil {
		return generated.OrderItem{}, fmt.Errorf("parse order item id: %w", err)
	}
	orderID, err := uuid.Parse(item.OrderID)
	if err != nil {
		return generated.OrderItem{}, fmt.Errorf("parse order item order id: %w", err)
	}
	productID, err := uuid.Parse(item.ProductID)
	if err != nil {
		return generated.OrderItem{}, fmt.Errorf("parse order item product id: %w", err)
	}
	createdAt := item.CreatedAt
	return generated.OrderItem{
		Id:              id,
		OrderId:         orderID,
		ProductId:       productID,
		ProductName:     item.ProductName,
		VariantName:     item.VariantName,
		Sku:             item.SKU,
		Quantity:        item.Quantity,
		UnitPriceAmount: item.UnitPriceAmount,
		LineTotalAmount: item.LineTotalAmount,
		CreatedAt:       &createdAt,
	}, nil
}
