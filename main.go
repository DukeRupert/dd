package main

import (
	"os"
	"time"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/db"
	"github.com/dukerupert/dd/email"
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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Initialize validator
	e.Validator = api.NewValidator()

	// Initialize logger
	logger := log.With().Str("component", "main").Logger()

	// Initialize database with migrations
	sqlite, err := config.InitializeDatabase(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer sqlite.Close()

	// Configure sessions
    if err := config.ConfigureSessions(e, cfg.DatabaseURL, cfg.SessionSecret); err != nil {
        logger.Fatal().Err(err).Msg("Failed to configure sessions")
    }

	// Initialize rate limiter (5 attempts per 15 minutes)
	rateLimiter := ratelimit.New(5, 15*time.Minute)

	// Initialize email client
	var emailClient *email.Client
	if cfg.PostmarkServerToken != "" && cfg.FromEmail != "" {
		var err error
		emailClient, err = email.NewClient(
			cfg.PostmarkServerToken,
			cfg.FromEmail,
		)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to initialize email client")
		} else {
			logger.Info().
				Str("from_email", cfg.FromEmail).
				Msg("Email client initialized successfully")
		}
	} else {
		logger.Warn().Msg("Email client not configured, password reset functionality will be disabled")
	}

	// Initialize application
	appConfig := handler.Config{
		Queries:     db.New(sqlite),
		Logger:      logger,
		RateLimiter: rateLimiter,
		Mailer:      emailClient,
	}

	app := handler.NewHandler(appConfig)
	app.ApplyMiddleware(e)
	app.CreateRoutes(e)

	// Start server
	serverAddr := ":8080"
	logger.Info().Str("address", serverAddr).Msg("Starting server")
	if err := e.Start(serverAddr); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
