package main

import (
	"database/sql"
	"errors"

	"os"
	"time"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/auth"
	"github.com/dukerupert/dd/db"
	"github.com/dukerupert/dd/handler"
	"github.com/dukerupert/dd/ratelimit"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})
}

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

func initializeDatabase() (*sql.DB, error) {
	logger := log.With().Str("component", "database").Logger()
	logger.Info().Msg("Initializing database")

	db, err := sql.Open("sqlite3", "vinyl.db")
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

func main() {
	// Initialize logger
	logger := log.With().Str("component", "main").Logger()

	// Initialize database with migrations
	sqlite, err := initializeDatabase()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer sqlite.Close()

	// Initialize auth manager
	authConfig := auth.DefaultConfig()
	authManager := auth.NewManager(authConfig)

	// Initialize rate limiter (5 attempts per 15 minutes)
	rateLimiter := ratelimit.New(5, 15*time.Minute)

	// Initialize application
	appConfig := handler.Config{
		Queries: 	db.New(sqlite),
		Logger:		logger,
		Auth:		authManager,
		RateLimiter: rateLimiter,
	}

	app := handler.NewHandler(appConfig)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Initialize validator
	e.Validator = api.NewValidator()

	app.ApplyMiddleware(e)
	app.CreateRoutes(e)

	// Start server
	serverAddr := ":8080"
	logger.Info().Str("address", serverAddr).Msg("Starting server")
	if err := e.Start(serverAddr); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
