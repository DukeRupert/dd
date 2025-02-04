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
	UserID       int64  `json:"user_id"`
	EnableSearch bool   `json:"enable_search"`
	Artist       string `json:"artist"`
	Album        string `json:"album"`
	EnableGenre  bool   `json:"enable_genre"`
	Genre        string `json:"genre"`
	Limit        int64  `json:"limit"`
	Offset       int64  `json:"offset"`
}

type GetUserRecordsCountParams struct {
	UserID int64
	Search sql.NullString
	Genre  sql.NullString
}

const recordsPerPage = 12

func (app *application) showRecordForm(c echo.Context) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	user, err := app.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "No user found")
	}

	// Check if we're editing an existing record
	if id := c.Param("id"); id != "" {
		recordID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}

		record, err := app.queries.GetRecord(c.Request().Context(), db.GetRecordParams{
			ID:     recordID,
			UserID: userID,
		})
		if err != nil {
			return err
		}

		// Verify record belongs to user
		if record.UserID != user.ID {
			return echo.NewHTTPError(http.StatusForbidden, "Not authorized to edit this record")
		}

		return views.RecordForm(types.RecordFormPage{
			Page: types.Page{
				Title: "Edit Record",
				User:  &user,
			},
			Record: &record,
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Show form for new record
	return views.RecordForm(types.RecordFormPage{
		Page: types.Page{
			Title: "Add New Record",
			User:  &user,
		},
	}).Render(c.Request().Context(), c.Response().Writer)
}

func (app *application) createRecord(c echo.Context) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	user, err := app.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "No user found")
	}

	// Parse form data
	record := db.Record{
		Artist:    c.FormValue("artist"),
		Album:     c.FormValue("album"),
		Genre:     c.FormValue("genre"),
		Condition: c.FormValue("condition"),
		UserID:    user.ID,
	}

	// Parse and validate year
	year, err := strconv.ParseInt(c.FormValue("year"), 10, 64)
	if err != nil {
		return views.RecordForm(types.RecordFormPage{
			Page: types.Page{
				Title: "Add New Record",
				User:  &user,
			},
			Record: &record,
			FormError: map[string]string{
				"year": "Invalid year format",
			},
		}).Render(c.Request().Context(), c.Response().Writer)
	}
	record.Year = year

	// Validate input
	formErrors := make(map[string]string)
	if record.Artist == "" {
		formErrors["artist"] = "Artist is required"
	}
	if record.Album == "" {
		formErrors["album"] = "Album is required"
	}
	if record.Year < 1900 || record.Year > 2024 {
		formErrors["year"] = "Year must be between 1900 and 2024"
	}
	if record.Genre == "" {
		formErrors["genre"] = "Genre is required"
	}
	if record.Condition == "" {
		formErrors["condition"] = "Condition is required"
	}

	if len(formErrors) > 0 {
		return views.RecordForm(types.RecordFormPage{
			Page: types.Page{
				Title: "Add New Record",
				User:  &user,
			},
			Record:    &record,
			FormError: formErrors,
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Create record in database
	record, err = app.queries.CreateRecord(c.Request().Context(), db.CreateRecordParams{
		UserID:    record.UserID,
		Artist:    record.Artist,
		Album:     record.Album,
		Year:      record.Year,
		Genre:     record.Genre,
		Condition: record.Condition,
	})
	if err != nil {
		app.logger.Error().Err(err).Msg("Failed to create record")
		return err
	}

	app.logger.Info().
		Int64("id", record.ID).
		Int64("user_id", userID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record created")

	if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
		c.Response().Header().Set("HX-Redirect", "/records")
		return nil
	}

	// Redirect to records page with success message
	return c.JSON(http.StatusCreated, record)
}

func (app *application) updateRecord(c echo.Context) error {
	// Get user ID from auth
	userID, err := auth.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get user
	user, err := app.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "No user found")
	}

	// Get record ID from URL parameter
	recordID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}

	// Check if record exists and belongs to user
	existingRecord, err := app.queries.GetRecord(c.Request().Context(), db.GetRecordParams{
		ID:     recordID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}
	if existingRecord.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "Not authorized to update this record")
	}

	// Parse form data
	record := db.Record{
		ID:        recordID,
		Artist:    c.FormValue("artist"),
		Album:     c.FormValue("album"),
		Genre:     c.FormValue("genre"),
		Condition: c.FormValue("condition"),
		UserID:    user.ID,
	}

	// Parse and validate year
	year, err := strconv.ParseInt(c.FormValue("year"), 10, 64)
	if err != nil {
		return views.RecordForm(types.RecordFormPage{
			Page: types.Page{
				Title: "Edit Record",
				User:  &user,
			},
			Record: &record,
			FormError: map[string]string{
				"year": "Invalid year format",
			},
		}).Render(c.Request().Context(), c.Response().Writer)
	}
	record.Year = year

	// Validate input
	formErrors := make(map[string]string)
	if record.Artist == "" {
		formErrors["artist"] = "Artist is required"
	}
	if record.Album == "" {
		formErrors["album"] = "Album is required"
	}
	if record.Year < 1900 || record.Year > 2024 {
		formErrors["year"] = "Year must be between 1900 and 2024"
	}
	if record.Genre == "" {
		formErrors["genre"] = "Genre is required"
	}
	if record.Condition == "" {
		formErrors["condition"] = "Condition is required"
	}

	if len(formErrors) > 0 {
		return views.RecordForm(types.RecordFormPage{
			Page: types.Page{
				Title: "Edit Record",
				User:  &user,
			},
			Record:    &record,
			FormError: formErrors,
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Update record in database
	updatedRecord, err := app.queries.UpdateRecord(c.Request().Context(), db.UpdateRecordParams{
		ID:        recordID,
		UserID:    record.UserID,
		Artist:    record.Artist,
		Album:     record.Album,
		Year:      record.Year,
		Genre:     record.Genre,
		Condition: record.Condition,
	})
	if err != nil {
		app.logger.Error().Err(err).Msg("Failed to update record")
		return err
	}

	app.logger.Info().
		Int64("id", recordID).
		Int64("user_id", userID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record updated")

	if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
		c.Response().Header().Set("HX-Redirect", "/records")
		return nil
	}

	// Return updated record
	return c.JSON(http.StatusOK, updatedRecord)
}

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

	// Calculate offset
	offset := int64((page - 1) * recordsPerPage)

	var records []db.Record
	if sort == "created_asc" {
		records, err = app.queries.GetUserRecordsAsc(c.Request().Context(), db.GetUserRecordsAscParams{
			UserID: userID,
			Artist: "%" + search + "%",
			Album:  "%" + search + "%",
			Limit:  recordsPerPage,
			Offset: offset,
		})
		if err != nil {
			return err
		}
	} else {
		records, err = app.queries.GetUserRecords(c.Request().Context(), db.GetUserRecordsParams{
			UserID: userID,
			Artist: "%" + search + "%",
			Album:  "%" + search + "%",
			Limit:  recordsPerPage,
			Offset: offset,
		})
		if err != nil {
			return err
		}
	}

	// Get total count for pagination
	total, err := app.queries.GetUserRecordsCount(c.Request().Context(), db.GetUserRecordsCountParams{
		UserID: userID,
		Artist: "%" + search + "%",
		Album:  "%" + search + "%",
	})
	if err != nil {
		return err
	}

	totalPages := (int(total) + recordsPerPage - 1) / recordsPerPage

	app.logger.Info().
		Str("route", "/records").
		Int64("user_id", userID).
		Int("records_found", len(records)).
		Str("search", search).
		Int("page", page).
		Msg("Records retrieved")

	// If this is a JSON request, retern JSON
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusOK, records)
	}

	// If this is an HTMX request just return the records
	if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
		return views.Records(types.RecordsPage{
			Records:     records,
			CurrentPage: page,
			TotalPages:  totalPages,
			SortBy:      "sortBy",
			SortOrder:   "sortOrder",
			Genre:       genre,
			Search:      search,
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Otherwise render the full page
	return views.RecordsPage(types.RecordsPage{
		Records:     records,
		CurrentPage: page,
		TotalPages:  totalPages,
		SortBy:      "sortBy",
		SortOrder:   "sortOrder",
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

func (app *application) deleteRecord(c echo.Context) error {
	// Get user ID from context
	userID, err := auth.GetUserID(c)
	if err != nil {
		app.logger.Error().Err(err).Msg("Authentication failed in delete handler")
		return api.NewUnauthorizedError("Not authenticated")
	}

	// Get and validate record ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		app.logger.Error().
			Err(err).
			Str("id_param", c.Param("id")).
			Msg("Invalid record ID format in delete request")
		return api.NewBadRequestError("invalid id format")
	}

	// Verify record exists and belongs to user before deleting
	record, err := app.queries.GetRecord(c.Request().Context(), db.GetRecordParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			app.logger.Info().
				Int64("id", id).
				Int64("user_id", userID).
				Msg("Attempted to delete non-existent record")
			return api.NewNotFoundError("record not found")
		}
		app.logger.Error().Err(err).Msg("Database error when fetching record")
		return api.NewDatabaseError(err)
	}

	// Check record ownership
	if record.UserID != userID {
		app.logger.Warn().
			Int64("record_id", id).
			Int64("user_id", userID).
			Int64("record_owner_id", record.UserID).
			Msg("Unauthorized attempt to delete record")
		return api.NewForbiddenError("not authorized to delete this record")
	}

	// Delete the record
	err = app.queries.DeleteRecord(c.Request().Context(), db.DeleteRecordParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		app.logger.Error().
			Err(err).
			Int64("id", id).
			Int64("user_id", userID).
			Msg("Failed to delete record")
		return api.NewDatabaseError(err)
	}

	app.logger.Info().
		Int64("id", id).
		Int64("user_id", userID).
		Str("artist", record.Artist).
		Str("album", record.Album).
		Msg("Record deleted successfully")

	// For HTMX requests, return 200 with empty body to remove the element
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.String(http.StatusOK, "")
	}

	// For non-HTMX requests, return 204 No Content
	return c.NoContent(http.StatusNoContent)
}
