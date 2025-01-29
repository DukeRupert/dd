package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

// Record represents a vinyl record in the collection
type Record struct {
	ID        string `json:"id"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Year      int    `json:"year"`
	Genre     string `json:"genre"`
	Condition string `json:"condition"`
}

func main() {
	// Create a new Echo instance
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
	e.GET("/records", getAllRecords)
	e.GET("/records/:id", getRecord)
	e.POST("/records", createRecord)
	e.PUT("/records/:id", updateRecord)
	e.DELETE("/records/:id", deleteRecord)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// Handler functions
func getAllRecords(c echo.Context) error {
	// TODO: Implement database integration
	records := []Record{
		{
			ID:        "1",
			Artist:    "Pink Floyd",
			Album:     "Dark Side of the Moon",
			Year:      1973,
			Genre:     "Progressive Rock",
			Condition: "Very Good",
		},
	}
	return c.JSON(http.StatusOK, records)
}

func getRecord(c echo.Context) error {
	id := c.Param("id")
	// TODO: Implement database lookup
	record := Record{
		ID:        id,
		Artist:    "Pink Floyd",
		Album:     "Dark Side of the Moon",
		Year:      1973,
		Genre:     "Progressive Rock",
		Condition: "Very Good",
	}
	return c.JSON(http.StatusOK, record)
}

func createRecord(c echo.Context) error {
	record := new(Record)
	if err := c.Bind(record); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// TODO: Implement database creation
	return c.JSON(http.StatusCreated, record)
}

func updateRecord(c echo.Context) error {
	id := c.Param("id")
	record := new(Record)
	if err := c.Bind(record); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	record.ID = id
	// TODO: Implement database update
	return c.JSON(http.StatusOK, record)
}

func deleteRecord(c echo.Context) error {
	// id := c.Param("id")
	// TODO: Implement database deletion
	return c.NoContent(http.StatusNoContent)
}
