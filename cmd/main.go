package main

import (
	"os"
	"time"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/internal/handlers"
	pocketbase "github.com/dukerupert/dd/pb"

	"github.com/labstack/echo/v4/middleware"
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
	_, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create Echo instance
	e := echo.New()
	e.Static("/static", "assets")
	e.Static("/uploads", "uploads")

	// Initialize logger
	logger := log.With().Str("component", "main").Logger()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Initialize the Pocketbase client
	pb := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(pb)
	
	// Public routes
	e.POST("/auth/login", authHandler.Login)
	e.GET("/login", handlers.LoginHandler)
	
	// Protected routes
	protected := e.Group("/api")
	protected.Use(authHandler.AuthMiddleware)

	// Start server
	serverAddr := ":8080"
	logger.Info().Str("address", serverAddr).Msg("Starting server")
	if err := e.Start(serverAddr); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
