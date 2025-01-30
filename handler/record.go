package handler

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/auth"
	"github.com/dukerupert/dd/db"
	"github.com/labstack/echo/v4"
)

type createRecordRequest struct {
	Artist    string `json:"artist" validate:"required,min=1,max=100"`
	Album     string `json:"album" validate:"required,min=1,max=100"`
	Year      int64  `json:"year" validate:"required,min=1900,max=2100"`
	Genre     string `json:"genre" validate:"required,min=1,max=50"`
	Condition string `json:"condition" validate:"required,oneof=Mint Near-Mint Very-Good Good Fair Poor"`
}

func (app *application) getAllRecords(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	records, err := app.queries.ListRecords(context.Background(), userID)
	if err != nil {
		return api.NewDatabaseError(err)
	}

	app.logger.Debug().
		Int64("user_id", userID).
		Int("count", len(records)).
		Msg("Records retrieved")
	
	return c.JSON(http.StatusOK, records)
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