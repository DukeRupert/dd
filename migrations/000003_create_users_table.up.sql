-- Add users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- First, add the column as nullable to allow existing records
ALTER TABLE records ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;

-- If you have existing records, you would need to set a default user_id here
-- UPDATE records SET user_id = 1 WHERE user_id IS NULL;

-- Then add the NOT NULL constraint
CREATE TABLE records_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist TEXT NOT NULL,
    album TEXT NOT NULL,
    year INTEGER NOT NULL,
    genre TEXT NOT NULL,
    condition TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy existing data
INSERT INTO records_new SELECT * FROM records;

-- Drop old table and rename new one
DROP TABLE records;
ALTER TABLE records_new RENAME TO records;

-- Create index for email lookups
CREATE INDEX idx_users_email ON users(email);

-- Create index for record ownership
CREATE INDEX idx_records_user_id ON records(user_id);