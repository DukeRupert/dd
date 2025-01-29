-- Remove indices
DROP INDEX IF EXISTS idx_records_user_id;
DROP INDEX IF EXISTS idx_users_email;

-- Create a new records table without user_id
CREATE TABLE records_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist TEXT NOT NULL,
    album TEXT NOT NULL,
    year INTEGER NOT NULL,
    genre TEXT NOT NULL,
    condition TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy existing data without user_id
INSERT INTO records_new 
SELECT id, artist, album, year, genre, condition, created_at, updated_at 
FROM records;

-- Drop old table and rename new one
DROP TABLE records;
ALTER TABLE records_new RENAME TO records;

-- Finally drop users table
DROP TABLE IF EXISTS users;