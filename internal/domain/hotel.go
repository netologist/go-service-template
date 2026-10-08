package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinNameLength    = 2
	MaxNameLength    = 255
	MinAddressLength = 2
	MaxAddressLength = 255
	MinCityLength    = 2
	MaxCityLength    = 100
)

// Hotel represents a named accommodation establishment.
type Hotel struct {
	ID        uuid.UUID
	Name      string
	Address   string
	City      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewHotel validates and creates a new Hotel domain entity enforcing invariants.
func NewHotel(name, address, city string) (Hotel, error) {
	name = strings.TrimSpace(name)
	address = strings.TrimSpace(address)
	city = strings.TrimSpace(city)

	nameLen := utf8.RuneCountInString(name)
	if nameLen < MinNameLength || nameLen > MaxNameLength {
		return Hotel{}, fmt.Errorf(
			"%w: name length must be between %d and %d characters",
			ErrInvalidHotel, MinNameLength, MaxNameLength,
		)
	}

	addressLen := utf8.RuneCountInString(address)
	if addressLen < MinAddressLength || addressLen > MaxAddressLength {
		return Hotel{}, fmt.Errorf(
			"%w: address length must be between %d and %d characters",
			ErrInvalidHotel, MinAddressLength, MaxAddressLength,
		)
	}

	cityLen := utf8.RuneCountInString(city)
	if cityLen < MinCityLength || cityLen > MaxCityLength {
		return Hotel{}, fmt.Errorf(
			"%w: city length must be between %d and %d characters",
			ErrInvalidHotel, MinCityLength, MaxCityLength,
		)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Hotel{}, fmt.Errorf("generate id: %w", err)
	}

	return Hotel{
		ID:      id,
		Name:    name,
		Address: address,
		City:    city,
	}, nil
}

// HotelRepository interface
type HotelRepository interface {
	List(ctx context.Context) ([]Hotel, error)
	Create(ctx context.Context, hotel Hotel) (Hotel, error)
}
