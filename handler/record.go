package handler

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/auth"
	"github.com/dukerupert/dd/db"
	"github.com/dukerupert/dd/internal/types"
	"github.com/dukerupert/dd/views"
	"github.com/labstack/echo/v4"
)

type createRecordRequest struct {
	Artist    string `json:"artist" validate:"required,min=1,max=100"`
	Album     string `json:"album" validate:"required,min=1,max=100"`
	Year      int64  `json:"year" validate:"required,min=1900,max=2100"`
	Genre     string `json:"genre" validate:"required,min=1,max=50"`
	Condition string `json:"condition" validate:"required,oneof=Mint Near-Mint Very-Good Good Fair Poor"`
}

type GetUserRecordsParams struct {
    UserID     int64
    Search     sql.NullString
    Genre      sql.NullString
    SortBy     string
    SortOrder  string
    Limit      int32
    Offset     int32
}

type GetUserRecordsCountParams struct {
    UserID     int64
    Search     sql.NullString
    Genre      sql.NullString
}

const recordsPerPage = 12

func (app *application) getAllRecords(c echo.Context) error {
    // Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

    // Get query parameters
    page, _ := strconv.Atoi(c.QueryParam("page"))
    if page < 1 {
        page = 1
    }
    
    search := c.QueryParam("search")
    sort := c.QueryParam("sort")
    genre := c.QueryParam("genre")

    // Parse sort parameter
    sortBy := "created_at"
    sortOrder := "desc"
    if sort != "" {
        switch sort {
        case "album_asc":
            sortBy, sortOrder = "album", "asc"
        case "album_desc":
            sortBy, sortOrder = "album", "desc"
        case "artist_asc":
            sortBy, sortOrder = "artist", "asc"
        case "artist_desc":
            sortBy, sortOrder = "artist", "desc"
        case "year_asc":
            sortBy, sortOrder = "year", "asc"
        case "year_desc":
            sortBy, sortOrder = "year", "desc"
        }
    }

    // Calculate offset
    offset := int64((page - 1) * recordsPerPage)

    // Get records from database
    records, err := app.queries.GetUserRecords(c.Request().Context(), db.GetUserRecordsParams{
		UserID: userID,
		Column4: search != "", // boolean to enable/disable search
		Artist: "%" + search + "%",
		Album:  "%" + search + "%",
		Genre:  genre,
		Limit:  recordsPerPage,
		Offset: offset,
	})
    if err != nil {
        return err
    }

    // Get total count for pagination
    total, err := app.queries.GetUserRecordsCount(c.Request().Context(), db.GetUserRecordsCountParams{
		UserID: userID,
		Column4: search != "", // boolean to enable/disable search
		Artist: "%" + search + "%",
		Album:  "%" + search + "%",
		Genre:  genre,
	})
    if err != nil {
        return err
    }

    totalPages := (int(total) + recordsPerPage - 1) / recordsPerPage

	app.logger.Debug().
		Int64("user_id", userID).
		Int("count", len(records)).
		Msg("Records retrieved")
	
	// If this is a JSON request, retern JSON
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusOK, records)
	}

    // Otherwise render the full page
    return views.Records(types.RecordsPage{
		Records:     records,
		CurrentPage: page,
		TotalPages:  totalPages,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
		Genre:       genre,
		Search:      search,
	}).Render(c.Request().Context(), c.Response().Writer)
}

func (app *application) getRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return api.NewBadRequestError("invalid id format")
	}

	record, err := app.queries.GetRecord(context.Background(), db.GetRecordParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewNotFoundError("record")
		}
		return api.NewDatabaseError(err)
	}

	return c.JSON(http.StatusOK, record)
}

func (app *application) createRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	// Validate the request
	if err := c.Validate(&req); err != nil {
		return err // Our custom validator already returns an api.ValidationError
	}

	params := db.CreateRecordParams{
		UserID:    userID,
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.CreateRecord(context.Background(), params)
	if err != nil {
		return api.NewDatabaseError(err)
	}

	app.logger.Info().
		Int64("id", record.ID).
		Int64("user_id", userID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record created")

	return c.JSON(http.StatusCreated, record)
}

func (app *application) updateRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return api.NewBadRequestError("invalid id format")
	}

	var req createRecordRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	// Validate the request
	if err := c.Validate(&req); err != nil {
		return err // Our custom validator already returns an api.ValidationError
	}

	params := db.UpdateRecordParams{
		ID:        id,
		UserID:    userID,
		Artist:    req.Artist,
		Album:     req.Album,
		Year:      req.Year,
		Genre:     req.Genre,
		Condition: req.Condition,
	}

	record, err := app.queries.UpdateRecord(context.Background(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewNotFoundError("record")
		}
		return api.NewDatabaseError(err)
	}

	app.logger.Info().
		Int64("id", record.ID).
		Int64("user_id", userID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record updated")

	return c.JSON(http.StatusOK, record)
}

func (app *application) deleteRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return api.NewBadRequestError("invalid id format")
	}

	err = app.queries.DeleteRecord(context.Background(), db.DeleteRecordParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewNotFoundError("record")
		}
		return api.NewDatabaseError(err)
	}

	app.logger.Info().
		Int64("id", id).
		Int64("user_id", userID).
		Msg("Record deleted")

	return c.NoContent(http.StatusNoContent)
}