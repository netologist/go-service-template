package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netologist/go-service-template/internal/domain"
	domainmocks "github.com/netologist/go-service-template/internal/gen/mocks/domain"
	"github.com/netologist/go-service-template/internal/usecase"
)

const (
	testItemName        = "Mechanical Keyboard"
	testItemDescription = "Ergonomic keyboard"
	testItemPrice       = int64(9900)
)

func TestNewItemService(t *testing.T) {
	mockRepo := domainmocks.NewMockItemRepository(t)
	service := usecase.NewItemService(mockRepo)
	require.NotNil(t, service)
}

func TestItemService_FindItems_Success(t *testing.T) {
	mockRepo := domainmocks.NewMockItemRepository(t)
	service := usecase.NewItemService(mockRepo)
	ctx := t.Context()

	item1ID := uuid.New()
	item2ID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	domainItems := []domain.Item{
		{
			ID:          item1ID,
			Name:        "Item One",
			Description: "First description",
			Price:       1000,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
		{
			ID:          item2ID,
			Name:        "Item Two",
			Description: "Second description",
			Price:       2000,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
	}

	mockRepo.EXPECT().List(ctx).Return(domainItems, nil).Once()

	output, err := service.FindItems(ctx)
	require.NoError(t, err)
	require.Len(t, output, 2)

	assert.Equal(t, item1ID, output[0].ID)
	assert.Equal(t, "Item One", output[0].Name)
	assert.Equal(t, "First description", output[0].Description)
	assert.Equal(t, int64(1000), output[0].Price)
	assert.True(t, fixedTime.Equal(output[0].CreatedAt))

	assert.Equal(t, item2ID, output[1].ID)
	assert.Equal(t, "Item Two", output[1].Name)
	assert.Equal(t, "Second description", output[1].Description)
	assert.Equal(t, int64(2000), output[1].Price)
	assert.True(t, fixedTime.Equal(output[1].CreatedAt))
}

func TestItemService_FindItems_Empty(t *testing.T) {
	mockRepo := domainmocks.NewMockItemRepository(t)
	service := usecase.NewItemService(mockRepo)
	ctx := t.Context()

	mockRepo.EXPECT().List(ctx).Return([]domain.Item{}, nil).Once()

	output, err := service.FindItems(ctx)
	require.NoError(t, err)
	assert.Empty(t, output)
}

func TestItemService_FindItems_Error(t *testing.T) {
	mockRepo := domainmocks.NewMockItemRepository(t)
	service := usecase.NewItemService(mockRepo)
	ctx := t.Context()

	expectedErr := errors.New("repository read error")
	mockRepo.EXPECT().List(ctx).Return(nil, expectedErr).Once()

	output, err := service.FindItems(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, output)
}

func TestItemService_CreateItem_Success(t *testing.T) {
	mockRepo := domainmocks.NewMockItemRepository(t)
	service := usecase.NewItemService(mockRepo)
	ctx := t.Context()

	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	mockRepo.EXPECT().
		Create(ctx, mock.MatchedBy(func(item domain.Item) bool {
			return item.Name == testItemName &&
				item.Description == testItemDescription &&
				item.Price == testItemPrice &&
				item.ID != uuid.Nil
		})).
		RunAndReturn(func(_ context.Context, item domain.Item) (domain.Item, error) {
			item.CreatedAt = fixedTime
			item.UpdatedAt = fixedTime
			return item, nil
		}).
		Once()

	output, err := service.CreateItem(ctx, testItemName, testItemDescription, testItemPrice)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, output.ID)
	assert.Equal(t, testItemName, output.Name)
	assert.Equal(t, testItemDescription, output.Description)
	assert.Equal(t, testItemPrice, output.Price)
	assert.True(t, fixedTime.Equal(output.CreatedAt))
	assert.True(t, fixedTime.Equal(output.UpdatedAt))
}

func TestItemService_CreateItem_Error(t *testing.T) {
	mockRepo := domainmocks.NewMockItemRepository(t)
	service := usecase.NewItemService(mockRepo)
	ctx := t.Context()

	expectedErr := errors.New("database insert failure")
	mockRepo.EXPECT().
		Create(ctx, mock.Anything).
		Return(domain.Item{}, expectedErr).
		Once()

	output, err := service.CreateItem(ctx, testItemName, testItemDescription, testItemPrice)
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	assert.Equal(t, domain.Item{}, output)
}

func TestItemService_CreateItem_DomainInvariantFailure(t *testing.T) {
	mockRepo := domainmocks.NewMockItemRepository(t)
	service := usecase.NewItemService(mockRepo)
	ctx := t.Context()

	output, err := service.CreateItem(ctx, "A", testItemDescription, testItemPrice)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrInvalidItem)
	assert.Equal(t, domain.Item{}, output)
}
