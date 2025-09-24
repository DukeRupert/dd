-- name: CreateLocation :one
INSERT INTO locations (name, description, is_default)
VALUES (?, ?, ?)
RETURNING id, name, description, is_default, created_at, updated_at;

-- name: GetLocation :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE id = ?;

-- name: GetLocationByName :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE name = ?;

-- name: GetDefaultLocation :one
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE is_default = 1
LIMIT 1;

-- name: ListLocations :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
ORDER BY name ASC;

-- name: ListLocationsWithPagination :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: SearchLocationsByName :many
SELECT id, name, description, is_default, created_at, updated_at
FROM locations
WHERE name LIKE '%' || ? || '%'
ORDER BY name ASC;

-- name: UpdateLocation :one
UPDATE locations
SET name = ?, description = ?, is_default = ?
WHERE id = ?
RETURNING id, name, description, is_default, created_at, updated_at;

-- name: UpdateLocationName :one
UPDATE locations
SET name = ?
WHERE id = ?
RETURNING id, name, description, is_default, created_at, updated_at;

-- name: SetDefaultLocation :exec
UPDATE locations
SET is_default = CASE
    WHEN id = ? THEN 1
    ELSE 0
END;

-- name: DeleteLocation :exec
DELETE FROM locations
WHERE id = ?;

-- name: CountLocations :one
SELECT COUNT(*) FROM locations;