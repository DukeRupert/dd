package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
)

// HTML Handlers

func handleGetArtistsPage(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all artists
		// - Query database
		// - Handle search query param
		// - Render artists list page
		artists, err := queries.ListArtists(r.Context())
		if err != nil {
			logger.Error("Failed to retrieve artists", slog.String("error", err.Error()))
			http.Error(w, "Failed to retrieve artists", http.StatusInternalServerError)
			return
		}

		renderer.Render(w, "artists", map[string]interface{}{
			"Title":   "Artists",
			"Artists": artists,
		})
	})
}

func handleGetArtistNewForm(renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render new artist form
		renderer.Render(w, "create-artist-form", nil)
	})
}

func handlePostArtist(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type CreateArtistRequest struct {
			Name string `form:"name" validate:"required,min=1,max=100"`
		}

		var req CreateArtistRequest
		if err := bind(r, &req); err != nil {
			// Check if it's a validation error
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(formatValidationErrorsHTML(validationErrs)))
				return
			}
			// Other binding errors
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		artist, err := queries.CreateArtist(r.Context(), req.Name)
		if err != nil {
			logger.Error("Failed to create artist", slog.String("error", err.Error()), slog.String("name", req.Name))
			http.Error(w, "Failed to create artist", http.StatusInternalServerError)
			return
		}

		logger.Info("Artist created", slog.Int64("artistID", artist.ID), slog.String("name", artist.Name))

		// Render the artist row partial
		err = renderer.Render(w, "artists-row", artist)
		if err != nil {
			logger.Error("Failed to render template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleGetArtist(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			logger.Info("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Error("Invalid parameter id", slog.String("error", err.Error()))
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the artist exist?
		artist, err := queries.GetArtist(r.Context(), artistID)
		if err != nil {
			logger.Error("Failed to retrieve artist", slog.String("error", err.Error()), slog.Int64("artistID", artistID))
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}

		// fetch artist's records
		records, err := queries.GetRecordsByArtist(r.Context(), sql.NullInt64{Int64: artistID, Valid: true})
		if err != nil {
			logger.Error("Failed to retrieve artist records", slog.String("error", err.Error()), slog.Int64("artistID", artistID))
			// Not fatal - just show empty records list
			records = []store.Record{}
		}

		// render artist detail page
		logger.Info("Artist retrieved", slog.Int64("artistID", artistID), slog.String("name", artist.Name), slog.Int("recordCount", len(records)))
		
		err = renderer.Render(w, "artist-detail", map[string]interface{}{
			"Title":   artist.Name,
			"Artist":  artist,
			"Records": records,
		})
		if err != nil {
			logger.Error("Failed to render template", slog.String("error", err.Error()))
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	})
}

func handleGetArtistEditForm(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render edit artist form
		// - Parse ID from path
		// - Query artist
		// - Render edit form with current data
		id := r.PathValue("id")
		logger.Info("Get artist edit form handler called", slog.String("id", id))
		renderer.Render(w, "update-artist-form", nil)
	})
}

func handlePutArtist(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update artist
		// - Parse ID from path
		// - Parse form data
		// - Validate input
		// - Update artist in database
		// - Return updated artist row partial for HTMX
		id := r.PathValue("id")
		logger.Info("Update artist handler called", slog.String("id", id))
		renderer.Render(w, "artists-row", nil)
	})
}

func handleDeleteArtist(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete artist
		// - Parse ID from path
		// - Check if artist has records (maybe prevent deletion or cascade)
		// - Delete artist from database
		// - Return 200 OK for HTMX to remove row
		id := r.PathValue("id")
		logger.Info("Delete artist handler called", slog.String("id", id))
		w.WriteHeader(http.StatusOK)
	})
}

// API Handlers

func handleAPIGetArtists(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get artists list
		// - Parse query params (page, limit, search)
		// - Query database with filters
		// - Return JSON array
		artists, err := queries.ListArtists(r.Context())
		if err != nil {
			logger.Error("Failed to retrieve artists", slog.String("error", err.Error()))
			writeErrorJSON(w, "Failed to retrieve artists", http.StatusInternalServerError)
			return
		}
		writeJSON(w, artists, http.StatusOK)
	})
}

func handleAPIPostArtist(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type CreateArtistRequest struct {
			Name string `json:"name" validate:"required,min=1,max=100"`
		}

		var req CreateArtistRequest
		if err := bind(r, &req); err != nil {
			// Check if it's a validation error
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ValidationErrorResponse{
					Error:   "Validation failed",
					Message: "Please check your input",
					Details: getValidationErrors(validationErrs),
				})
				return
			}
			// Other binding errors
			writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		artist, err := queries.CreateArtist(r.Context(), req.Name)
		if err != nil {
			logger.Error("Failed to create artist", slog.String("error", err.Error()), slog.String("name", req.Name))
			writeErrorJSON(w, "Failed to create artist", http.StatusInternalServerError)
			return
		}

		logger.Info("Artist created via API", slog.Int64("artistID", artist.ID), slog.String("name", artist.Name))

		writeJSON(w, artist, http.StatusCreated)
	})
}

func handleAPIGetArtist(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get single artist
		// - Parse ID from path
		// - Query artist from database
		// - Return JSON
		id := r.PathValue("id")
		artistID, _ := strconv.ParseInt(id, 10, 64)
		artist, err := queries.GetArtist(r.Context(), artistID)
		if err != nil {
			logger.Error("Failed to retrieve artist", slog.String("error", err.Error()))
			writeErrorJSON(w, "Artist not found", http.StatusNotFound)
			return
		}
		writeJSON(w, artist, http.StatusOK)
	})
}

func handleAPIPutArtist(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update artist via API
		// - Parse ID from path
		// - Parse JSON body
		// - Validate input
		// - Update artist in database
		// - Return updated artist
		id := r.PathValue("id")
		logger.Info("API update artist handler called", slog.String("id", id))
		writeJSON(w, map[string]string{"message": "artist updated"}, http.StatusOK)
	})
}

func handleAPIDeleteArtist(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete artist via API
		// - Parse ID from path
		// - Delete artist from database
		// - Return 204 No Content
		id := r.PathValue("id")
		logger.Info("API delete artist handler called", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)
	})
}

func handleAPIGetArtistRecords(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all records by artist
		// - Parse ID from path
		// - Query records for artist
		// - Return JSON array
		id := r.PathValue("id")
		logger.Info("API get artist records handler called", slog.String("id", id))
		writeJSON(w, []interface{}{}, http.StatusOK)
	})
}