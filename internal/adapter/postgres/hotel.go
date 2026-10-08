package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netologist/hotel-booking/internal/domain"
	"github.com/netologist/hotel-booking/internal/gen/sqlc"
)

// HotelRepository implements domain.HotelRepository via SQLC.
type hotelRepository struct {
	queries sqlc.Querier
}

var _ domain.HotelRepository = (*hotelRepository)(nil)

// NewHotelRepository creates a new hotelRepository.
func NewHotelRepository(pool *pgxpool.Pool) *hotelRepository {
	return NewHotelRepositoryWithQuerier(sqlc.New(pool))
}

// NewHotelRepositoryWithQuerier creates a new hotelRepository with a custom Querier.
func NewHotelRepositoryWithQuerier(queries sqlc.Querier) *hotelRepository {
	return &hotelRepository{queries: queries}
}

func (h *hotelRepository) List(ctx context.Context) ([]domain.Hotel, error) {
	rows, err := h.queries.ListHotels(ctx)
	if err != nil {
		return nil, err
	}
	hotels := make([]domain.Hotel, len(rows))
	for i, r := range rows {
		hotels[i] = toDomainHotel(r)
	}
	return hotels, nil
}

func (h *hotelRepository) Create(ctx context.Context, hotel domain.Hotel) (domain.Hotel, error) {
	row, err := h.queries.CreateHotel(ctx, sqlc.CreateHotelParams{
		ID:      hotel.ID,
		Name:    hotel.Name,
		Address: hotel.Address,
		City:    hotel.City,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Hotel{}, fmt.Errorf("%w: %s", domain.ErrHotelAlreadyExists, pgErr.Detail)
		}
		return domain.Hotel{}, err
	}
	return toDomainHotel(row), nil
}

// Model mappers
func toDomainHotel(row sqlc.Hotel) domain.Hotel {
	return domain.Hotel{
		ID:        row.ID,
		Name:      row.Name,
		Address:   row.Address,
		City:      row.City,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
