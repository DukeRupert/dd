-- Schema definitions
CREATE TABLE IF NOT EXISTS records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist TEXT NOT NULL,
    album TEXT NOT NULL,
    year INTEGER NOT NULL,
    genre TEXT NOT NULL,
    condition TEXT NOT NULL,
    image_filename TEXT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE password_resets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME,
    UNIQUE(token)
);

-- Create email_verifications table
CREATE TABLE email_verifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME,
    UNIQUE(token)
);

-- Record queries
-- name: GetRecord :one
SELECT * FROM records 
WHERE id = ? AND user_id = ? 
LIMIT 1;

-- name: ListRecords :many
SELECT * FROM records 
WHERE user_id = ? 
ORDER BY artist;

-- name: GetUserRecords :many
SELECT * FROM records
WHERE user_id = ?
    AND (artist LIKE ? OR album LIKE ?) 
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetUserRecordsAsc :many
SELECT * FROM records
WHERE user_id = ?
    AND (artist LIKE ? OR album LIKE ?) 
ORDER BY created_at ASC
LIMIT ? OFFSET ?;

-- name: GetUserRecordsCount :one
SELECT COUNT(*) 
FROM records
WHERE user_id = ?
    AND (artist LIKE ? OR album LIKE ?);

-- name: CreateRecord :one
INSERT INTO records (
    artist, album, year, genre, condition, user_id, image_filename
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateRecord :one
UPDATE records
SET artist = ?,
    album = ?,
    year = ?,
    genre = ?,
    condition = ?,
    image_filename = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: UpdateRecordImage :one
UPDATE records
SET image_filename = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: DeleteRecord :exec
DELETE FROM records 
WHERE id = ? AND user_id = ?;

-- User queries
-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, first_name, last_name
) VALUES (
    ?, ?, ?, ?
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET email = ?,
    first_name = ?,
    last_name = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: DeleteUserByEmail :exec
DELETE FROM users WHERE email = ?;

-- name: CreatePasswordReset :one
INSERT INTO password_resets (
    user_id, token, expires_at
) VALUES (
    ?, ?, ?
)
RETURNING *;

-- name: GetPasswordResetByToken :one
SELECT pr.*, u.email 
FROM password_resets pr
JOIN users u ON pr.user_id = u.id
WHERE pr.token = ? 
  AND pr.expires_at > datetime('now')
  AND pr.used_at IS NULL
LIMIT 1;

-- name: MarkPasswordResetUsed :exec
UPDATE password_resets
SET used_at = CURRENT_TIMESTAMP
WHERE token = ?;

-- name: CreateEmailVerification :one
INSERT INTO email_verifications (
    user_id, token, expires_at
) VALUES (
    ?, ?, ?
)
RETURNING *;

-- name: GetEmailVerificationByToken :one
SELECT ev.*, u.email 
FROM email_verifications ev
JOIN users u ON ev.user_id = u.id
WHERE ev.token = ? 
  AND ev.expires_at > datetime('now')
  AND ev.used_at IS NULL
LIMIT 1;

-- name: MarkEmailVerificationUsed :exec
UPDATE email_verifications
SET used_at = CURRENT_TIMESTAMP
WHERE token = ?;

-- name: MarkEmailVerified :exec
UPDATE users
SET verified_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: IsEmailVerified :one
SELECT verified_at IS NOT NULL as is_verified
FROM users
WHERE id = ?;