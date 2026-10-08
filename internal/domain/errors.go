package domain

import "errors"

// ErrInvalidHotel indicates a domain invariant violation on a hotel entity.
var ErrInvalidHotel = errors.New("invalid hotel")

// ErrHotelAlreadyExists indicates a hotel with the same unique attribute already exists.
var ErrHotelAlreadyExists = errors.New("hotel already exists")
