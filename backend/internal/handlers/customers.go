package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/customers"
)

func (s *Server) GetCustomers(
	ctx context.Context,
	_ generated.GetCustomersRequestObject,
) (generated.GetCustomersResponseObject, error) {
	result, err := s.customers.GetCustomers(ctx)
	if err != nil {
		return nil, err
	}
	response := make(generated.GetCustomers200JSONResponse, 0, len(result))
	for _, customer := range result {
		apiCustomer, err := toAPICustomer(customer)
		if err != nil {
			return nil, err
		}
		response = append(response, apiCustomer)
	}
	return response, nil
}

func (s *Server) GetCustomer(
	ctx context.Context,
	request generated.GetCustomerRequestObject,
) (generated.GetCustomerResponseObject, error) {
	customer, err := s.customers.GetCustomer(ctx, request.CustomerId.String())
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return generated.GetCustomer404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("customer")),
			}, nil
		}
		return nil, err
	}
	response, err := toAPICustomer(customer)
	if err != nil {
		return nil, err
	}
	return generated.GetCustomer200JSONResponse(response), nil
}

func (s *Server) CreateCustomer(
	ctx context.Context,
	request generated.CreateCustomerRequestObject,
) (generated.CreateCustomerResponseObject, error) {
	if request.Body == nil {
		return generated.CreateCustomer400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}
	marketingConsent := false
	if request.Body.MarketingConsent != nil {
		marketingConsent = *request.Body.MarketingConsent
	}
	customer, err := s.customers.CreateCustomer(ctx, customers.CreateCustomerInput{
		Email:            string(request.Body.Email),
		FirstName:        optionalString(request.Body.FirstName),
		LastName:         optionalString(request.Body.LastName),
		Phone:            optionalString(request.Body.Phone),
		MarketingConsent: marketingConsent,
	})
	if err != nil {
		switch {
		case errors.Is(err, customers.ErrInvalidInput):
			return generated.CreateCustomer400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, customers.ErrConflict):
			return generated.CreateCustomer409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(generated.Error{
					Code: generated.ErrorCodeConflict, Message: "a customer with this email already exists",
				}),
			}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPICustomer(customer)
	if err != nil {
		return nil, err
	}
	return generated.CreateCustomer201JSONResponse(response), nil
}

func (s *Server) UpdateCustomer(
	ctx context.Context,
	request generated.UpdateCustomerRequestObject,
) (generated.UpdateCustomerResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateCustomer400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(
				errors.New("request body is required"),
			)),
		}, nil
	}
	customer, err := s.customers.UpdateCustomer(
		ctx,
		request.CustomerId.String(),
		customers.UpdateCustomerInput{
			Email:            string(request.Body.Email),
			FirstName:        optionalString(request.Body.FirstName),
			LastName:         optionalString(request.Body.LastName),
			Phone:            optionalString(request.Body.Phone),
			MarketingConsent: request.Body.MarketingConsent,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, customers.ErrInvalidInput):
			return generated.UpdateCustomer400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse(invalidRequestError(err)),
			}, nil
		case errors.Is(err, customers.ErrNotFound):
			return generated.UpdateCustomer404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("customer")),
			}, nil
		case errors.Is(err, customers.ErrConflict):
			return generated.UpdateCustomer409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse(generated.Error{
					Code: generated.ErrorCodeConflict, Message: "a customer with this email already exists",
				}),
			}, nil
		default:
			return nil, err
		}
	}
	response, err := toAPICustomer(customer)
	if err != nil {
		return nil, err
	}
	return generated.UpdateCustomer200JSONResponse(response), nil
}

func (s *Server) DeleteCustomer(
	ctx context.Context,
	request generated.DeleteCustomerRequestObject,
) (generated.DeleteCustomerResponseObject, error) {
	if err := s.customers.DeleteCustomer(ctx, request.CustomerId.String()); err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return generated.DeleteCustomer404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("customer")),
			}, nil
		}
		return nil, err
	}
	return generated.DeleteCustomer204Response{}, nil
}

func toAPICustomer(customer customers.Customer) (generated.Customer, error) {
	id, err := uuid.Parse(customer.ID)
	if err != nil {
		return generated.Customer{}, fmt.Errorf("parse customer id: %w", err)
	}
	createdAt := customer.CreatedAt
	updatedAt := customer.UpdatedAt
	return generated.Customer{
		Id:                 id,
		Email:              openapi_types.Email(customer.Email),
		FirstName:          customer.FirstName,
		LastName:           customer.LastName,
		Phone:              customer.Phone,
		StripeCustomerId:   customer.StripeCustomerID,
		AuthUserId:         customer.AuthUserID,
		MarketingConsent:   customer.MarketingConsent,
		MarketingConsentAt: customer.MarketingConsentAt,
		CreatedAt:          &createdAt,
		UpdatedAt:          &updatedAt,
	}, nil
}
