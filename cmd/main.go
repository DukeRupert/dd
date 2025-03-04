package main

import (
	"io"
	"net/http"
	"text/template"

	"github.com/dukerupert/dd/api/handler"
	"github.com/dukerupert/dd/pkg/logger"
	"github.com/dukerupert/dd/pkg/pocketbase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize logger
	log := logger.InitLogger()

	// Create PocketBase client
	client := pocketbase.NewClient("https://pb.angmar.dev", log)

	// Authenticate
	err := client.Authenticate("valid@example.com", "valid123")
	if err != nil {
		log.Error().Err(err).Msg("Authentication failed")
		return
	}

	// Parse templates
	t := &Template{
		templates: template.Must(template.ParseGlob("web/views/*.html")),
	}
	// Also parse the partials
	template.Must(t.templates.ParseGlob("web/views/partials/*.html"))

	// Create Echo instance
	e := echo.New()
	e.Renderer = t

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create handlers and register routes
	handler := handler.New(client, log)
	handler.RegisterRoutes(e)

	// Start server
	log.Info().Msg("Starting server on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Error().Err(err).Msg("Server error")
	}
}

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
