-- name: CreateArtist :one
INSERT INTO artists (name)
VALUES ($1)
RETURNING id, name, created_at, updated_at;

-- name: GetArtist :one
SELECT id, name, created_at, updated_at
FROM artists
WHERE id = $1;

-- name: GetArtistByName :one
SELECT id, name, created_at, updated_at
FROM artists
WHERE name = $1;

-- name: ListArtists :many
SELECT id, name, created_at, updated_at
FROM artists
ORDER BY name ASC;

-- name: ListArtistsWithPagination :many
SELECT id, name, created_at, updated_at
FROM artists
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: SearchArtistsByName :many
SELECT id, name, created_at, updated_at
FROM artists
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name ASC;

-- name: UpdateArtist :one
UPDATE artists
SET name = $2
WHERE id = $1
RETURNING id, name, created_at, updated_at;

-- name: DeleteArtist :exec
DELETE FROM artists
WHERE id = $1;

-- name: CountArtists :one
SELECT COUNT(*) FROM artists;