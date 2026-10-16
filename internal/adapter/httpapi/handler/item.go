package handler

import (
	"context"
	"errors"

	"github.com/netologist/go-service-template/internal/adapter/httpapi/problem"
	"github.com/netologist/go-service-template/internal/domain"
	"github.com/netologist/go-service-template/internal/gen/goventer"
	"github.com/netologist/go-service-template/internal/gen/openapi"
	"github.com/netologist/go-service-template/internal/usecase"
)

func NewItemHandler(itemService usecase.ItemService) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
		converter:   goventer.NewItemConverter(),
	}
}

type ItemHandler struct {
	itemService usecase.ItemService
	converter   goventer.ItemConverter
}

func (h *ItemHandler) CreateItem(
	ctx context.Context,
	request openapi.CreateItemRequestObject,
) (openapi.CreateItemResponseObject, error) {
	var desc string
	if request.Body.Description != nil {
		desc = *request.Body.Description
	}

	model, err := h.itemService.CreateItem(ctx, request.Body.Name, desc, request.Body.Price)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidItem) {
			prob := problem.BadRequest(err.Error(), "/api/v1/items", "INVALID_ITEM", nil)
			return openapi.CreateItem400ApplicationProblemPlusJSONResponse{
				BadRequestApplicationProblemPlusJSONResponse: openapi.BadRequestApplicationProblemPlusJSONResponse(prob),
			}, nil
		}
		if errors.Is(err, domain.ErrItemAlreadyExists) {
			prob := problem.Conflict(domain.ErrItemAlreadyExists.Error(), "/api/v1/items", "ITEM_ALREADY_EXISTS")
			return openapi.CreateItem409ApplicationProblemPlusJSONResponse{
				ConflictApplicationProblemPlusJSONResponse: openapi.ConflictApplicationProblemPlusJSONResponse(prob),
			}, nil
		}
		return nil, err
	}

	return openapi.CreateItem201JSONResponse(h.converter.ToOpenAPIItem(model)), nil
}

func (h *ItemHandler) FindItems(
	ctx context.Context,
	request openapi.FindItemsRequestObject,
) (openapi.FindItemsResponseObject, error) {
	model, err := h.itemService.FindItems(ctx)
	if err != nil {
		return nil, err
	}

	return openapi.FindItems200JSONResponse(h.converter.ToOpenAPIItemList(model)), nil
}
