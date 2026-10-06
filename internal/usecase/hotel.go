package usecase

import (
	"context"

	"github.com/netologist/hotel-booking/internal/domain"
)

// HotelService interface
type HotelService interface {
	FindHotels(ctx context.Context) ([]*domain.Hotel, error)
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

func (hs *hotelService) FindHotels(ctx context.Context) ([]*domain.Hotel, error) {
	return hs.hotelRepo.List(ctx)
}
