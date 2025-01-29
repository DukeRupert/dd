package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/auth"
	"github.com/dukerupert/dd/db"

	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type application struct {
	queries *db.Queries
	logger  zerolog.Logger
	auth    *auth.Manager
}

type createRecordRequest struct {
	Artist    string `json:"artist" validate:"required,min=1,max=100"`
	Album     string `json:"album" validate:"required,min=1,max=100"`
	Year      int64  `json:"year" validate:"required,min=1900,max=2100"`
	Genre     string `json:"genre" validate:"required,min=1,max=50"`
	Condition string `json:"condition" validate:"required,oneof=Mint Near-Mint Very-Good Good Fair Poor"`
}

// Custom validator
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return api.NewValidationError(err.Error())
	}
	return nil
}

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

	// Initialize application
	app := &application{
		queries: db.New(sqlite),
		logger:  logger,
		auth:    authManager,
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	
	// Initialize validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware - order is important
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(api.ErrorHandlerMiddleware(logger))
	e.Use(middleware.CORS())

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Vinyl Collection API")
	})
	// TODO: Add login/register endpoints here

	// Protected routes
	protected := e.Group("")
	protected.Use(authManager.Middleware())

	// Record endpoints
	protected.GET("/records", app.getAllRecords)
	protected.GET("/records/:id", app.getRecord)
	protected.POST("/records", app.createRecord)
	protected.PUT("/records/:id", app.updateRecord)
	protected.DELETE("/records/:id", app.deleteRecord)

	// Start server
	serverAddr := ":8080"
	logger.Info().Str("address", serverAddr).Msg("Starting server")
	if err := e.Start(serverAddr); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}

func (app *application) getAllRecords(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	records, err := app.queries.ListRecords(context.Background(), userID)
	if err != nil {
		return api.NewDatabaseError(err)
	}

	app.logger.Debug().
		Int64("user_id", userID).
		Int("count", len(records)).
		Msg("Records retrieved")
	
	return c.JSON(http.StatusOK, records)
}

func (app *application) getRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return api.NewBadRequestError("invalid id format")
	}

	record, err := app.queries.GetRecord(context.Background(), db.GetRecordParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewNotFoundError("record")
		}
		return api.NewDatabaseError(err)
	}

	return c.JSON(http.StatusOK, record)
}

func (app *application) createRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	// Validate the request
	if err := c.Validate(&req); err != nil {
		return err // Our custom validator already returns an api.ValidationError
	}

	params := db.CreateRecordParams{
		UserID:    userID,
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.CreateRecord(context.Background(), params)
	if err != nil {
		return api.NewDatabaseError(err)
	}

	app.logger.Info().
		Int64("id", record.ID).
		Int64("user_id", userID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record created")

	return c.JSON(http.StatusCreated, record)
}

func (app *application) updateRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return api.NewBadRequestError("invalid id format")
	}

	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	// Validate the request
	if err := c.Validate(&req); err != nil {
		return err // Our custom validator already returns an api.ValidationError
	}

	params := db.UpdateRecordParams{
		ID:        id,
		UserID:    userID,
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.UpdateRecord(context.Background(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewNotFoundError("record")
		}
		return api.NewDatabaseError(err)
	}

	app.logger.Info().
		Int64("id", record.ID).
		Int64("user_id", userID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record updated")

	return c.JSON(http.StatusOK, record)
}

func (app *application) deleteRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return api.NewBadRequestError("invalid id format")
	}

	err = app.queries.DeleteRecord(context.Background(), db.DeleteRecordParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewNotFoundError("record")
		}
		return api.NewDatabaseError(err)
	}

	app.logger.Info().
		Int64("id", id).
		Int64("user_id", userID).
		Msg("Record deleted")

	return c.NoContent(http.StatusNoContent)
}
