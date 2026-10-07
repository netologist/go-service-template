-- +goose Up
CREATE TABLE IF NOT EXISTS hotels (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    address VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hotels_name ON hotels (name);
-- +goose Down
DROP TABLE IF EXISTS hotels;
