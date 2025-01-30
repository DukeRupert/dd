-- Add verified_at to users table
ALTER TABLE users ADD COLUMN verified_at DATETIME;

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

-- Create index for token lookups
CREATE INDEX idx_email_verifications_token ON email_verifications(token);