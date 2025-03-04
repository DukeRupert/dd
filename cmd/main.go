package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

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

	// Example: Fetch albums with pagination, sorting, and expansion
    params := pocketbase.QueryParams{
        Page:     1,
        PerPage:  10,
        Sort:     "-created,title", // Sort by created desc, then title asc
        Expand:   "artist_id,location_id", // Expand related records
    }
    
    albums, err := client.ListAlbums(params)
    if err != nil {
        logger.Error().Err(err).Msg("Failed to fetch albums")
        return
    }
    
    // Process the results
    logger.Info().
        Int("total_albums", albums.TotalItems).
        Int("current_page", albums.Page).
        Int("per_page", albums.PerPage).
        Msg("Fetched albums")
    
    // Access the typed results
    for _, album := range albums.Items {
        logger.Info().
            Str("id", album.ID).
            Str("title", album.Title).
            Int("year", album.ReleaseYear).
            Str("genre", album.Genre).
            Msg("Album details")
        
        // Access expanded relations if available
        if album.Expand.Artist != nil {
            logger.Info().
                Str("artist_id", album.ArtistID).
                Str("artist_name", album.Expand.Artist.Name).
                Msg("Album artist")
        }
    }

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
