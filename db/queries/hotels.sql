-- name: ListHotels :many
SELECT id, name, address, city, created_at, updated_at
FROM hotels
ORDER BY name ASC;
