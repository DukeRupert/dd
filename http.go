package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

// Validation error response
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details"`
}

// Convert validator errors to friendly messages
func getValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationError := ValidationError{
				Field: fieldError.Field(),
				Tag:   fieldError.Tag(),
				Value: fieldError.Param(),
			}

			// Custom error messages based on tag
			switch fieldError.Tag() {
			case "required":
				validationError.Message = fmt.Sprintf("%s is required", fieldError.Field())
			case "min":
				validationError.Message = fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param())
			case "max":
				validationError.Message = fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param())
			case "email":
				validationError.Message = fmt.Sprintf("%s must be a valid email", fieldError.Field())
			default:
				validationError.Message = fmt.Sprintf("%s failed validation for '%s'", fieldError.Field(), fieldError.Tag())
			}

			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

func writeValidationErrorJSON(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	validationErrors := getValidationErrors(err)

	response := ValidationErrorResponse{
		Error:   "Validation Failed",
		Message: "The request contains invalid data",
		Details: validationErrors,
	}

	json.NewEncoder(w).Encode(response)
}

func bind(r *http.Request, v interface{}) error {
	if r.Body == nil && r.Method != "GET" {
		return fmt.Errorf("request body is required")
	}

	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		return bindJSON(r, v)
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return bindForm(r, v)
	case strings.Contains(contentType, "multipart/form-data"):
		return bindForm(r, v)
	default:
		return fmt.Errorf("unsupported content-type: %s", contentType)
	}
}

func bindJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return validate.Struct(v)
}

func bindForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	// Use reflection to map form values to struct fields
	return mapFormToStruct(r, v)
}

func mapFormToStruct(r *http.Request, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get form tag value
		formTag := field.Tag.Get("form")
		if formTag == "" {
			continue
		}

		// Get value from form
		formValue := r.FormValue(formTag)
		if formValue == "" {
			continue
		}

		// Set value based on field type
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(formValue)
		case reflect.Int, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(formValue, 10, 64); err == nil {
				fieldValue.SetInt(intVal)
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, err := strconv.ParseFloat(formValue, 64); err == nil {
				fieldValue.SetFloat(floatVal)
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(formValue); err == nil {
				fieldValue.SetBool(boolVal)
			}
		}
	}

	return validate.Struct(v)
}

// Regular error response helper (for non-validation errors)
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func writeErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

type contextKey string // avoids collisions

const (
	RequestIDKey contextKey = "requestID"
	UserIDKey    contextKey = "userID"
	StartTimeKey contextKey = "startTime"
)

// RequestID middleware adds a unique request ID to context
func RequestIDMiddleware(h http.Handler) http.Handler {
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
		h.ServeHTTP(w, r)
	})
}

// Logging middleware adds a logger with request context
func LoggingMiddleware(h http.Handler, l zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := GetRequestID(r.Context())
		userID := GetUserID(r.Context())

		logger := l.With().
			Str("requestID", requestID).
			Str("userID", userID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Logger()

		h.ServeHTTP(w, r)

		// Log duration with the contextualized logger
		elapsed := time.Since(start)
		logger.Info().Str("duration", elapsed.String()).Msg("Request")
	})
}

// Authentication middleware - extracts user from header/token
func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple example - in reality you'd validate JWT/session
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
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

func homeHandler(t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		type DashboardStats struct {
			TotalArtists   int
			TotalAlbums    int
			TotalLocations int
			EstimatedValue string
		}

		type DashboardData struct {
			Title string
			Stats DashboardStats
		}

		data := DashboardData{
			Title: "Home",
			Stats: DashboardStats{
				TotalArtists:   42,
				TotalAlbums:    156,
				TotalLocations: 3,
				EstimatedValue: "2,450",
			},
		}

		err := t.Render(w, "home", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
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

func getArtistsHandler(logger zerolog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// do the records exist?
		artists, err := queries.ListArtists(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve artists from database")
			http.Error(w, "Artists not found", http.StatusInternalServerError)
			return
		}

		type PageData struct {
			Artists []store.Artist
		}
		data := PageData{
			Artists: artists,
		}

		// great success!
		logger.Info().Int("artists", len(artists)).Msg("List artists")
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(artists)
			return
		}
		err = t.Render(w, "index.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func getArtistHandler(logger zerolog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info().Msg("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid parameter id")
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the record exist?
		artist, err := queries.GetArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error().Err(err).Int64("artistID", artistID).Msg("Failed to retrieve artist record")
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		// great success!
		logger.Info().Int64("artistID", artistID).Str("name", artist.Name).Msg("Artist record retrieved")
		json.NewEncoder(w).Encode(artist)

	})
}

