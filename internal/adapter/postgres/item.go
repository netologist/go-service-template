package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/netologist/go-service-template/internal/domain"
	"github.com/netologist/go-service-template/internal/gen/goventer"
	"github.com/netologist/go-service-template/internal/gen/sqlc"
)

// ItemRepository implements domain.ItemRepository via SQLC.
type itemRepository struct {
	queries   sqlc.Querier
	converter goventer.ItemConverter
}

var _ domain.ItemRepository = (*itemRepository)(nil)

// NewItemRepository creates a new itemRepository.
func NewItemRepository(pool *pgxpool.Pool) *itemRepository {
	return NewItemRepositoryWithQuerier(sqlc.New(pool))
}

// NewItemRepositoryWithQuerier creates a new itemRepository with a custom Querier.
func NewItemRepositoryWithQuerier(queries sqlc.Querier) *itemRepository {
	return &itemRepository{
		queries:   queries,
		converter: goventer.NewItemConverter(),
	}
}

func (h *itemRepository) List(ctx context.Context) ([]domain.Item, error) {
	rows, err := h.queries.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	return h.converter.ToDomainItemList(rows), nil
}

func (h *itemRepository) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	row, err := h.queries.CreateItem(ctx, sqlc.CreateItemParams{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Item{}, fmt.Errorf("%w: %w", domain.ErrItemAlreadyExists, err)
		}
		return domain.Item{}, err
	}
	return h.converter.ToDomainItem(row), nil
}

// Model mappers
func toDomainItem(row sqlc.Item) domain.Item {
	return goventer.NewItemConverter().ToDomainItem(row)
}
