-- schema.sql
CREATE TABLE records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist TEXT NOT NULL,
    album TEXT NOT NULL,
    year INTEGER NOT NULL,
    genre TEXT NOT NULL,
    condition TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: GetRecord :one
SELECT * FROM records WHERE id = ? LIMIT 1;

-- name: ListRecords :many
SELECT * FROM records ORDER BY artist;

-- name: CreateRecord :one
INSERT INTO records (
    artist, album, year, genre, condition
) VALUES (
    ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateRecord :one
UPDATE records
SET artist = ?,
    album = ?,
    year = ?,
    genre = ?,
    condition = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteRecord :exec
DELETE FROM records WHERE id = ?;