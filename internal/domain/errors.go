package domain

import "errors"

// ErrInvalidItem indicates a domain invariant violation on an item entity.
var ErrInvalidItem = errors.New("invalid item")

// ErrItemAlreadyExists indicates an item with the same unique attribute already exists.
var ErrItemAlreadyExists = errors.New("item already exists")

// ErrItemNotFound indicates the requested item does not exist.
var ErrItemNotFound = errors.New("item not found")
