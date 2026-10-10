package handler_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netologist/go-service-template/internal/adapter/httpapi/handler"
	"github.com/netologist/go-service-template/internal/domain"
	usecasemocks "github.com/netologist/go-service-template/internal/gen/mocks/usecase"
	"github.com/netologist/go-service-template/internal/gen/openapi"
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

	desc := testItemDescription
	resp, err := h.CreateItem(t.Context(), openapi.CreateItemRequestObject{
		Body: &openapi.CreateItemJSONRequestBody{
			Name:        testItemName,
			Description: &desc,
			Price:       testItemPrice,
		},
	})

	require.NoError(t, err)
	successResp, ok := resp.(openapi.CreateItem201JSONResponse)
	require.True(t, ok)
	assert.Equal(t, itemID, successResp.Id)
	assert.Equal(t, testItemName, successResp.Name)
	assert.Equal(t, testItemDescription, successResp.Description)
	assert.Equal(t, testItemPrice, successResp.Price)
}

func TestItemHandler_CreateItem_ValidationError(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	validationErr := domain.ErrInvalidItem
	mockService.EXPECT().CreateItem(mock.Anything, "A", "", int64(-1)).Return(domain.Item{}, validationErr).Once()

	resp, err := h.CreateItem(t.Context(), openapi.CreateItemRequestObject{
		Body: &openapi.CreateItemJSONRequestBody{
			Name:  "A",
			Price: -1,
		},
	})

	require.NoError(t, err)
	badReqResp, ok := resp.(openapi.CreateItem400ApplicationProblemPlusJSONResponse)
	require.True(t, ok)
	assert.Equal(t, 400, badReqResp.Status)
	assert.Equal(t, "INVALID_ITEM", *badReqResp.Code)
}

func TestItemHandler_CreateItem_Conflict(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	conflictErr := domain.ErrItemAlreadyExists
	mockService.EXPECT().
		CreateItem(mock.Anything, testItemName, "", testItemPrice).
		Return(domain.Item{}, conflictErr).
		Once()

	resp, err := h.CreateItem(t.Context(), openapi.CreateItemRequestObject{
		Body: &openapi.CreateItemJSONRequestBody{
			Name:  testItemName,
			Price: testItemPrice,
		},
	})

	require.NoError(t, err)
	conflictResp, ok := resp.(openapi.CreateItem409ApplicationProblemPlusJSONResponse)
	require.True(t, ok)
	assert.Equal(t, 409, conflictResp.Status)
	assert.Equal(t, "ITEM_ALREADY_EXISTS", *conflictResp.Code)
}

func TestItemHandler_CreateItem_InternalError(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	internalErr := errors.New("db failure")
	mockService.EXPECT().
		CreateItem(mock.Anything, testItemName, "", testItemPrice).
		Return(domain.Item{}, internalErr).
		Once()

	resp, err := h.CreateItem(t.Context(), openapi.CreateItemRequestObject{
		Body: &openapi.CreateItemJSONRequestBody{
			Name:  testItemName,
			Price: testItemPrice,
		},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestItemHandler_FindItems_Success(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	items := []domain.Item{
		{
			ID:          uuid.New(),
			Name:        "Item 1",
			Description: "Desc 1",
			Price:       1000,
		},
	}

	mockService.EXPECT().FindItems(mock.Anything).Return(items, nil).Once()

	resp, err := h.FindItems(t.Context(), openapi.FindItemsRequestObject{})

	require.NoError(t, err)
	listResp, ok := resp.(openapi.FindItems200JSONResponse)
	require.True(t, ok)
	assert.Len(t, listResp, 1)
	assert.Equal(t, "Item 1", listResp[0].Name)
}

func TestItemHandler_FindItems_Error(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	h := handler.NewItemHandler(mockService)

	mockService.EXPECT().FindItems(mock.Anything).Return(nil, errors.New("db failure")).Once()

	resp, err := h.FindItems(t.Context(), openapi.FindItemsRequestObject{})

	require.Error(t, err)
	assert.Nil(t, resp)
}
