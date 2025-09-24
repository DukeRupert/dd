-- +goose Up
-- +goose StatementBegin
-- Enable foreign key constraints (important for SQLite)
PRAGMA foreign_keys = ON;

-- Create artists table
CREATE TABLE artists (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_artist_name UNIQUE (name)
);

-- Create locations table
CREATE TABLE locations (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create records table
CREATE TABLE records (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    artist_id INTEGER REFERENCES artists(id) ON DELETE SET NULL,
    album_title TEXT,
    release_year INTEGER,
    current_location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    home_location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    catalog_number TEXT,
    condition TEXT,
    notes TEXT,
    last_played_at DATETIME,
    play_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_records_artist_id ON records(artist_id);
CREATE INDEX idx_records_current_location_id ON records(current_location_id);
CREATE INDEX idx_records_home_location_id ON records(home_location_id);
CREATE INDEX idx_records_title ON records(title);
CREATE INDEX idx_artists_name ON artists(name);

-- Create triggers to automatically update updated_at (SQLite syntax)
CREATE TRIGGER update_artists_updated_at
    AFTER UPDATE ON artists
    FOR EACH ROW
BEGIN
    UPDATE artists SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_locations_updated_at
    AFTER UPDATE ON locations
    FOR EACH ROW
BEGIN
    UPDATE locations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_records_updated_at
    AFTER UPDATE ON records
    FOR EACH ROW
BEGIN
    UPDATE records SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert some default locations (using 1/0 for boolean values)
INSERT INTO locations (name, description, is_default) VALUES
('Main Collection', 'Primary storage location for vinyl records', 1),
('Currently Playing', 'Records that are currently out being played', 0),
('Cleaning Station', 'Records waiting to be cleaned', 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop triggers
DROP TRIGGER IF EXISTS update_records_updated_at;
DROP TRIGGER IF EXISTS update_locations_updated_at;
DROP TRIGGER IF EXISTS update_artists_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_records_artist_id;
DROP INDEX IF EXISTS idx_records_current_location_id;
DROP INDEX IF EXISTS idx_records_home_location_id;
DROP INDEX IF EXISTS idx_records_title;
DROP INDEX IF EXISTS idx_artists_name;

-- Drop tables (SQLite handles foreign keys automatically when dropping)
DROP TABLE IF EXISTS records;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS artists;
-- +goose StatementEnd