-- Create new records table with user_id
CREATE TABLE new_records (
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

-- Copy existing data (if any) with a default user_id
-- You might want to adjust this default value based on your needs
INSERT INTO new_records (id, artist, album, year, genre, condition, user_id, created_at, updated_at)
SELECT id, artist, album, year, genre, condition, 1, created_at, updated_at
FROM records;

-- Drop old table and rename new one
DROP TABLE records;
ALTER TABLE new_records RENAME TO records;

-- Create index for record ownership
CREATE INDEX idx_records_user_id ON records(user_id);