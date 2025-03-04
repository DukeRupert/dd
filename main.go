package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	// Configure global settings
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create the main logger
	var logger zerolog.Logger
	if *debug {
		// Pretty console logging in debug mode
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		// JSON logging for production
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

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

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Vinyl Collection API")
	})
	e.GET("/hello", Hello)

	// Record endpoints

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
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
