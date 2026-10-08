package usecase

import (
	"context"
	"fmt"

	"github.com/netologist/hotel-booking/internal/domain"
)

// HotelService is the inbound port for hotel business operations.
type HotelService interface {
	FindHotels(ctx context.Context) ([]domain.Hotel, error)
	CreateHotel(ctx context.Context, name, address, city string) (domain.Hotel, error)
}

// hotelService implements HotelService.
type hotelService struct {
	hotelRepo domain.HotelRepository
}

var _ HotelService = (*hotelService)(nil)

// NewHotelService constructor returns a concrete *hotelService.
func NewHotelService(hotelRepo domain.HotelRepository) *hotelService {
	return &hotelService{
		hotelRepo: hotelRepo,
	}
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
		return domain.Hotel{}, fmt.Errorf("create hotel: %w", err)
	}

	return entity, nil
}
