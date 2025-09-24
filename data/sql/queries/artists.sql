-- name: CreateArtist :one
INSERT INTO artists (name)
VALUES (?)
RETURNING id, name, created_at, updated_at;

-- name: GetArtist :one
SELECT id, name, created_at, updated_at
FROM artists
WHERE id = ?;

-- name: GetArtistByName :one
SELECT id, name, created_at, updated_at
FROM artists
WHERE name = ?;

-- name: ListArtists :many
SELECT id, name, created_at, updated_at
FROM artists
ORDER BY name ASC;

-- name: ListArtistsWithPagination :many
SELECT id, name, created_at, updated_at
FROM artists
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: SearchArtistsByName :many
SELECT id, name, created_at, updated_at
FROM artists
WHERE name LIKE '%' || ? || '%'
ORDER BY name ASC;

-- name: UpdateArtist :one
UPDATE artists
SET name = ?
WHERE id = ?
RETURNING id, name, created_at, updated_at;

-- name: DeleteArtist :exec
DELETE FROM artists
WHERE id = ?;

-- name: CountArtists :one
SELECT COUNT(*) FROM artists;