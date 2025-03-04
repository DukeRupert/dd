// pkg/handler/handler.go
package handler

import (
	"net/http"
	"github.com/dukerupert/dd/pkg/pocketbase" // Adjust this import path to match your project

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Handler contains all HTTP handler for your application
type Handler struct {
	Client *pocketbase.Client
	Logger *zerolog.Logger
}

// New creates a new Handler instance
func New(client *pocketbase.Client, logger *zerolog.Logger) *Handler {
	return &Handler{
		Client: client,
		Logger: logger,
	}
}

// AlbumsHandler returns all albums
func (h *Handler) AlbumsHandler(c echo.Context) error {
	h.Logger.Debug().Msg("Handling albums request")

	// Create default query parameters
	params := pocketbase.QueryParams{
		Page:     1,
		PerPage:  5,
		Sort:     "title",
		Expand:   "artist_id,location_id",
	}

	// Fetch albums
	albums, err := h.Client.ListAlbums(params)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to fetch albums")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch albums",
		})
	}

	h.Logger.Info().Int("count", len(albums.Items)).Msg("Successfully fetched albums")
	return c.JSON(http.StatusOK, albums)
}

// AlbumByIDHandler returns a single album by ID
func (h *Handler) AlbumByIDHandler(c echo.Context) error {
	// Get album ID from URL parameters
	albumID := c.Param("id")

	h.Logger.Debug().Str("album_id", albumID).Msg("Handling album by ID request")

	// Fetch album
	album, err := h.Client.GetAlbum(albumID)
	if err != nil {
		h.Logger.Error().Err(err).Str("id", albumID).Msg("Failed to fetch album")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch album",
		})
	}

	h.Logger.Info().Str("id", album.ID).Str("title", album.Title).Msg("Successfully fetched album")
	return c.JSON(http.StatusOK, album)
}

// ArtistsHandler returns all albums
func (h *Handler) ArtistsHandler(c echo.Context) error {
	h.Logger.Debug().Msg("Handling artists request")

	// Create default query parameters
	params := pocketbase.QueryParams{
		Page:     1,
		PerPage:  5,
		Sort:     "name",
		Expand:   "",
	}

	// Fetch albums
	artists, err := h.Client.ListArtists(params)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to fetch artists")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch artists",
		})
	}

	h.Logger.Info().Int("count", len(artists.Items)).Msg("Successfully fetched albums")
	return c.JSON(http.StatusOK, artists)
}

// ArtistByIDHandler returns a single album by ID
func (h *Handler) ArtistByIDHandler(c echo.Context) error {
	// Get album ID from URL parameters
	artistID := c.Param("id")

	h.Logger.Debug().Str("artist_id", artistID).Msg("Handling artist by ID request")

	// Fetch album
	artist, err := h.Client.GetArtist(artistID)
	if err != nil {
		h.Logger.Error().Err(err).Str("id", artistID).Msg("Failed to fetch artist")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch artist",
		})
	}

	h.Logger.Info().Str("id", artist.ID).Str("name", artist.Name).Msg("Successfully fetched artist")
	return c.JSON(http.StatusOK, artist)
}

// RegisterRoutes registers all routes to the Echo instance
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/albums", h.AlbumsHandler)
	e.GET("/albums/:id", h.AlbumByIDHandler)
	e.GET("/artists", h.ArtistsHandler)
	e.GET("/artists/:id", h.ArtistByIDHandler)
}