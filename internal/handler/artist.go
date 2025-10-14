package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
)

// HTML Endpoints

// GET /artists
func (h *Handler) GetArtists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all artists
		// - Query database
		// - Handle search query param
		// - Render artists list page
		artists, err := h.queries.ListArtists(r.Context())
		if err != nil {
			h.logger.Error("Failed to retrieve artists", slog.String("error", err.Error()))
			http.Error(w, "Failed to retrieve artists", http.StatusInternalServerError)
			return
		}

		h.renderer.Render(w, "artists", map[string]interface{}{
			"Title":   "Artists",
			"Artists": artists,
		})
	}
}

// POST /artists
func (h *Handler) CreateArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type CreateArtistRequest struct {
			Name string `form:"name" validate:"required,min=2,max=100"`
		}

		var req CreateArtistRequest
		if err := h.bind(r, &req); err != nil {
			// Check if it's a validation error
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(h.formatValidationErrorsHTML(validationErrs)))
				return
			}
			// Other binding errors
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		artist, err := h.queries.CreateArtist(r.Context(), req.Name)
		if err != nil {
			h.logger.Error("Failed to create artist", slog.String("error", err.Error()), slog.String("name", req.Name))
			http.Error(w, "Failed to create artist", http.StatusInternalServerError)
			return
		}

		h.logger.Info("Artist created", slog.Int64("artistID", artist.ID), slog.String("name", artist.Name))

		// Render the artist row partial
		err = h.renderer.Render(w, "artists-row", artist)
		if err != nil {
			h.logger.Error("Failed to render template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GET /artists/{id}
func (h *Handler) GetArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// is there an id?
		id := r.PathValue("id")
		if id == "" {
			h.logger.Info("Bad request. Missing path id")
			http.Error(w, "Missing parameter: id", http.StatusBadRequest)
			return
		}

		// is it an integer?
		artistID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			h.logger.Error("Invalid parameter id", slog.String("error", err.Error()))
			http.Error(w, "Invalid parameter: id", http.StatusBadRequest)
			return
		}

		// does the artist exist?
		artist, err := h.queries.GetArtist(r.Context(), artistID)
		if err != nil {
			h.logger.Error("Failed to retrieve artist", slog.String("error", err.Error()), slog.Int64("artistID", artistID))
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}

		// fetch artist's records
		records, err := h.queries.GetRecordsByArtist(r.Context(), sql.NullInt64{Int64: artistID, Valid: true})
		if err != nil {
			h.logger.Error("Failed to retrieve artist records", slog.String("error", err.Error()), slog.Int64("artistID", artistID))
			// Not fatal - just show empty records list
			records = []store.Record{}
		}

		// render artist detail page
		h.logger.Info("Artist retrieved", slog.Int64("artistID", artistID), slog.String("name", artist.Name), slog.Int("recordCount", len(records)))
		
		err = h.renderer.Render(w, "artist-detail", map[string]interface{}{
			"Title":   artist.Name,
			"Artist":  artist,
			"Records": records,
		})
		if err != nil {
			h.logger.Error("Failed to render template", slog.String("error", err.Error()))
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	}
}

// PUT /artists/{id}
func  (h *Handler) UpdateArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update artist
		// - Parse ID from path
		// - Parse form data
		// - Validate input
		// - Update artist in database
		// - Return updated artist row partial for HTMX
		id := r.PathValue("id")
		h.logger.Info("Update artist handler called", slog.String("id", id))
		h.renderer.Render(w, "artists-row", nil)
	}
}

// DELETE /artists/{id}
func (h *Handler) DeleteArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete artist
		// - Parse ID from path
		// - Check if artist has records (maybe prevent deletion or cascade)
		// - Delete artist from database
		// - Return 200 OK for HTMX to remove row
		id := r.PathValue("id")
		h.logger.Info("Delete artist handler called", slog.String("id", id))
		w.WriteHeader(http.StatusOK)
	}
}

// GET /artists/new
func (h *Handler) GetCreateArtistForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("GetCreateArtistForm()")
		// TODO: Render new artist form
		h.renderer.RenderPartial(w, "create-artist-form.html", nil)
	}
}

// GET /artists/{id}/edit
func (h *Handler) GetUpdateArtistForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render edit artist form
		// - Parse ID from path
		// - Query artist
		// - Render edit form with current data
		id := r.PathValue("id")
		h.logger.Info("Get artist edit form handler called", slog.String("id", id))
		h.renderer.Render(w, "update-artist-form", nil)
	}
}


// JSON API Endpoints

// GET /api/v1/artists
func (h *Handler) JsonGetArtists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get artists list
		// - Parse query params (page, limit, search)
		// - Query database with filters
		// - Return JSON array
		artists, err := h.queries.ListArtists(r.Context())
		if err != nil {
			h.logger.Error("Failed to retrieve artists", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Failed to retrieve artists", http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, artists, http.StatusOK)
	}
}

// POST /api/v1/artists
func (h *Handler) JsonCreateArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type CreateArtistRequest struct {
			Name string `json:"name" validate:"required,min=1,max=100"`
		}

		var req CreateArtistRequest
		if err := h.bind(r, &req); err != nil {
			// Check if it's a validation error
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ValidationErrorResponse{
					Error:   "Validation failed",
					Message: "Please check your input",
					Details: h.getValidationErrors(validationErrs),
				})
				return
			}
			// Other binding errors
			h.writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		artist, err := h.queries.CreateArtist(r.Context(), req.Name)
		if err != nil {
			h.logger.Error("Failed to create artist", slog.String("error", err.Error()), slog.String("name", req.Name))
			h.writeErrorJSON(w, "Failed to create artist", http.StatusInternalServerError)
			return
		}

		h.logger.Info("Artist created via API", slog.Int64("artistID", artist.ID), slog.String("name", artist.Name))

		h.writeJSON(w, artist, http.StatusCreated)
	}
}

// GET /api/v1/artists/{id}
func (h *Handler) JsonGetArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get single artist
		// - Parse ID from path
		// - Query artist from database
		// - Return JSON
		id := r.PathValue("id")
		artistID, _ := strconv.ParseInt(id, 10, 64)
		artist, err := h.queries.GetArtist(r.Context(), artistID)
		if err != nil {
			h.logger.Error("Failed to retrieve artist", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Artist not found", http.StatusNotFound)
			return
		}
		h.writeJSON(w, artist, http.StatusOK)
	}
}

// PUT /api/v1/artists/{id}
func (h *Handler) JsonUpdateArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update artist via API
		// - Parse ID from path
		// - Parse JSON body
		// - Validate input
		// - Update artist in database
		// - Return updated artist
		id := r.PathValue("id")
		h.logger.Info("API update artist handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "artist updated"}, http.StatusOK)
	}
}

// DELETE /api/v1/artists/{id}
func (h *Handler) JsonDeleteArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete artist via API
		// - Parse ID from path
		// - Delete artist from database
		// - Return 204 No Content
		id := r.PathValue("id")
		h.logger.Info("API delete artist handler called", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)
	}
}

// GET /api/v1/artists/{id}/records
func (h *Handler) JsonGetRecordsByArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all records by artist
		// - Parse ID from path
		// - Query records for artist
		// - Return JSON array
		id := r.PathValue("id")
		h.logger.Info("API get artist records handler called", slog.String("id", id))
		h.writeJSON(w, []interface{}{}, http.StatusOK)
	}
}
