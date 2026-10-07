package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Hotel represents a named accommodation establishment.
type Hotel struct {
	ID        uuid.UUID
	Name      string
	Address   string
	City      string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// HotelRepository interface
type HotelRepository interface {
	List(ctx context.Context) ([]*Hotel, error)
	Create(ctx context.Context, hotel *Hotel) (*Hotel, error)
}
