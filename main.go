package main

import (
	"fmt"
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
	"github.com/dukerupert/dd/ratelimit"

	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type application struct {
	queries *db.Queries
	logger  zerolog.Logger
	auth    *auth.Manager
	rateLimiter *ratelimit.RateLimiter
}

type createRecordRequest struct {
	Artist    string `json:"artist" validate:"required,min=1,max=100"`
	Album     string `json:"album" validate:"required,min=1,max=100"`
	Year      int64  `json:"year" validate:"required,min=1900,max=2100"`
	Genre     string `json:"genre" validate:"required,min=1,max=50"`
	Condition string `json:"condition" validate:"required,oneof=Mint Near-Mint Very-Good Good Fair Poor"`
}

type registerUserRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	FirstName string `json:"first_name" validate:"required,max=50"`
	LastName  string `json:"last_name" validate:"required,max=50"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type tokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
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

	// Initialize rate limiter (5 attempts per 15 minutes)
	rateLimiter := ratelimit.New(5, 15*time.Minute)

	// Initialize application
	app := &application{
		queries:     db.New(sqlite),
		logger:      logger,
		auth:        authManager,
		rateLimiter: rateLimiter,
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
	e.POST("/register", app.registerUser)
	e.POST("/login", app.loginUser)
	e.POST("/refresh", app.refreshToken)

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

func (app *application) registerUser(c echo.Context) error {
	var req registerUserRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Check if user already exists
	_, err := app.queries.GetUserByEmail(context.Background(), req.Email)
	if err == nil {
		return api.NewBadRequestError("email already registered")
	} else if err != sql.ErrNoRows {
		return api.NewDatabaseError(err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		app.logger.Error().Err(err).Msg("Failed to hash password")
		return api.NewInternalError(err)
	}

	// Create user
	params := db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	user, err := app.queries.CreateUser(context.Background(), params)
	if err != nil {
		return api.NewDatabaseError(err)
	}

	// Generate JWT token
	token, err := app.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return api.NewInternalError(err)
	}

	app.logger.Info().
		Int64("user_id", user.ID).
		Str("email", user.Email).
		Msg("User registered successfully")

	// Return user data and token
	return c.JSON(http.StatusCreated, echo.Map{
		"user": userResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
		},
		"token": token,
	})
}

func (app *application) loginUser(c echo.Context) error {
	// Get IP address for rate limiting
	ip := c.RealIP()

	// Check rate limit
	if !app.rateLimiter.Allow(ip) {
		remaining, duration := app.rateLimiter.GetRemainingAttempts(ip)
		app.logger.Warn().
			Str("ip", ip).
			Int("remaining_attempts", remaining).
			Dur("lockout_duration", duration).
			Msg("Rate limit exceeded for login attempts")
		
		return api.NewTooManyRequestsError(fmt.Sprintf("Too many login attempts. Try again in %v", duration.Round(time.Minute)))
	}

	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user by email
	user, err := app.queries.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			app.logger.Info().
				Str("ip", ip).
				Str("email", req.Email).
				Msg("Failed login attempt - user not found")
			return api.NewUnauthorizedError("invalid credentials")
		}
		return api.NewDatabaseError(err)
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		app.logger.Info().
			Str("ip", ip).
			Str("email", req.Email).
			Msg("Failed login attempt - invalid password")
		return api.NewUnauthorizedError("invalid credentials")
	}

	// Generate token
	token, err := app.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return api.NewInternalError(err)
	}

	// Generate refresh token
	refreshToken, err := app.auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return api.NewInternalError(err)
	}

	app.logger.Info().
		Str("ip", ip).
		Int64("user_id", user.ID).
		Str("email", user.Email).
		Msg("User logged in successfully")

	return c.JSON(http.StatusOK, echo.Map{
		"user": userResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
		},
		"token":         token,
		"refresh_token": refreshToken,
	})
}

func (app *application) refreshToken(c echo.Context) error {
	var req refreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate refresh token
	userID, err := app.auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		app.logger.Info().
			Str("error", err.Error()).
			Msg("Invalid refresh token")
		return api.NewUnauthorizedError("invalid refresh token")
	}

	// Get user details
	user, err := app.queries.GetUserByID(context.Background(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewUnauthorizedError("user not found")
		}
		return api.NewDatabaseError(err)
	}

	// Generate new tokens
	newToken, err := app.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return api.NewInternalError(err)
	}

	newRefreshToken, err := app.auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return api.NewInternalError(err)
	}

	app.logger.Info().
		Int64("user_id", user.ID).
		Msg("Tokens refreshed successfully")

	return c.JSON(http.StatusOK, tokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	})
}
