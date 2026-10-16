package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/netologist/go-service-template/internal/domain"
	itemv1 "github.com/netologist/go-service-template/internal/gen/grpc/item/v1"
	"github.com/netologist/go-service-template/internal/usecase"
)

// ItemHandler implements the gRPC ItemServiceServer interface.
type ItemHandler struct {
	itemv1.UnimplementedItemServiceServer
	itemService usecase.ItemService
	converter   ItemConverter
}

var _ itemv1.ItemServiceServer = (*ItemHandler)(nil)

// NewItemHandler constructs a new ItemHandler with the given usecase service.
func NewItemHandler(itemService usecase.ItemService) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
		converter:   &ItemConverterImpl{},
	}
}

// CreateItem registers a new item through the domain usecase layer.
func (h *ItemHandler) CreateItem(
	ctx context.Context,
	req *itemv1.CreateItemRequest,
) (*itemv1.CreateItemResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	model, err := h.itemService.CreateItem(ctx, req.GetName(), req.GetDescription(), req.GetPrice())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidItem) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, domain.ErrItemAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		if errors.Is(err, domain.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		slog.ErrorContext(ctx, "failed to create item via gRPC", "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &itemv1.CreateItemResponse{
		Item: h.converter.ToProtoItem(model),
	}, nil
}

// FindItems lists registered items and optionally filters by name query.
func (h *ItemHandler) FindItems(
	ctx context.Context,
	req *itemv1.FindItemsRequest,
) (*itemv1.FindItemsResponse, error) {
	items, err := h.itemService.FindItems(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		slog.ErrorContext(ctx, "failed to find items via gRPC", "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var nameFilter string
	if req != nil {
		nameFilter = strings.TrimSpace(req.GetName())
	}

	protoItems := make([]*itemv1.Item, 0, len(items))
	for _, it := range items {
		if nameFilter != "" && !strings.Contains(strings.ToLower(it.Name), strings.ToLower(nameFilter)) {
			continue
		}
		protoItems = append(protoItems, h.converter.ToProtoItem(it))
	}

	return &itemv1.FindItemsResponse{
		Items: protoItems,
	}, nil
}
