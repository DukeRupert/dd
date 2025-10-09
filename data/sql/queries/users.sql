-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (id, email, username, password_hash, role, is_active, email_verified)
VALUES (?, ?, ?, ?, ?, 1, 0)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP LIMIT 1;

-- name: GetAPITokenByToken :one
SELECT * FROM api_tokens WHERE token = ? AND expires_at > CURRENT_TIMESTAMP LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, token, ip_address, user_agent, expires_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = ?;

-- name: CreateAPIToken :one
INSERT INTO api_tokens (id, user_id, token, name, scopes, expires_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;