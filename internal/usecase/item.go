package usecase

import (
	"context"
	"fmt"

	"github.com/netologist/go-service-template/internal/domain"
)

// ItemService is the inbound port for item business operations.
type ItemService interface {
	FindItems(ctx context.Context) ([]domain.Item, error)
	CreateItem(ctx context.Context, name, description string, price int64) (domain.Item, error)
}

// itemService implements ItemService.
type itemService struct {
	itemRepo domain.ItemRepository
}

var _ ItemService = (*itemService)(nil)

// NewItemService constructor returns a concrete *itemService.
func NewItemService(itemRepo domain.ItemRepository) *itemService {
	return &itemService{
		itemRepo: itemRepo,
	}
}

func (u *itemService) FindItems(ctx context.Context) ([]domain.Item, error) {
	return u.itemRepo.List(ctx)
}

func (u *itemService) CreateItem(ctx context.Context, name, description string, price int64) (domain.Item, error) {
	item, err := domain.NewItem(name, description, price)
	if err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}

	entity, err := u.itemRepo.Create(ctx, item)
	if err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}

	return entity, nil
}
