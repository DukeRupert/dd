package main

import (
	"os"
	"time"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/auth"
	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/db"
	"github.com/dukerupert/dd/handler"
	"github.com/dukerupert/dd/ratelimit"

	"github.com/labstack/echo/v4"
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

func main() {
	// Initialize logger
	logger := log.With().Str("component", "main").Logger()

	// Initialize database with migrations
	sqlite, err := config.InitializeDatabase()
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
		Queries:     db.New(sqlite),
		Logger:      logger,
		Auth:        authManager,
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
