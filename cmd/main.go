package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/internal/database"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contextKey string // avoids collisions

const (
	RequestIDKey contextKey = "requestID"
	UserIDKey    contextKey = "userID"
	StartTimeKey contextKey = "startTime"
	LoggerKey    contextKey = "logger"
)

// Middleware function type
type Middleware func(http.Handler) http.Handler

// RequestID middleware adds a unique request ID to context
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate random ID
		bytes := make([]byte, 8)
		rand.Read(bytes)
		requestID := hex.EncodeToString(bytes)

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add to response headers for debugging
		w.Header().Set("X-Request-ID", requestID)

		// return handler
		next.ServeHTTP(w, r)
	})
}

// Logging middleware adds a logger with request context
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := GetRequestID(r.Context())
		userID := GetUserID(r.Context())

		logger := log.With().
			Str("requestID", requestID).
			Str("userID", userID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Logger()

		ctx := context.WithValue(r.Context(), LoggerKey, logger)
		ctx = context.WithValue(ctx, StartTimeKey, start)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		// Log duration with the contextualized logger
		elapsed := time.Since(start)
		logger.Info().Str("duration", elapsed.String()).Msg("Request")
	})
}

// Authentication middleware - extracts user from header/token
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple example - in reality you'd validate JWT/session
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Chain middleware helper
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		// Apply middleware in reverse order
		for _, middleware := range slices.Backward(middlewares) {
			final = middleware(final)
		}
		return final
	}
}

// Context value extractors - type-safe helpers
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return "anonymous"
}

func GetStartTime(ctx context.Context) time.Time {
	if t, ok := ctx.Value(StartTimeKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}

func GetLogger(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(zerolog.Logger); ok {
		return logger
	}
	fmt.Println("No logger found in context. Return default logger.")
	return log.Logger
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	requestID := GetRequestID(r.Context())
	userID := GetUserID(r.Context())
	startTime := GetStartTime(r.Context())
	fmt.Fprintf(w, "Hello! Request ID: %s, User: %s, Started: %v\n",
		requestID, userID, startTime.Format(time.RFC3339))
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id") // Go 1.22+ path parameter
	requestID := GetRequestID(r.Context())

	fmt.Fprintf(w, "User profile for %s (Request: %s)\n", userID, requestID)
}

// Protected route example
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())

	if userID == "anonymous" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "Protected content for user: %s\n", userID)
}

func (app *App) getArtistsHandler(w http.ResponseWriter, r *http.Request) {
	logger := GetLogger(r.Context())

	// do the records exist?
	artists, err := app.DB.Queries.ListArtists(r.Context())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve artists from database")
		http.Error(w, "Artists not found", http.StatusInternalServerError)
		return
	}

	// great success!
	logger.Info().Int("artists", len(artists)).Msg("List artists")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artists)
}

func (app *App) getArtistHandler(w http.ResponseWriter, r *http.Request) {
	logger := GetLogger(r.Context())

	// is there an id?
	id := r.PathValue("id")
	if id == "" {
		logger.Info().Msg("Bad request. Missing path id")
		http.Error(w, "Missing parameter: id", http.StatusBadRequest)
		return
	}

	// is it an integer?
	artistID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid parameter id")
		http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
		return
	}

	// does the record exist?
	artist, err := app.DB.Queries.GetArtist(r.Context(), int32(artistID))
	if err != nil {
		logger.Error().Err(err).Int64("artistID", artistID).Msg("Failed to retrieve artist record")
		http.Error(w, "Missing record", http.StatusNotFound)
		return
	}

	// great success!
	logger.Info().Int64("artistID", artistID).Str("name", artist.Name).Msg("Artist record retrieved")
	json.NewEncoder(w).Encode(artist)
}

type App struct {
	DB     *database.Database
	Logger zerolog.Logger
}

func setupLogger(environment string, logLevel string) zerolog.Logger {
	// Setup logger based on environment
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if environment  == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return log.Logger
}

func main() {
	// Load configuration
	dbConfig, appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
	
	logger := setupLogger(appConfig.Environment, appConfig.LogLevel)

	// Create database connection
	db, err := database.New(dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = ":1323"
	} else {
		port = ":" + port
	}

	log.Info().
		Str("port", port).
		Msg("Starting server")

	app := &App{DB: db, Logger: logger}
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /users/{id}", userHandler)
	mux.HandleFunc("GET /protected", protectedHandler)
	mux.HandleFunc("GET /artists", app.getArtistsHandler)
	mux.HandleFunc("GET /artists/{id}", app.getArtistHandler)

	// Chain all middleware
	handler := ChainMiddleware(
		RequestIDMiddleware,
		LoggingMiddleware,
	)(mux)

	fmt.Println("Server starting on :8080")
	log.Fatal().Err(http.ListenAndServe(":8080", handler))
}

// func setupRoutes(e *echo.Echo, h *handlers.Handlers) {
// 	// API v1 routes
// 	api := e.Group("/api/v1")

// 	// Artist routes
// 	artists := api.Group("/artists")
// 	artists.POST("", h.CreateArtist)
// 	artists.GET("", h.ListArtists)
// 	artists.GET("/:id", h.GetArtist)
// 	artists.PUT("/:id", h.UpdateArtist)
// 	artists.DELETE("/:id", h.DeleteArtist)

// 	// Health check
// 	e.GET("/health", func(c echo.Context) error {
// 		return c.JSON(200, map[string]string{
// 			"status": "healthy",
// 		})
// 	})
// }
