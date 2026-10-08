package postgres

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
	"github.com/netologist/hotel-booking/internal/gen/sqlc"
)

const (
	testHotelName    = "Grand Budapest Hotel"
	testHotelAddress = "100 Alpine Way"
	testHotelCity    = "Lutz"
)

type mockQuerier struct {
	mock.Mock
}

var _ sqlc.Querier = (*mockQuerier)(nil)

func (m *mockQuerier) CreateHotel(ctx context.Context, arg sqlc.CreateHotelParams) (sqlc.Hotel, error) {
	args := m.Called(ctx, arg)
	if row, ok := args.Get(0).(sqlc.Hotel); ok {
		return row, args.Error(1)
	}
	return sqlc.Hotel{}, args.Error(1)
}

func (m *mockQuerier) ListHotels(ctx context.Context) ([]sqlc.Hotel, error) {
	args := m.Called(ctx)
	if list, ok := args.Get(0).([]sqlc.Hotel); ok {
		return list, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestNewHotelRepository(t *testing.T) {
	repo := NewHotelRepository(nil)
	require.NotNil(t, repo)
}

func TestNewHotelRepositoryWithQuerier(t *testing.T) {
	q := &mockQuerier{}
	repo := NewHotelRepositoryWithQuerier(q)
	require.NotNil(t, repo)
}

func TestHotelRepository_List_Success(t *testing.T) {
	q := &mockQuerier{}
	repo := NewHotelRepositoryWithQuerier(q)
	ctx := context.Background()

	hotel1ID := uuid.New()
	hotel2ID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	rows := []sqlc.Hotel{
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

	q.On("ListHotels", ctx).Return(rows, nil).Once()

	hotels, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, hotels, 2)

	assert.Equal(t, hotel1ID, hotels[0].ID)
	assert.Equal(t, "Hotel One", hotels[0].Name)
	assert.Equal(t, "123 First St", hotels[0].Address)
	assert.Equal(t, "Miami", hotels[0].City)
	assert.True(t, fixedTime.Equal(hotels[0].CreatedAt))

	assert.Equal(t, hotel2ID, hotels[1].ID)
	assert.Equal(t, "Hotel Two", hotels[1].Name)
	assert.Equal(t, "456 Second St", hotels[1].Address)
	assert.Equal(t, "Orlando", hotels[1].City)
	assert.True(t, fixedTime.Equal(hotels[1].CreatedAt))

	q.AssertExpectations(t)
}

func TestHotelRepository_List_Empty(t *testing.T) {
	q := &mockQuerier{}
	repo := NewHotelRepositoryWithQuerier(q)
	ctx := context.Background()

	q.On("ListHotels", ctx).Return([]sqlc.Hotel{}, nil).Once()

	hotels, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, hotels)

	q.AssertExpectations(t)
}

func TestHotelRepository_List_Error(t *testing.T) {
	q := &mockQuerier{}
	repo := NewHotelRepositoryWithQuerier(q)
	ctx := context.Background()
	dbErr := errors.New("database connection failed")

	q.On("ListHotels", ctx).Return(nil, dbErr).Once()

	hotels, err := repo.List(ctx)
	require.ErrorIs(t, err, dbErr)
	assert.Nil(t, hotels)

	q.AssertExpectations(t)
}

func TestHotelRepository_Create_Success(t *testing.T) {
	q := &mockQuerier{}
	repo := NewHotelRepositoryWithQuerier(q)
	ctx := context.Background()

	hotelID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	input := domain.Hotel{
		ID:      hotelID,
		Name:    testHotelName,
		Address: testHotelAddress,
		City:    testHotelCity,
	}

	expectedArg := sqlc.CreateHotelParams{
		ID:      hotelID,
		Name:    testHotelName,
		Address: testHotelAddress,
		City:    testHotelCity,
	}

	createdRow := sqlc.Hotel{
		ID:        hotelID,
		Name:      testHotelName,
		Address:   testHotelAddress,
		City:      testHotelCity,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}
	q.On("CreateHotel", ctx, expectedArg).Return(createdRow, nil).Once()

	res, err := repo.Create(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, hotelID, res.ID)
	assert.Equal(t, testHotelName, res.Name)
	assert.Equal(t, testHotelAddress, res.Address)
	assert.Equal(t, testHotelCity, res.City)
	assert.True(t, fixedTime.Equal(res.CreatedAt))
	assert.True(t, fixedTime.Equal(res.UpdatedAt))

	q.AssertExpectations(t)
}

func TestHotelRepository_Create_Error(t *testing.T) {
	q := &mockQuerier{}
	repo := NewHotelRepositoryWithQuerier(q)
	ctx := context.Background()
	dbErr := errors.New("unique constraint violation")

	hotelID := uuid.New()
	input := domain.Hotel{
		ID:      hotelID,
		Name:    testHotelName,
		Address: testHotelAddress,
		City:    testHotelCity,
	}

	expectedArg := sqlc.CreateHotelParams{
		ID:      hotelID,
		Name:    testHotelName,
		Address: testHotelAddress,
		City:    testHotelCity,
	}
	q.On("CreateHotel", ctx, expectedArg).Return(sqlc.Hotel{}, dbErr).Once()

	res, err := repo.Create(ctx, input)
	require.ErrorIs(t, err, dbErr)
	assert.Zero(t, res)

	q.AssertExpectations(t)
}

func TestToDomainHotel(t *testing.T) {
	hotelID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	row := sqlc.Hotel{
		ID:        hotelID,
		Name:      testHotelName,
		Address:   testHotelAddress,
		City:      testHotelCity,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}
	got := toDomainHotel(row)
	require.NotNil(t, got)

	assert.Equal(t, hotelID, got.ID)
	assert.Equal(t, testHotelName, got.Name)
	assert.Equal(t, testHotelAddress, got.Address)
	assert.Equal(t, testHotelCity, got.City)
	assert.True(t, fixedTime.Equal(got.CreatedAt))
	assert.True(t, fixedTime.Equal(got.UpdatedAt))
}
