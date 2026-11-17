package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	generated "github.com/zdenaforero/svg-piggies/backend/api/generated"
	"github.com/zdenaforero/svg-piggies/backend/internal/productfiles"
)

func fileError(code generated.ErrorCode, message string) generated.Error {
	return generated.Error{Code: code, Message: message}
}

func toAPIProductFile(file productfiles.ProductFile) (generated.AdminProductFile, error) {
	id, err := uuid.Parse(file.ID)
	if err != nil {
		return generated.AdminProductFile{}, err
	}
	productID, err := uuid.Parse(file.ProductID)
	if err != nil {
		return generated.AdminProductFile{}, err
	}
	return generated.AdminProductFile{Id: id, ProductId: productID, FileName: file.FileName,
		DisplayName: file.DisplayName, Format: generated.AdminProductFileFormat(file.Format), ContentType: file.ContentType,
		FileSizeBytes: file.FileSizeBytes, AvailabilityStatus: generated.AdminProductFileAvailabilityStatus(file.AvailabilityStatus),
		DisplayOrder: file.DisplayOrder, CreatedAt: &file.CreatedAt, UpdatedAt: &file.UpdatedAt}, nil
}

func (s *Server) GetProductFiles(ctx context.Context, request generated.GetProductFilesRequestObject) (generated.GetProductFilesResponseObject, error) {
	files, err := s.productFiles.List(ctx, request.ProductId.String())
	if err != nil {
		if errors.Is(err, productfiles.ErrInvalidInput) {
			return generated.GetProductFiles400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "invalid product id"))}, nil
		}
		if errors.Is(err, productfiles.ErrReferenceNotFound) {
			return generated.GetProductFiles404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("product"))}, nil
		}
		return nil, err
	}
	response := make(generated.GetProductFiles200JSONResponse, 0, len(files))
	for _, file := range files {
		item, err := toAPIProductFile(file)
		if err != nil {
			return nil, err
		}
		response = append(response, item)
	}
	return response, nil
}

func (s *Server) RequestProductFileUpload(ctx context.Context, request generated.RequestProductFileUploadRequestObject) (generated.RequestProductFileUploadResponseObject, error) {
	if request.Body == nil {
		return generated.RequestProductFileUpload400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "request body is required"))}, nil
	}
	a, err := s.productFiles.RequestUpload(ctx, productfiles.RequestUploadInput{ProductID: request.ProductId.String(), FileName: request.Body.FileName,
		Format: string(request.Body.Format), ContentType: request.Body.ContentType, FileSizeBytes: request.Body.FileSizeBytes})
	if err != nil {
		switch {
		case errors.Is(err, productfiles.ErrInvalidUpload), errors.Is(err, productfiles.ErrInvalidInput):
			return generated.RequestProductFileUpload400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "invalid file upload request"))}, nil
		case errors.Is(err, productfiles.ErrReferenceNotFound):
			return generated.RequestProductFileUpload404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("product"))}, nil
		case errors.Is(err, productfiles.ErrStorageFailure):
			return generated.RequestProductFileUpload502JSONResponse{BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError())}, nil
		case errors.Is(err, productfiles.ErrUploadUnavailable):
			return generated.RequestProductFileUpload503JSONResponse{ServiceUnavailableJSONResponse: generated.ServiceUnavailableJSONResponse(uploadUnavailableError())}, nil
		default:
			return nil, err
		}
	}
	return generated.RequestProductFileUpload201JSONResponse{UploadToken: a.UploadToken, UploadUrl: a.UploadURL,
		Headers: a.Headers, ExpiresAt: a.ExpiresAt}, nil
}

