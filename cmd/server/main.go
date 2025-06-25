package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/internal/database"
	

	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	dbConfig, _, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// config.PrintConfig(dbConfig, serverConfig)

	// Create database connection
	db, err := database.New(dbConfig)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

	// Use the queries
	ctx := context.Background()
	_, err = db.Queries.CreateArtist(ctx, "The Beatles")
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
