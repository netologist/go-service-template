package usecase

import (
	"context"
	"fmt"

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
	hotel, err := domain.NewHotel(name, address, city)
	if err != nil {
		return domain.Hotel{}, fmt.Errorf("create hotel: %w", err)
	}

	entity, err := u.hotelRepo.Create(ctx, hotel)
	if err != nil {
		return domain.Hotel{}, fmt.Errorf("db error: %w", err)
	}

	return entity, nil
}
