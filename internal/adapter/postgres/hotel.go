package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/netologist/hotel-booking/db/generated"
	"github.com/netologist/hotel-booking/internal/domain"
)

// HotelRepository interface
type HotelRepository interface {
	List(ctx context.Context) ([]*domain.Hotel, error)
}

// hotelRepository implements domain.HotelRepository via SQLC.
type hotelRepository struct {
	queries *db.Queries
}

// NewHotelRepository creates a new HotelRepository.
func NewHotelRepository(pool *pgxpool.Pool) HotelRepository {
	return &hotelRepository{queries: db.New(pool)}
}

func (h *hotelRepository) List(ctx context.Context) ([]*domain.Hotel, error) {
	rows, err := h.queries.ListHotels(ctx)
	if err != nil {
		return nil, err
	}
	hotels := make([]*domain.Hotel, len(rows))
	for i, r := range rows {
		hotels[i] = toDomainHotel(r)
	}
	return hotels, nil
}

// Model mappers
func toDomainHotel(row db.Hotel) *domain.Hotel {
	return &domain.Hotel{
		ID:        uuid.UUID(row.ID.Bytes),
		Name:      row.Name,
		Address:   row.Address,
		City:      row.City,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
