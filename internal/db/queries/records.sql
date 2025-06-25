-- name: CreateRecord :one
INSERT INTO records (
    title, artist_id, album_title, release_year, 
    current_location_id, home_location_id, catalog_number, 
    condition, notes, play_count
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, title, artist_id, album_title, release_year, 
          current_location_id, home_location_id, catalog_number, 
          condition, notes, last_played_at, play_count, 
          created_at, updated_at;

-- name: GetRecord :one
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
WHERE id = $1;

-- name: GetRecordWithDetails :one
SELECT r.id, r.title, r.album_title, r.release_year, 
       r.catalog_number, r.condition, r.notes, 
       r.last_played_at, r.play_count, r.created_at, r.updated_at,
       a.id as artist_id, a.name as artist_name,
       cl.id as current_location_id, cl.name as current_location_name,
       hl.id as home_location_id, hl.name as home_location_name
FROM records r
LEFT JOIN artists a ON r.artist_id = a.id
LEFT JOIN locations cl ON r.current_location_id = cl.id
LEFT JOIN locations hl ON r.home_location_id = hl.id
WHERE r.id = $1;

-- name: ListRecords :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
ORDER BY title ASC;

-- name: ListRecordsWithDetails :many
SELECT r.id, r.title, r.album_title, r.release_year, 
       r.catalog_number, r.condition, r.notes, 
       r.last_played_at, r.play_count, r.created_at, r.updated_at,
       a.id as artist_id, a.name as artist_name,
       cl.id as current_location_id, cl.name as current_location_name,
       hl.id as home_location_id, hl.name as home_location_name
FROM records r
LEFT JOIN artists a ON r.artist_id = a.id
LEFT JOIN locations cl ON r.current_location_id = cl.id
LEFT JOIN locations hl ON r.home_location_id = hl.id
ORDER BY r.title ASC;

-- name: ListRecordsWithPagination :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
ORDER BY title ASC
LIMIT $1 OFFSET $2;

-- name: SearchRecordsByTitle :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
WHERE title ILIKE '%' || $1 || '%'
ORDER BY title ASC;

-- name: SearchRecordsByAlbum :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
WHERE album_title ILIKE '%' || $1 || '%'
ORDER BY album_title ASC;

-- name: GetRecordsByArtist :many
SELECT r.id, r.title, r.artist_id, r.album_title, r.release_year, 
       r.current_location_id, r.home_location_id, r.catalog_number, 
       r.condition, r.notes, r.last_played_at, r.play_count, 
       r.created_at, r.updated_at
FROM records r
WHERE r.artist_id = $1
ORDER BY r.title ASC;

-- name: GetRecordsByLocation :many
SELECT r.id, r.title, r.artist_id, r.album_title, r.release_year, 
       r.current_location_id, r.home_location_id, r.catalog_number, 
       r.condition, r.notes, r.last_played_at, r.play_count, 
       r.created_at, r.updated_at
FROM records r
WHERE r.current_location_id = $1
ORDER BY r.title ASC;

-- name: GetRecordsByReleaseYear :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
WHERE release_year = $1
ORDER BY title ASC;

-- name: GetRecordsByCondition :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
WHERE condition = $1
ORDER BY title ASC;

-- name: GetRecentlyPlayedRecords :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
WHERE last_played_at IS NOT NULL
ORDER BY last_played_at DESC
LIMIT $1;

-- name: GetMostPlayedRecords :many
SELECT id, title, artist_id, album_title, release_year, 
       current_location_id, home_location_id, catalog_number, 
       condition, notes, last_played_at, play_count, 
       created_at, updated_at
FROM records
WHERE play_count > 0
ORDER BY play_count DESC
LIMIT $1;

-- name: UpdateRecord :one
UPDATE records
SET title = $2, artist_id = $3, album_title = $4, release_year = $5,
    current_location_id = $6, home_location_id = $7, catalog_number = $8,
    condition = $9, notes = $10
WHERE id = $1
RETURNING id, title, artist_id, album_title, release_year, 
          current_location_id, home_location_id, catalog_number, 
          condition, notes, last_played_at, play_count, 
          created_at, updated_at;

-- name: UpdateRecordLocation :one
UPDATE records
SET current_location_id = $2
WHERE id = $1
RETURNING id, title, artist_id, album_title, release_year, 
          current_location_id, home_location_id, catalog_number, 
          condition, notes, last_played_at, play_count, 
          created_at, updated_at;

-- name: UpdateRecordCondition :one
UPDATE records
SET condition = $2
WHERE id = $1
RETURNING id, title, artist_id, album_title, release_year, 
          current_location_id, home_location_id, catalog_number, 
          condition, notes, last_played_at, play_count, 
          created_at, updated_at;

-- name: RecordPlayback :one
UPDATE records
SET last_played_at = CURRENT_TIMESTAMP, play_count = play_count + 1
WHERE id = $1
RETURNING id, title, artist_id, album_title, release_year, 
          current_location_id, home_location_id, catalog_number, 
          condition, notes, last_played_at, play_count, 
          created_at, updated_at;

-- name: DeleteRecord :exec
DELETE FROM records
WHERE id = $1;

-- name: CountRecords :one
SELECT COUNT(*) FROM records;

-- name: CountRecordsByArtist :one
SELECT COUNT(*) FROM records WHERE artist_id = $1;

-- name: CountRecordsByLocation :one
SELECT COUNT(*) FROM records WHERE current_location_id = $1;