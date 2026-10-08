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

	"github.com/netologist/hotel-booking/internal/domain"
	domainmocks "github.com/netologist/hotel-booking/internal/gen/mocks/domain"
	"github.com/netologist/hotel-booking/internal/usecase"
)

const (
	testHotelName    = "Grand Budapest Hotel"
	testHotelAddress = "100 Alpine Way"
	testHotelCity    = "Lutz"
)

func TestNewHotelService(t *testing.T) {
	mockRepo := domainmocks.NewMockHotelRepository(t)
	service := usecase.NewHotelService(mockRepo)
	require.NotNil(t, service)
}

func TestHotelService_FindHotels_Success(t *testing.T) {
	mockRepo := domainmocks.NewMockHotelRepository(t)
	service := usecase.NewHotelService(mockRepo)
	ctx := context.Background()

	hotel1ID := uuid.New()
	hotel2ID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	domainHotels := []domain.Hotel{
		{
			ID:        hotel1ID,
			Name:      "Hotel One",
			Address:   "123 First St",
			City:      "Miami",
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
		{
			ID:        hotel2ID,
			Name:      "Hotel Two",
			Address:   "456 Second St",
			City:      "Orlando",
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
	}

	mockRepo.EXPECT().List(ctx).Return(domainHotels, nil).Once()

	output, err := service.FindHotels(ctx)
	require.NoError(t, err)
	require.Len(t, output, 2)

	assert.Equal(t, hotel1ID, output[0].ID)
	assert.Equal(t, "Hotel One", output[0].Name)
	assert.Equal(t, "123 First St", output[0].Address)
	assert.Equal(t, "Miami", output[0].City)
	assert.True(t, fixedTime.Equal(output[0].CreatedAt))

	assert.Equal(t, hotel2ID, output[1].ID)
	assert.Equal(t, "Hotel Two", output[1].Name)
	assert.Equal(t, "456 Second St", output[1].Address)
	assert.Equal(t, "Orlando", output[1].City)
	assert.True(t, fixedTime.Equal(output[1].CreatedAt))
}

func TestHotelService_FindHotels_Empty(t *testing.T) {
	mockRepo := domainmocks.NewMockHotelRepository(t)
	service := usecase.NewHotelService(mockRepo)
	ctx := context.Background()

	mockRepo.EXPECT().List(ctx).Return([]domain.Hotel{}, nil).Once()

	output, err := service.FindHotels(ctx)
	require.NoError(t, err)
	assert.Empty(t, output)
}

func TestHotelService_FindHotels_Error(t *testing.T) {
	mockRepo := domainmocks.NewMockHotelRepository(t)
	service := usecase.NewHotelService(mockRepo)
	ctx := context.Background()
	repoErr := errors.New("repository list failed")

	mockRepo.EXPECT().List(ctx).Return(nil, repoErr).Once()

	output, err := service.FindHotels(ctx)
	require.ErrorIs(t, err, repoErr)
	assert.Nil(t, output)
}

func TestHotelService_CreateHotel_Success(t *testing.T) {
	mockRepo := domainmocks.NewMockHotelRepository(t)
	service := usecase.NewHotelService(mockRepo)
	ctx := context.Background()

	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	var capturedID uuid.UUID
	mockRepo.EXPECT().Create(ctx, mock.MatchedBy(func(h domain.Hotel) bool {
		capturedID = h.ID
		return h.Name == testHotelName &&
			h.Address == testHotelAddress &&
			h.City == testHotelCity &&
			h.ID != uuid.Nil
	})).RunAndReturn(func(_ context.Context, h domain.Hotel) (domain.Hotel, error) {
		return domain.Hotel{
			ID:        h.ID,
			Name:      h.Name,
			Address:   h.Address,
			City:      h.City,
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		}, nil
	}).Once()

	res, err := service.CreateHotel(ctx, testHotelName, testHotelAddress, testHotelCity)
	require.NoError(t, err)

	assert.Equal(t, capturedID, res.ID)
	assert.Equal(t, testHotelName, res.Name)
	assert.Equal(t, testHotelAddress, res.Address)
	assert.Equal(t, testHotelCity, res.City)
	assert.True(t, fixedTime.Equal(res.CreatedAt))
	assert.True(t, fixedTime.Equal(res.UpdatedAt))
}

func TestHotelService_CreateHotel_Error(t *testing.T) {
	mockRepo := domainmocks.NewMockHotelRepository(t)
	service := usecase.NewHotelService(mockRepo)
	ctx := context.Background()
	repoErr := errors.New("repository connection failed")

	mockRepo.EXPECT().Create(ctx, mock.MatchedBy(func(h domain.Hotel) bool {
		return h.Name == testHotelName &&
			h.Address == testHotelAddress &&
			h.City == testHotelCity &&
			h.ID != uuid.Nil
	})).Return(domain.Hotel{}, repoErr).Once()

	res, err := service.CreateHotel(ctx, testHotelName, testHotelAddress, testHotelCity)
	require.ErrorIs(t, err, repoErr)
	require.ErrorContains(t, err, "create hotel")
	assert.Equal(t, uuid.Nil, res.ID)
}

func TestHotelService_CreateHotel_DomainInvariantFailure(t *testing.T) {
	mockRepo := domainmocks.NewMockHotelRepository(t)
	service := usecase.NewHotelService(mockRepo)
	ctx := context.Background()

	res, err := service.CreateHotel(ctx, " ", testHotelAddress, testHotelCity)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrInvalidHotel)
	assert.Equal(t, uuid.Nil, res.ID)
}
