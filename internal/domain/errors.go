package domain

import "errors"

// ErrInvalidHotel indicates a domain invariant violation on a hotel entity.
var ErrInvalidHotel = errors.New("invalid hotel")
