package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/dukerupert/dd/db"

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

// Validator functions
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Return validation errors as a string
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, formatValidationError(err))
		}
		return echo.NewHTTPError(http.StatusBadRequest, strings.Join(errorMessages, "; "))
	}
	return nil
}

func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", field, err.Tag())
	}
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

	// Initialize application
	app := &application{
		queries: db.New(sqlite),
		logger:  logger,
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Initialize validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Custom logger middleware using zerolog
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			app.logger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Str("method", c.Request().Method).
				Str("ip", c.RealIP()).
				Str("user_agent", c.Request().UserAgent()).
				Dur("latency", v.Latency).
				Msg("Request")
			return nil
		},
	}))

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Vinyl Collection API")
	})

	// Record endpoints
	e.GET("/records", app.getAllRecords)
	e.GET("/records/:id", app.getRecord)
	e.POST("/records", app.createRecord)
	e.PUT("/records/:id", app.updateRecord)
	e.DELETE("/records/:id", app.deleteRecord)

	// Start server
	serverAddr := ":8080"
	logger.Info().Str("address", serverAddr).Msg("Starting server")
	if err := e.Start(serverAddr); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}

func (app *application) getAllRecords(c echo.Context) error {
	records, err := app.queries.ListRecords(context.Background())
	if err != nil {
		app.logger.Error().Err(err).Msg("Failed to list records")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	app.logger.Debug().Int("count", len(records)).Msg("Records retrieved")
	return c.JSON(http.StatusOK, records)
}

func (app *application) getRecord(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		app.logger.Error().Err(err).Str("id", c.Param("id")).Msg("Invalid record ID")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	record, err := app.queries.GetRecord(context.Background(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			app.logger.Debug().Int64("id", id).Msg("Record not found")
			return echo.NewHTTPError(http.StatusNotFound, "record not found")
		}
		app.logger.Error().Err(err).Int64("id", id).Msg("Failed to get record")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	app.logger.Debug().Int64("id", id).Msg("Record retrieved")
	return c.JSON(http.StatusOK, record)
}

func (app *application) createRecord(c echo.Context) error {
	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		app.logger.Error().Err(err).Msg("Failed to bind request")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate the request
	if err := c.Validate(&req); err != nil {
		app.logger.Error().Err(err).Interface("request", req).Msg("Failed validation")
		return err // Our custom validator already returns an echo.HTTPError
	}

	params := db.CreateRecordParams{
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.CreateRecord(context.Background(), params)
	if err != nil {
		app.logger.Error().Err(err).
			Str("artist", req.Artist).
			Str("album", req.Album).
			Msg("Failed to create record")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	app.logger.Info().
		Int64("id", record.ID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record created")
	return c.JSON(http.StatusCreated, record)
}

func (app *application) updateRecord(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		app.logger.Error().Err(err).Str("id", c.Param("id")).Msg("Invalid record ID")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		app.logger.Error().Err(err).Msg("Failed to bind request")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate the request
	if err := c.Validate(&req); err != nil {
		app.logger.Error().Err(err).Interface("request", req).Msg("Failed validation")
		return err // Our custom validator already returns an echo.HTTPError
	}

	params := db.UpdateRecordParams{
		ID:        id,
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.UpdateRecord(context.Background(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			app.logger.Debug().Int64("id", id).Msg("Record not found")
			return echo.NewHTTPError(http.StatusNotFound, "record not found")
		}
		app.logger.Error().Err(err).Int64("id", id).Msg("Failed to update record")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	app.logger.Info().
		Int64("id", record.ID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record updated")
	return c.JSON(http.StatusOK, record)
}

func (app *application) deleteRecord(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		app.logger.Error().Err(err).Str("id", c.Param("id")).Msg("Invalid record ID")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	err = app.queries.DeleteRecord(context.Background(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			app.logger.Debug().Int64("id", id).Msg("Record not found")
			return echo.NewHTTPError(http.StatusNotFound, "record not found")
		}
		app.logger.Error().Err(err).Int64("id", id).Msg("Failed to delete record")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	app.logger.Info().Int64("id", id).Msg("Record deleted")
	return c.NoContent(http.StatusNoContent)
}
