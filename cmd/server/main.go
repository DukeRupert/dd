package main

import (
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

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/artists", func(c echo.Context) error  {
		ctx := c.Request().Context()

		artists, err := db.Queries.ListArtists(ctx)
		if err != nil {
			log.Default().Printf("Failed to retrieve artists from database: %s", err)
		}

		return c.JSON(http.StatusOK, artists)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
