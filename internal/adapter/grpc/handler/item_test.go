package handler_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/netologist/go-service-template/internal/adapter/grpc/handler"
	"github.com/netologist/go-service-template/internal/domain"
	itemv1 "github.com/netologist/go-service-template/internal/gen/grpc/item/v1"
	usecasemocks "github.com/netologist/go-service-template/internal/gen/mocks/usecase"
)

const (
	testItemName        = "Mechanical Keyboard"
	testItemDescription = "Ergonomic hot-swappable mechanical keyboard"
	testItemPrice       = int64(9900)
)

func TestNewItemHandler(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)
	require.NotNil(t, h)
}

func TestItemHandler_CreateItem_Success(t *testing.T) {
	itemID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().CreateItem(mock.Anything, testItemName, testItemDescription, testItemPrice).Return(domain.Item{
		ID:          itemID,
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}, nil).Once()

	resp, err := h.CreateItem(t.Context(), &itemv1.CreateItemRequest{
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Item)
	assert.Equal(t, itemID.String(), resp.Item.Id)
	assert.Equal(t, testItemName, resp.Item.Name)
	assert.Equal(t, testItemDescription, resp.Item.Description)
	assert.Equal(t, testItemPrice, resp.Item.Price)
	assert.Equal(t, fixedTime.Unix(), resp.Item.CreatedAt.Seconds)
}

func TestItemHandler_CreateItem_NilRequest(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	resp, err := h.CreateItem(t.Context(), nil)
	require.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestItemHandler_CreateItem_ValidationError(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().CreateItem(mock.Anything, "A", "", int64(10)).Return(
		domain.Item{},
		domain.ErrInvalidItem,
	).Once()

	resp, err := h.CreateItem(t.Context(), &itemv1.CreateItemRequest{
		Name:        "A",
		Description: "",
		Price:       10,
	})

	require.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, err.Error(), domain.ErrInvalidItem.Error())
}

func TestItemHandler_CreateItem_Conflict(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().CreateItem(mock.Anything, testItemName, testItemDescription, testItemPrice).Return(
		domain.Item{},
		domain.ErrItemAlreadyExists,
	).Once()

	resp, err := h.CreateItem(t.Context(), &itemv1.CreateItemRequest{
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	})

	require.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.Contains(t, err.Error(), domain.ErrItemAlreadyExists.Error())
}

func TestItemHandler_CreateItem_NotFound(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().CreateItem(mock.Anything, testItemName, testItemDescription, testItemPrice).Return(
		domain.Item{},
		domain.ErrItemNotFound,
	).Once()

	resp, err := h.CreateItem(t.Context(), &itemv1.CreateItemRequest{
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	})

	require.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestItemHandler_CreateItem_InternalError(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().CreateItem(mock.Anything, testItemName, testItemDescription, testItemPrice).Return(
		domain.Item{},
		errors.New("unexpected database error"),
	).Once()

	resp, err := h.CreateItem(t.Context(), &itemv1.CreateItemRequest{
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	})

	require.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "internal server error")
}

func TestItemHandler_FindItems_Success(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	items := []domain.Item{
		{
			ID:          uuid.New(),
			Name:        "Mechanical Keyboard",
			Description: "Clicky",
			Price:       9900,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			ID:          uuid.New(),
			Name:        "Mouse Pad",
			Description: "Smooth",
			Price:       1500,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	t.Run("all items", func(t *testing.T) {
		mockService.EXPECT().FindItems(mock.Anything).Return(items, nil).Once()

		resp, err := h.FindItems(t.Context(), &itemv1.FindItemsRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Items, 2)
	})

	t.Run("filtered by name", func(t *testing.T) {
		mockService.EXPECT().FindItems(mock.Anything).Return(items, nil).Once()

		resp, err := h.FindItems(t.Context(), &itemv1.FindItemsRequest{
			Name: "Keyboard",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Items, 1)
		assert.Equal(t, "Mechanical Keyboard", resp.Items[0].Name)
	})
}

func TestItemHandler_FindItems_Error(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().FindItems(mock.Anything).Return(nil, errors.New("db error")).Once()

	resp, err := h.FindItems(t.Context(), &itemv1.FindItemsRequest{})
	require.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestItemHandler_FindItems_NotFound(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().FindItems(mock.Anything).Return(nil, domain.ErrItemNotFound).Once()

	resp, err := h.FindItems(t.Context(), &itemv1.FindItemsRequest{})
	require.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
