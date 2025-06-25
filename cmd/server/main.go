package main

import (
	"os"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/internal/database"
	"github.com/dukerupert/dd/internal/handlers"
	"github.com/dukerupert/dd/internal/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
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

		// Create handlers with dependencies
	h := handlers.NewHandlers(db)

	// Setup Echo
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = logger.ErrorHandler()

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(logger.Middleware())
	e.Use(middleware.CORS())

	// Routes
	setupRoutes(e, h)

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


func setupRoutes(e *echo.Echo, h *handlers.Handlers) {
	// API v1 routes
	api := e.Group("/api/v1")

	// Artist routes
	artists := api.Group("/artists")
	artists.POST("", h.CreateArtist)
	artists.GET("", h.ListArtists)
	artists.GET("/:id", h.GetArtist)
	artists.PUT("/:id", h.UpdateArtist)
	artists.DELETE("/:id", h.DeleteArtist)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})
}
