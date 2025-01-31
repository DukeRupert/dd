package config

import (
	"database/sql"
	"errors"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

func runMigrations(db *sql.DB) error {
	logger := log.With().Str("component", "migrations").Logger()

	// Ensure migrations directory exists
	if err := os.MkdirAll("migrations", 0755); err != nil {
		logger.Error().Err(err).Msg("Failed to create migrations directory")
		return err
	}

	// Initialize sqlite driver for migrations
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize SQLite driver")
		return err
	}

	// Initialize migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create migration instance")
		return err
	}

	// Run migrations
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			logger.Error().Err(err).Msg("Failed to run migrations")
			return err
		}
		logger.Info().Msg("No migrations to run")
		return nil
	}

	logger.Info().Msg("Migrations completed successfully")
	return nil
}

func InitializeDatabase(dbPath string) (*sql.DB, error) {
	logger := log.With().Str("component", "database").Logger()
	logger.Info().Msg("Initializing database")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to open database")
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		logger.Error().Err(err).Msg("Failed to run migrations")
		return nil, err
	}

	logger.Info().Msg("Database initialization completed")
	return db, nil
}
