package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netologist/hotel-booking/internal/domain"
)

// HotelService interface
type HotelService interface {
	FindHotels(ctx context.Context) ([]domain.Hotel, error)
	CreateHotel(ctx context.Context, name, address, city string) (domain.Hotel, error)
}

// NewHotelService constructor
func NewHotelService(hotelRepo domain.HotelRepository) HotelService {
	return &hotelService{
		hotelRepo: hotelRepo,
	}
}

// hotelService implementation
type hotelService struct {
	hotelRepo domain.HotelRepository
}

func (u *hotelService) FindHotels(ctx context.Context) ([]domain.Hotel, error) {
	return u.hotelRepo.List(ctx)
}

func (u *hotelService) CreateHotel(ctx context.Context, name, address, city string) (domain.Hotel, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.Hotel{}, fmt.Errorf("id not created: %w", err)
	}

	entity, err := u.hotelRepo.Create(ctx, domain.Hotel{
		ID:      id,
		Name:    name,
		Address: address,
		City:    city,
	})
	if err != nil {
		return domain.Hotel{}, fmt.Errorf("db error: %w", err)
	}

	return entity, nil
}
