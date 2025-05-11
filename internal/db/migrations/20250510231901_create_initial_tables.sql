-- +goose Up
-- +goose StatementBegin
-- Create artists table
CREATE TABLE artists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create locations table
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create records table
CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist_id INTEGER REFERENCES artists(id) ON DELETE SET NULL,
    album_title VARCHAR(255),
    release_year INTEGER,
    current_location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    home_location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    catalog_number VARCHAR(100),
    condition VARCHAR(50),
    notes TEXT,
    last_played_at TIMESTAMP WITH TIME ZONE,
    play_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_records_artist_id ON records(artist_id);
CREATE INDEX idx_records_current_location_id ON records(current_location_id);
CREATE INDEX idx_records_home_location_id ON records(home_location_id);
CREATE INDEX idx_records_title ON records(title);
CREATE INDEX idx_artists_name ON artists(name);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_artists_updated_at
    BEFORE UPDATE ON artists
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_locations_updated_at
    BEFORE UPDATE ON locations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_records_updated_at
    BEFORE UPDATE ON records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert some default locations
INSERT INTO locations (name, description, is_default) VALUES
('Main Collection', 'Primary storage location for vinyl records', TRUE),
('Currently Playing', 'Records that are currently out being played', FALSE),
('Cleaning Station', 'Records waiting to be cleaned', FALSE);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop triggers
DROP TRIGGER IF EXISTS update_records_updated_at ON records;
DROP TRIGGER IF EXISTS update_locations_updated_at ON locations;
DROP TRIGGER IF EXISTS update_artists_updated_at ON artists;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_records_artist_id;
DROP INDEX IF EXISTS idx_records_current_location_id;
DROP INDEX IF EXISTS idx_records_home_location_id;
DROP INDEX IF EXISTS idx_records_title;
DROP INDEX IF EXISTS idx_artists_name;

-- Drop tables (cascade to handle foreign key constraints)
DROP TABLE IF EXISTS records CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS artists CASCADE;
-- +goose StatementEnd
