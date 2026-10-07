package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netologist/hotel-booking/internal/domain"
)

type CreateHotelInput struct {
	Name    string
	Address string
	City    string
}

type CreateHotelOutput struct {
	HotelOutput
}

type HotelOutput struct {
	ID        uuid.UUID
	Name      string
	Address   string
	City      string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// HotelService interface
type HotelService interface {
	FindHotels(ctx context.Context) ([]HotelOutput, error)
	CreateHotel(ctx context.Context, input CreateHotelInput) (CreateHotelOutput, error)
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

func (u *hotelService) FindHotels(ctx context.Context) ([]HotelOutput, error) {
	hotels, err := u.hotelRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	output := []HotelOutput{}
	for _, h := range hotels {
		output = append(output, HotelOutput{
			ID:        h.ID,
			Name:      h.Name,
			Address:   h.Address,
			City:      h.City,
			CreatedAt: h.CreatedAt,
			UpdatedAt: h.UpdatedAt,
		})
	}
	return output, nil
}

func (u *hotelService) CreateHotel(ctx context.Context, input CreateHotelInput) (CreateHotelOutput, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return CreateHotelOutput{}, fmt.Errorf("id not created: %w", err)
	}

	entity, err := u.hotelRepo.Create(ctx, &domain.Hotel{
		ID:      id,
		Name:    input.Name,
		Address: input.Address,
		City:    input.City,
	})

	if err != nil {
		return CreateHotelOutput{}, fmt.Errorf("db error %w", err)
	}

	return toCreateHotelOutput(entity), err
}

func toCreateHotelOutput(entity *domain.Hotel) CreateHotelOutput {
	return CreateHotelOutput{
		ID:        entity.ID,
		Name:      entity.Name,
		Address:   entity.Address,
		City:      entity.City,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
