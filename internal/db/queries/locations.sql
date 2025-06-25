-- name: CreateLocation :one
INSERT INTO locations (name, description, is_default)
VALUES ($1, $2, $3)
RETURNING id, name, description, is_default, created_at, updated_at;

-- name: GetLocation :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE id = $1;

-- name: GetLocationByName :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE name = $1;

-- name: GetDefaultLocation :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE is_default = true
LIMIT 1;

-- name: ListLocations :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
ORDER BY name ASC;

-- name: ListLocationsWithPagination :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: SearchLocationsByName :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name ASC;

-- name: UpdateLocation :one
UPDATE locations
SET name = $2, description = $3, is_default = $4
WHERE id = $1
RETURNING id, name, description, is_default, created_at, updated_at;

-- name: UpdateLocationName :one
UPDATE locations
SET name = $2
WHERE id = $1
RETURNING id, name, description, is_default, created_at, updated_at;

-- name: SetDefaultLocation :exec
UPDATE locations
SET is_default = CASE
    WHEN id = $1 THEN true
    ELSE false
END;

-- name: DeleteLocation :exec
DELETE FROM locations
WHERE id = $1;

-- name: CountLocations :one
SELECT COUNT(*) FROM locations;