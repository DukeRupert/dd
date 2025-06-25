package main

import (
	"net/http"
	"os"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/internal/database"
	"github.com/dukerupert/dd/internal/logger"

	"github.com/rs/zerolog/log"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/echo/v4"
)

func main() {
	// Setup logger based on environment
	isDevelopment := os.Getenv("ENV") != "production"
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		if isDevelopment {
			logLevel = "debug"
		} else {
			logLevel = "info"
		}
	}
	
	logger.Setup(logger.Config{
		Development: isDevelopment,
		Level:       logLevel,
	})

	// Load configuration
	dbConfig, _, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create database connection
	db, err := database.New(dbConfig)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to connect to database")
    }
    defer db.Close()

	e := echo.New()
	e.HideBanner = true

	// Set custom error handler
	e.HTTPErrorHandler = logger.ErrorHandler()

		// Middleware
	e.Use(middleware.RequestID()) // Add request ID for tracing
	e.Use(middleware.Recover())
	e.Use(logger.Middleware())
	
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/artists", func(c echo.Context) error  {
		ctx := c.Request().Context()

		artists, err := db.Queries.ListArtists(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to retrieve artists from database")
		}

		return c.JSON(http.StatusOK, artists)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}

	log.Info().
		Str("port", port).
		Bool("development", isDevelopment).
		Msg("Starting server")

	if err := e.Start(":" + port); err != nil {
		log.Fatal().Err(err).Msg("Server startup failed")
	}
}
