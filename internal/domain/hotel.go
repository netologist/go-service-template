package domain

import (
	"time"

	"github.com/google/uuid"
)

// Hotel represents a named accommodation establishment.
type Hotel struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
