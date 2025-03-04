package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/dukerupert/dd/pocketbase"
	"github.com/rs/zerolog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	logger := initLogger()

	// Parse template files
	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}
	// Create a new Echo instance
	e := echo.New()
	e.Renderer = t

	// Middleware
	headers := []string{echo.HeaderAuthorization}
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogMethod:  true,
		LogHeaders: headers,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// Create a map to store header values
			headerValues := make(map[string]string)

			// Extract each header value from the request
			for _, header := range headers {
				headerValues[header] = c.Request().Header.Get(header)
			}
			logger.Info().
				Str("Method", v.Method).
				Str("URI", v.URI).
				Int("Status", v.Status).
				Fields(map[string]interface{}{
					"headers": headers,
				}).Fields(map[string]interface{}{
				"headers": headerValues,
			}).
				Msg("request")

			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create a new client
	client := pocketbase.NewClient("https://pb.angmar.dev", logger)

	// Authenticate with credentials
	err := client.Authenticate("valid@example.com", "valid123")
	if err != nil {
		logger.Error().Err(err).Msg("Authentication failed")
		return
	}

	// Make a GET request
	response, err := client.Get("/api/collections/albums/records")
	if err != nil {
		logger.Error().Err(err).Msg("GET request failed")
		return
	}

	fmt.Println("Response:", string(response))


	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Vinyl Collection API")
	})
	e.GET("/hello", Hello)

	// Record endpoints

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func initLogger() *zerolog.Logger {
	// Define the log level flag
	logLevelFlag := flag.String("log-level", "info", "sets log level: panic, fatal, error, warn, info, debug, trace")
	flag.Parse()
	
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	
	// Set log level based on parameter
	switch strings.ToLower(*logLevelFlag) {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		// Default to info level if invalid level provided
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	
	// Create the main logger
	var logger zerolog.Logger
	
	// Check if debug level or lower is set for pretty console logging
	if zerolog.GlobalLevel() <= zerolog.DebugLevel {
		// Pretty console logging in debug mode
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		// JSON logging for production
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
	
	return &logger
}


// Types
// Create struct to unmarshal the response
type AuthResponse struct {
	Record struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		// Add other fields if needed
	} `json:"record"`
	Token string `json:"token"`
}

type User struct {
	AuthProviders    []interface{} `json:"authProviders"`
	UsernamePassword bool          `json:"usernamePassword"`
	EmailPassword    bool          `json:"emailPassword"`
	OnlyVerified     bool          `json:"onlyVerified"`
}

// PocketBaseTime is a custom time type to handle PocketBase's time format
type PocketBaseTime time.Time

// UnmarshalJSON implements json.Unmarshaler
func (pbt *PocketBaseTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		*pbt = PocketBaseTime(time.Time{})
		return nil
	}

	// PocketBase format: "2025-02-24 03:27:41.065Z"
	t, err := time.Parse("2006-01-02 15:04:05.999Z", s)
	if err != nil {
		// Try alternative formats if the first one fails
		t, err = time.Parse("2006-01-02 15:04:05.999", s)
		if err != nil {
			return err
		}
	}
	
	*pbt = PocketBaseTime(t)
	return nil
}

// MarshalJSON implements json.Marshaler
func (pbt PocketBaseTime) MarshalJSON() ([]byte, error) {
	t := time.Time(pbt)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format("2006-01-02 15:04:05.999Z"))
}

// String returns the time as a formatted string
func (pbt PocketBaseTime) String() string {
	t := time.Time(pbt)
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05.999Z")
}

// Time returns the underlying time.Time
func (pbt PocketBaseTime) Time() time.Time {
	return time.Time(pbt)
}

// Album represents an album record from PocketBase
type Album struct {
	CollectionID   string         `json:"collectionId"`
	CollectionName string         `json:"collectionName"`
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	ArtistID       string         `json:"artist_id"`
	LocationID     string         `json:"location_id"`
	ReleaseYear    int            `json:"release_year"`
	Genre          string         `json:"genre"`
	Condition      string         `json:"condition"`
	PurchaseDate   PocketBaseTime `json:"purchase_date"`
	PurchasePrice  float64        `json:"purchase_price"`
	Notes          string         `json:"notes"`
	Created        PocketBaseTime `json:"created"`
	Updated        PocketBaseTime `json:"updated"`
}

// AlbumListResponse represents a paginated list of albums
type AlbumListResponse struct {
	Page       int     `json:"page"`
	PerPage    int     `json:"perPage"`
	TotalPages int     `json:"totalPages"`
	TotalItems int     `json:"totalItems"`
	Items      []Album `json:"items"`
}

// Handlers

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Hello(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderAcceptEncoding, "application/json")
	return c.Render(http.StatusOK, "hello", "World")
}
