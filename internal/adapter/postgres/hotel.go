package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netologist/hotel-booking/internal/domain"
	"github.com/netologist/hotel-booking/internal/gen/sqlc"
)

// NewHotelRepository creates a new HotelRepository.
func NewHotelRepository(pool *pgxpool.Pool) domain.HotelRepository {
	return &hotelRepository{queries: sqlc.New(pool)}
}

// hotelRepository implements domain.HotelRepository via SQLC.
type hotelRepository struct {
	queries *sqlc.Queries
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
func toDomainHotel(row sqlc.Hotel) *domain.Hotel {
	return &domain.Hotel{
		ID:        uuid.UUID(row.ID.Bytes),
		Name:      row.Name,
		Address:   row.Address,
		City:      row.City,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
