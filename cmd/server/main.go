package main

import (
	"log"
	"net/http"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/internal/db"

	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	dbConfig, serverConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	config.PrintConfig(dbConfig, serverConfig)

	// In your main function
	database, err := db.NewDB(dbConfig)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.Close()

	// Use the database connection
	rows, err := database.DB.Query("SELECT * FROM artists")
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}
	defer rows.Close()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
