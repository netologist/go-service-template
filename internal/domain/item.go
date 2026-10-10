package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinNameLength        = 2
	MaxNameLength        = 255
	MaxDescriptionLength = 1000
)

// Item represents a domain entity for catalog or business items.
type Item struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewItem validates and creates a new Item domain entity enforcing invariants.
func NewItem(name, description string, price int64) (Item, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	nameLen := utf8.RuneCountInString(name)
	if nameLen < MinNameLength || nameLen > MaxNameLength {
		return Item{}, fmt.Errorf(
			"%w: name length must be between %d and %d characters",
			ErrInvalidItem, MinNameLength, MaxNameLength,
		)
	}

	descLen := utf8.RuneCountInString(description)
	if descLen > MaxDescriptionLength {
		return Item{}, fmt.Errorf(
			"%w: description length must not exceed %d characters",
			ErrInvalidItem, MaxDescriptionLength,
		)
	}

	if price < 0 {
		return Item{}, fmt.Errorf(
			"%w: price must be non-negative",
			ErrInvalidItem,
		)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Item{}, fmt.Errorf("generate id: %w", err)
	}

	return Item{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
	}, nil
}

// ItemRepository defines the outbound port for item persistence operations.
type ItemRepository interface {
	List(ctx context.Context) ([]Item, error)
	Create(ctx context.Context, item Item) (Item, error)
}
