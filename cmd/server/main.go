package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/dukerupert/dd/config"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	dbConfig, serverConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	config.PrintConfig(dbConfig, serverConfig)

	// Connect to database
	db, err := sql.Open("postgres", dbConfig.GetConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(dbConfig.MaxConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database!")

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
