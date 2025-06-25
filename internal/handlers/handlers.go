package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/dukerupert/dd/internal/database"
	"github.com/dukerupert/dd/internal/db"
	"github.com/dukerupert/dd/internal/logger"
)

// Handlers contains all handler dependencies
type Handlers struct {
	db      *database.Database
	queries *db.Queries
	logger  zerolog.Logger
}

// NewHandlers creates a new handlers instance with dependencies
func NewHandlers(database *database.Database) *Handlers {
	return &Handlers{
		db:      database,
		queries: database.Queries,
		logger:  logger.Get(),
	}
}

// CreateArtist handles POST /artists
func (h *Handlers) CreateArtist(c echo.Context) error {
	ctx := c.Request().Context()
	ctxLogger := logger.WithContext(c)

	var req struct {
		Name string `json:"name" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		ctxLogger.Warn().Err(err).Msg("Failed to bind artist data")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Artist name is required")
	}

	// Check if artist already exists
	existing, err := h.queries.GetArtistByName(ctx, req.Name)
	if err == nil {
		ctxLogger.Info().
			Int32("existing_id", existing.ID).
			Str("name", req.Name).
			Msg("Artist already exists")
		return echo.NewHTTPError(http.StatusConflict, "Artist already exists")
	}

	// Create new artist
	artist, err := h.queries.CreateArtist(ctx, req.Name)
	if err != nil {
		ctxLogger.Error().Err(err).Msg("Failed to create artist")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create artist")
	}

	ctxLogger.Info().
		Int32("artist_id", artist.ID).
		Str("name", artist.Name).
		Msg("Artist created successfully")

	return c.JSON(http.StatusCreated, artist)
}

// GetArtist handles GET /artists/:id
func (h *Handlers) GetArtist(c echo.Context) error {
	ctx := c.Request().Context()
	ctxLogger := logger.WithContext(c)

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Artist ID is required")
	}

	// Convert string to int32 (you might want to use a helper for this)
	var artistID int32
	if _, err := fmt.Sscanf(id, "%d", &artistID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid artist ID")
	}

	artist, err := h.queries.GetArtist(ctx, artistID)
	if err != nil {
		ctxLogger.Warn().
			Err(err).
			Str("artist_id", id).
			Msg("Artist not found")
		return echo.NewHTTPError(http.StatusNotFound, "Artist not found")
	}

	return c.JSON(http.StatusOK, artist)
}

// ListArtists handles GET /artists
func (h *Handlers) ListArtists(c echo.Context) error {
	ctx := c.Request().Context()
	ctxLogger := logger.WithContext(c)

	artists, err := h.queries.ListArtists(ctx)
	if err != nil {
		ctxLogger.Error().Err(err).Msg("Failed to list artists")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve artists")
	}

	ctxLogger.Info().Int("count", len(artists)).Msg("Listed artists")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"artists": artists,
		"count":   len(artists),
	})
}

// UpdateArtist handles PUT /artists/:id
func (h *Handlers) UpdateArtist(c echo.Context) error {
	ctx := c.Request().Context()
	ctxLogger := logger.WithContext(c)

	id := c.Param("id")
	var artistID int32
	if _, err := fmt.Sscanf(id, "%d", &artistID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid artist ID")
	}

	var req struct {
		Name string `json:"name" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	artist, err := h.queries.UpdateArtist(ctx, db.UpdateArtistParams{
		ID:   artistID,
		Name: req.Name,
	})
	if err != nil {
		ctxLogger.Error().Err(err).Int32("artist_id", artistID).Msg("Failed to update artist")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update artist")
	}

	ctxLogger.Info().
		Int32("artist_id", artistID).
		Str("new_name", req.Name).
		Msg("Artist updated successfully")

	return c.JSON(http.StatusOK, artist)
}

// DeleteArtist handles DELETE /artists/:id
func (h *Handlers) DeleteArtist(c echo.Context) error {
	ctx := c.Request().Context()
	ctxLogger := logger.WithContext(c)

	id := c.Param("id")
	var artistID int32
	if _, err := fmt.Sscanf(id, "%d", &artistID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid artist ID")
	}

	err := h.queries.DeleteArtist(ctx, artistID)
	if err != nil {
		ctxLogger.Error().Err(err).Int32("artist_id", artistID).Msg("Failed to delete artist")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete artist")
	}

	ctxLogger.Info().Int32("artist_id", artistID).Msg("Artist deleted successfully")

	return c.NoContent(http.StatusNoContent)
}