func postArtistHandler(logger zerolog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type CreateArtistRequest struct {
			Name string `json:"name" form:"name" validate:"required,min=1,max=100"`
		}

		var req CreateArtistRequest
		if err := bind(r, &req); err != nil {
			if _, ok := err.(validator.ValidationErrors); ok {
				writeValidationErrorJSON(w, err)
				return
			}
			writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		artist, err := queries.CreateArtist(r.Context(), req.Name)
		if err != nil {
			logger.Error().Err(err).Str("Name", req.Name).Msg("Failed to write artist record")
			writeErrorJSON(w, "Failed to write record", http.StatusInternalServerError)
			return
		}

		logger.Info().Int64("artistID", artist.ID).Str("Name", artist.Name).Msg("Artist record created")

		if r.Header.Get("HX-Request") == "true" {
			log.Debug().Msg("HX-Request header is present")
			tmpl := template.Must(template.ParseFiles("templates/partial/artists-row.html"))
			tmpl.Execute(w, artist)
			return
		}
		writeJSON(w, artist, http.StatusOK)
	})
}

func putArtistHandler(logger zerolog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info().Msg("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid parameter id")
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the record exist?
		artist, err := queries.GetArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error().Err(err).Int64("artistID", artistID).Msg("Failed to retrieve artist record")
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		type UpdateArtistRequest struct {
			Name string `json:"name" form:"name" validate:"required,min=1,max=100"`
		}
		var req UpdateArtistRequest

		if err := bind(r, &req); err != nil {
			if _, ok := err.(validator.ValidationErrors); ok {
				writeValidationErrorJSON(w, err)
				return
			}
			writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug().Int64("artistID", int64(artistID)).Str("artistName", req.Name).Msg("Request data after binding")

		artist, err = queries.UpdateArtist(r.Context(), store.UpdateArtistParams{ID: int64(artistID), Name: req.Name})
		if err != nil {
			logger.Error().Err(err).Str("Name", req.Name).Msg("Failed to write artist record")
			writeErrorJSON(w, "Failed to write record", http.StatusInternalServerError)
			return
		}

		logger.Info().Int64("artistID", artist.ID).Str("Name", artist.Name).Msg("Artist record updated")

		if r.Header.Get("HX-Request") == "true" {
			log.Debug().Msg("HX-Request header is present")
			tmpl := template.Must(template.ParseFiles("templates/partial/artists-row.html"))
			tmpl.Execute(w, artist)
			return
		}
		writeJSON(w, artist, http.StatusOK)
	})
}

func deleteArtistHandler(logger zerolog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info().Msg("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid parameter id")
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// delete the record
		err = queries.DeleteArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error().Err(err).Int64("artistID", artistID).Msg("Failed to retrieve artist record")
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		logger.Info().Int64("artistID", int64(artistID)).Msg("Artist record created")
		// Set the status code
		w.WriteHeader(http.StatusOK)

		// Write the response body
		fmt.Fprintf(w, "Record deleted")
	})
}

func getLocationsHandler(logger zerolog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do the records exist?
		locations, err := queries.ListLocations(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve artists from database")
			http.Error(w, "Artists not found", http.StatusInternalServerError)
			return
		}

		type PageData struct {
			Locations []store.Location
		}
		data := PageData{
			Locations: locations,
		}

		// great success!
		logger.Info().Int("artists", len(locations)).Msg("List locations")
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(locations)
		}
		err = t.Render(w, "locations.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func getRecordsHandler(logger zerolog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// do the records exist?
		records, err := queries.ListRecords(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve artists from database")
			http.Error(w, "Artists not found", http.StatusInternalServerError)
			return
		}

		type PageData struct {
			Records []store.Record
		}
		data := PageData{
			Records: records,
		}

		// great success!
		logger.Info().Int("records", len(records)).Msg("List records")
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(records)
		}
		err = t.Render(w, "albums.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func getCreateArtistForm(t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, "create-artist-form.html", nil)
	})
}

func getUpdateArtistForm(logger zerolog.Logger, queries *store.Queries, t *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info().Msg("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid parameter id")
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the record exist?
		artist, err := queries.GetArtist(r.Context(), int64(artistID))
		if err != nil {
			logger.Error().Err(err).Int64("artistID", artistID).Msg("Failed to retrieve artist record")
			http.Error(w, "Missing record", http.StatusNotFound)
			return
		}

		t.Render(w, "update-artist-form.html", artist)
	})
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

	if environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return log.Logger
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	// Get file extension
	ext := filepath.Ext(r.URL.Path)

	// Set appropriate MIME type
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	// Serve the file
	fs := http.FileServer(http.Dir("static/"))
	http.StripPrefix("/static/", fs).ServeHTTP(w, r)
}
