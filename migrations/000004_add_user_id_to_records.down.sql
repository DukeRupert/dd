DROP INDEX IF EXISTS idx_records_user_id;
CREATE TABLE new_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist TEXT NOT NULL,
    album TEXT NOT NULL,
    year INTEGER NOT NULL,
    genre TEXT NOT NULL,
    condition TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO new_records SELECT id, artist, album, year, genre, condition, created_at, updated_at FROM records;
DROP TABLE records;
ALTER TABLE new_records RENAME TO records;