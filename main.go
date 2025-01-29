package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/db"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	queries *db.Queries
}

func main() {
	// Open SQLite database
	sqlite, err := sql.Open("sqlite3", "vinyl.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqlite.Close()

	// Create tables if they don't exist
	if err := initDatabase(sqlite); err != nil {
		log.Fatal(err)
	}

	// Initialize queries
	app := &application{
		queries: db.New(sqlite),
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Vinyl Collection API")
	})

	// Record endpoints
	e.GET("/records", app.getAllRecords)
	e.GET("/records/:id", app.getRecord)
	e.POST("/records", app.createRecord)
	e.PUT("/records/:id", app.updateRecord)
	e.DELETE("/records/:id", app.deleteRecord)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func initDatabase(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		artist TEXT NOT NULL,
		album TEXT NOT NULL,
		year INTEGER NOT NULL,
		genre TEXT NOT NULL,
		condition TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(schema)
	return err
}

type createRecordRequest struct {
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Year      int64  `json:"year"`
	Genre     string `json:"genre"`
	Condition string `json:"condition"`
}

func (app *application) getAllRecords(c echo.Context) error {
	records, err := app.queries.ListRecords(context.Background())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, records)
}

func (app *application) getRecord(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	record, err := app.queries.GetRecord(context.Background(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "record not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, record)
}

func (app *application) createRecord(c echo.Context) error {
	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	params := db.CreateRecordParams{
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.CreateRecord(context.Background(), params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, record)
}

func (app *application) updateRecord(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	params := db.UpdateRecordParams{
		ID:        id,
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.UpdateRecord(context.Background(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "record not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, record)
}

func (app *application) deleteRecord(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	err = app.queries.DeleteRecord(context.Background(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "record not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}