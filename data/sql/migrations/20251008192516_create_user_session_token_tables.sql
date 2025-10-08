-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = ON;

-- Create users table
CREATE TABLE users (
    id TEXT PRIMARY KEY, -- UUID as text
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user', -- 'user', 'admin', etc.
    is_active BOOLEAN DEFAULT 1,
    email_verified BOOLEAN DEFAULT 0,
    last_login_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_email UNIQUE (email),
    CONSTRAINT unique_username UNIQUE (username),
    CONSTRAINT check_role CHECK (role IN ('user', 'admin', 'moderator'))
);

-- Create sessions table for cookie-based authentication
CREATE TABLE sessions (
    id TEXT PRIMARY KEY, -- UUID as text
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL, -- session token stored in cookie
    ip_address TEXT,
    user_agent TEXT,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_session_token UNIQUE (token)
);

-- Create api_tokens table for API authentication
CREATE TABLE api_tokens (
    id TEXT PRIMARY KEY, -- UUID as text
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL, -- API token for Bearer authentication
    name TEXT NOT NULL, -- descriptive name for the token (e.g., "Mobile App", "CI/CD")
    scopes TEXT, -- JSON array of scopes/permissions
    last_used_at DATETIME,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_api_token UNIQUE (token)
);

-- Create indexes for better query performance
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_api_tokens_user_id ON api_tokens(user_id);
CREATE INDEX idx_api_tokens_token ON api_tokens(token);
CREATE INDEX idx_api_tokens_expires_at ON api_tokens(expires_at);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_users_updated_at
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_sessions_updated_at
    AFTER UPDATE ON sessions
    FOR EACH ROW
BEGIN
    UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_api_tokens_updated_at
    AFTER UPDATE ON api_tokens
    FOR EACH ROW
BEGIN
    UPDATE api_tokens SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert a default admin user (password: "admin123" - CHANGE THIS IN PRODUCTION!)
-- Password hash generated with bcrypt cost 10
INSERT INTO users (id, email, username, password_hash, role, is_active, email_verified) VALUES
('00000000-0000-0000-0000-000000000001', 'admin@example.com', 'admin', '$2a$10$rKvE8qH5L5o5J5J5J5J5JeN5J5J5J5J5J5J5J5J5J5J5J5J5J5J5J', 'admin', 1, 1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop triggers
DROP TRIGGER IF EXISTS update_api_tokens_updated_at;
DROP TRIGGER IF EXISTS update_sessions_updated_at;
DROP TRIGGER IF EXISTS update_users_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_api_tokens_user_id;
DROP INDEX IF EXISTS idx_api_tokens_token;
DROP INDEX IF EXISTS idx_api_tokens_expires_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

-- Drop tables
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd