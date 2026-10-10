package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netologist/go-service-template/internal/domain"
	"github.com/netologist/go-service-template/internal/gen/sqlc"
)

const (
	testItemName        = "Mechanical Keyboard"
	testItemDescription = "Ergonomic keyboard"
	testItemPrice       = int64(9900)
)

type mockQuerier struct {
	mock.Mock
}

var _ sqlc.Querier = (*mockQuerier)(nil)

func (m *mockQuerier) CreateItem(ctx context.Context, arg sqlc.CreateItemParams) (sqlc.Item, error) {
	args := m.Called(ctx, arg)
	if row, ok := args.Get(0).(sqlc.Item); ok {
		return row, args.Error(1)
	}
	return sqlc.Item{}, args.Error(1)
}

func (m *mockQuerier) ListItems(ctx context.Context) ([]sqlc.Item, error) {
	args := m.Called(ctx)
	if rows, ok := args.Get(0).([]sqlc.Item); ok {
		return rows, args.Error(1)
	}
	return nil, args.Error(1)
}
func TestNewItemRepository(t *testing.T) {
	repo := NewItemRepository(nil)
	require.NotNil(t, repo)
}

func TestNewItemRepositoryWithQuerier(t *testing.T) {
	mockQ := &mockQuerier{}
	repo := NewItemRepositoryWithQuerier(mockQ)
	require.NotNil(t, repo)
}

func TestItemRepository_List_Success(t *testing.T) {
	t.Parallel()

	mockQ := &mockQuerier{}
	repo := NewItemRepositoryWithQuerier(mockQ)
	ctx := t.Context()

	id1 := uuid.New()
	id2 := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	sqlcRows := []sqlc.Item{
		{
			ID:          id1,
			Name:        "Item One",
			Description: "First item",
			Price:       1000,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
		{
			ID:          id2,
			Name:        "Item Two",
			Description: "Second item",
			Price:       2000,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		},
	}

	mockQ.On("ListItems", ctx).Return(sqlcRows, nil).Once()

	items, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, id1, items[0].ID)
	assert.Equal(t, "Item One", items[0].Name)
	assert.Equal(t, "First item", items[0].Description)
	assert.Equal(t, int64(1000), items[0].Price)
	assert.True(t, fixedTime.Equal(items[0].CreatedAt))

	assert.Equal(t, id2, items[1].ID)
	assert.Equal(t, "Item Two", items[1].Name)
	assert.Equal(t, "Second item", items[1].Description)
	assert.Equal(t, int64(2000), items[1].Price)
	assert.True(t, fixedTime.Equal(items[1].CreatedAt))

	mockQ.AssertExpectations(t)
}

func TestItemRepository_List_Empty(t *testing.T) {
	t.Parallel()

	mockQ := &mockQuerier{}
	repo := NewItemRepositoryWithQuerier(mockQ)
	ctx := t.Context()

	mockQ.On("ListItems", ctx).Return([]sqlc.Item{}, nil).Once()

	items, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)

	mockQ.AssertExpectations(t)
}

func TestItemRepository_List_Error(t *testing.T) {
	t.Parallel()

	mockQ := &mockQuerier{}
	repo := NewItemRepositoryWithQuerier(mockQ)
	ctx := t.Context()

	dbErr := errors.New("database connection failed")
	mockQ.On("ListItems", ctx).Return(nil, dbErr).Once()

	items, err := repo.List(ctx)
	require.Error(t, err)
	assert.Nil(t, items)
	require.ErrorIs(t, err, dbErr)

	mockQ.AssertExpectations(t)
}

func TestItemRepository_Create_Success(t *testing.T) {
	t.Parallel()

	mockQ := &mockQuerier{}
	repo := NewItemRepositoryWithQuerier(mockQ)
	ctx := t.Context()

	itemID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	domainItem := domain.Item{
		ID:          itemID,
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	}

	expectedParams := sqlc.CreateItemParams{
		ID:          itemID,
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	}

	sqlcRow := sqlc.Item{
		ID:          itemID,
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	mockQ.On("CreateItem", ctx, expectedParams).Return(sqlcRow, nil).Once()

	created, err := repo.Create(ctx, domainItem)
	require.NoError(t, err)

	assert.Equal(t, itemID, created.ID)
	assert.Equal(t, testItemName, created.Name)
	assert.Equal(t, testItemDescription, created.Description)
	assert.Equal(t, testItemPrice, created.Price)
	assert.True(t, fixedTime.Equal(created.CreatedAt))

	mockQ.AssertExpectations(t)
}

func TestItemRepository_Create_Error(t *testing.T) {
	t.Parallel()

	mockQ := &mockQuerier{}
	repo := NewItemRepositoryWithQuerier(mockQ)
	ctx := t.Context()

	itemID := uuid.New()
	domainItem := domain.Item{
		ID:          itemID,
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	}

	dbErr := errors.New("disk full")
	mockQ.On("CreateItem", ctx, mock.Anything).Return(nil, dbErr).Once()

	created, err := repo.Create(ctx, domainItem)
	require.Error(t, err)
	require.ErrorIs(t, err, dbErr)
	assert.Equal(t, domain.Item{}, created)

	mockQ.AssertExpectations(t)
}

func TestItemRepository_Create_UniqueViolation(t *testing.T) {
	t.Parallel()

	mockQ := &mockQuerier{}
	repo := NewItemRepositoryWithQuerier(mockQ)
	ctx := t.Context()

	domainItem := domain.Item{
		ID:          uuid.New(),
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
	}

	pgErr := &pgconn.PgError{
		Code:    "23505",
		Message: "duplicate key value violates unique constraint",
	}

	mockQ.On("CreateItem", ctx, mock.Anything).Return(nil, pgErr).Once()

	created, err := repo.Create(ctx, domainItem)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrItemAlreadyExists)
	assert.Equal(t, domain.Item{}, created)

	mockQ.AssertExpectations(t)
}

func TestToDomainItem(t *testing.T) {
	t.Parallel()

	itemID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	row := sqlc.Item{
		ID:          itemID,
		Name:        testItemName,
		Description: testItemDescription,
		Price:       testItemPrice,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}

	item := toDomainItem(row)

	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, testItemName, item.Name)
	assert.Equal(t, testItemDescription, item.Description)
	assert.Equal(t, testItemPrice, item.Price)
	assert.Equal(t, fixedTime, item.CreatedAt)
	assert.Equal(t, fixedTime, item.UpdatedAt)
}
