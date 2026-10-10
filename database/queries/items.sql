-- name: ListItems :many
SELECT id, name, description, price, created_at, updated_at
FROM items
ORDER BY name ASC;

-- name: CreateItem :one
INSERT INTO items (id, name, description, price)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, price, created_at, updated_at;