func (s *Server) ConfirmProductFileUpload(ctx context.Context, request generated.ConfirmProductFileUploadRequestObject) (generated.ConfirmProductFileUploadResponseObject, error) {
	if request.Body == nil {
		return generated.ConfirmProductFileUpload400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "request body is required"))}, nil
	}
	file, err := s.productFiles.ConfirmUpload(ctx, productfiles.ConfirmUploadInput{ProductID: request.ProductId.String(), UploadToken: request.Body.UploadToken,
		DisplayName: optionalString(request.Body.DisplayName), DisplayOrder: optionalInt(request.Body.DisplayOrder)})
	if err != nil {
		switch {
		case errors.Is(err, productfiles.ErrInvalidUpload), errors.Is(err, productfiles.ErrInvalidInput):
			return generated.ConfirmProductFileUpload400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "invalid file upload confirmation"))}, nil
		case errors.Is(err, productfiles.ErrReferenceNotFound), errors.Is(err, productfiles.ErrObjectNotFound):
			return generated.ConfirmProductFileUpload404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("product file"))}, nil
		case errors.Is(err, productfiles.ErrConflict), errors.Is(err, productfiles.ErrObjectMismatch):
			return generated.ConfirmProductFileUpload409JSONResponse{ConflictJSONResponse: generated.ConflictJSONResponse(fileError(generated.ErrorCodeConflict, "file upload conflicts with existing data"))}, nil
		case errors.Is(err, productfiles.ErrExpiredUpload):
			return generated.ConfirmProductFileUpload410JSONResponse{GoneJSONResponse: generated.GoneJSONResponse(fileError(generated.ErrorCodeExpiredUpload, "file upload authorization has expired"))}, nil
		case errors.Is(err, productfiles.ErrStorageFailure):
			return generated.ConfirmProductFileUpload502JSONResponse{BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError())}, nil
		case errors.Is(err, productfiles.ErrUploadUnavailable):
			return generated.ConfirmProductFileUpload503JSONResponse{ServiceUnavailableJSONResponse: generated.ServiceUnavailableJSONResponse(uploadUnavailableError())}, nil
		default:
			return nil, err
		}
	}
	apiFile, err := toAPIProductFile(file)
	if err != nil {
		return nil, err
	}
	return generated.ConfirmProductFileUpload201JSONResponse(apiFile), nil
}

func (s *Server) UpdateProductFileMetadata(ctx context.Context, request generated.UpdateProductFileMetadataRequestObject) (generated.UpdateProductFileMetadataResponseObject, error) {
	if request.Body == nil {
		return generated.UpdateProductFileMetadata400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "request body is required"))}, nil
	}
	file, err := s.productFiles.UpdateMetadata(ctx, productfiles.UpdateMetadataInput{ProductID: request.ProductId.String(), FileID: request.FileId.String(),
		DisplayName: request.Body.DisplayName, DisplayOrder: request.Body.DisplayOrder})
	if err != nil {
		if errors.Is(err, productfiles.ErrInvalidInput) {
			return generated.UpdateProductFileMetadata400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "invalid file metadata"))}, nil
		}
		if errors.Is(err, productfiles.ErrReferenceNotFound) || errors.Is(err, productfiles.ErrFileNotFound) {
			return generated.UpdateProductFileMetadata404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("product file"))}, nil
		}
		return nil, err
	}
	apiFile, err := toAPIProductFile(file)
	if err != nil {
		return nil, err
	}
	return generated.UpdateProductFileMetadata200JSONResponse(apiFile), nil
}

func (s *Server) DeleteProductFile(ctx context.Context, request generated.DeleteProductFileRequestObject) (generated.DeleteProductFileResponseObject, error) {
	err := s.productFiles.Delete(ctx, request.ProductId.String(), request.FileId.String())
	if err != nil {
		if errors.Is(err, productfiles.ErrInvalidInput) {
			return generated.DeleteProductFile400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "invalid file id"))}, nil
		}
		if errors.Is(err, productfiles.ErrReferenceNotFound) || errors.Is(err, productfiles.ErrFileNotFound) {
			return generated.DeleteProductFile404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("product file"))}, nil
		}
		if errors.Is(err, productfiles.ErrStorageFailure) {
			return generated.DeleteProductFile502JSONResponse{BadGatewayJSONResponse: generated.BadGatewayJSONResponse(storageUnavailableError())}, nil
		}
		return nil, err
	}
	return generated.DeleteProductFile204Response{}, nil
}

func (s *Server) ReorderProductFiles(ctx context.Context, request generated.ReorderProductFilesRequestObject) (generated.ReorderProductFilesResponseObject, error) {
	if request.Body == nil || request.Body.FileIds == nil {
		return generated.ReorderProductFiles400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "fileIds is required"))}, nil
	}
	ids := make([]string, len(request.Body.FileIds))
	for i, id := range request.Body.FileIds {
		ids[i] = id.String()
	}
	err := s.productFiles.Reorder(ctx, request.ProductId.String(), ids)
	if err != nil {
		if errors.Is(err, productfiles.ErrInvalidInput) {
			return generated.ReorderProductFiles400JSONResponse{BadRequestJSONResponse: generated.BadRequestJSONResponse(fileError(generated.ErrorCodeInvalidRequest, "fileIds must contain every current file exactly once"))}, nil
		}
		if errors.Is(err, productfiles.ErrReferenceNotFound) {
			return generated.ReorderProductFiles404JSONResponse{NotFoundJSONResponse: generated.NotFoundJSONResponse(notFoundError("product"))}, nil
		}
		return nil, err
	}
	return generated.ReorderProductFiles204Response{}, nil
}
