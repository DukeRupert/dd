package handler

import (
	"log/slog"
	"net/http"
)

type CreateRecordRequest struct {
	Title             string `form:"title" json:"title" validate:"required,min=1,max=200"`
	ArtistID          int64  `form:"artist_id" json:"artist_id" validate:"omitempty,min=1"`
	AlbumTitle        string `form:"album_title" json:"album_title" validate:"max=200"`
	ReleaseYear       int32  `form:"release_year" json:"release_year" validate:"omitempty,min=1900,max=2100"`
	CurrentLocationID int64  `form:"current_location_id" json:"current_location_id" validate:"omitempty,min=1"`
	HomeLocationID    int64  `form:"home_location_id" json:"home_location_id" validate:"omitempty,min=1"`
	CatalogNumber     string `form:"catalog_number" json:"catalog_number" validate:"max=100"`
	Condition         string `form:"condition" json:"condition" validate:"omitempty,oneof=Mint 'Near Mint' 'Very Good' Good Fair Poor"`
	Notes             string `form:"notes" json:"notes" validate:"max=1000"`
}

type UpdateRecordRequest struct {
	Title             string `form:"title" json:"title" validate:"required,min=1,max=200"`
	ArtistID          int64 `form:"artist_id" json:"artist_id" validate:"omitempty,min=1"`
	AlbumTitle        string `form:"album_title" json:"album_title" validate:"max=200"`
	ReleaseYear       int32  `form:"release_year" json:"release_year" validate:"omitempty,min=1900,max=2100"`
	CurrentLocationID int64  `form:"current_location_id" json:"current_location_id" validate:"omitempty,min=1"`
	HomeLocationID    int64  `form:"home_location_id" json:"home_location_id" validate:"omitempty,min=1"`
	CatalogNumber     string `form:"catalog_number" json:"catalog_number" validate:"max=100"`
	Condition         string `form:"condition" json:"condition" validate:"omitempty,oneof=Mint 'Near Mint' 'Very Good' Good Fair Poor"`
	Notes             string `form:"notes" json:"notes" validate:"max=1000"`
}

type UpdateRecordLocationRequest struct {
	CurrentLocationID int64 `form:"current_location_id" json:"current_location_id" validate:"required,min=1"`
}

type UpdateRecordConditionRequest struct {
	Condition string `form:"condition" json:"condition" validate:"required,oneof=Mint 'Near Mint' 'Very Good' Good Fair Poor"`
}

// HTML Handlers

// GET /records
func (h *Handler) GetRecords() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all records
		// - Query database with joins for artist/location names
		// - Handle search and filter query params
		// - Render records list page
		records, err := h.queries.ListRecordsWithDetails(r.Context())
		if err != nil {
			h.logger.Error("Failed to retrieve records", slog.String("error", err.Error()))
			http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
			return
		}

		h.renderer.Render(w, "albums", map[string]interface{}{
			"Title":   "Records",
			"Records": records,
		})
	}
}

// POST /records
func (h *Handler) CreateRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Create new record
		// - Parse form data
		// - Validate input
		// - Create record in database
		// - Return record row partial for HTMX
		h.logger.Info("Create record handler called")
		h.renderer.RenderPartial(w, "records-row", nil)
	}
}

// GET /records/{id}
func (h *Handler) GetRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get record details
		// - Parse ID from path
		// - Query record with artist/location details
		// - Render record detail page
		id := r.PathValue("id")
		h.logger.Info("Get record handler called", slog.String("id", id))
		h.renderer.Render(w, "record-detail", nil)
	}
}

// GET /records/new
func (h *Handler) GetCreateRecordForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render new record form
		// - Query artists for dropdown
		// - Query locations for dropdown
		// - Render form
		h.logger.Info("Get new record form handler called")
		h.renderer.RenderPartial(w, "create-record-form", nil)
	}
}

// GET /records/{id}/edit
func (h *Handler) GetUpdateRecordForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render edit record form
		// - Parse ID from path
		// - Query record
		// - Query artists for dropdown
		// - Query locations for dropdown
		// - Render edit form with current data
		id := r.PathValue("id")
		h.logger.Info("Get record edit form handler called", slog.String("id", id))
		h.renderer.RenderPartial(w, "update-record-form", nil)
	}
}

// PUT /records/{id}
func (h *Handler) UpdateRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update record
		// - Parse ID from path
		// - Parse form data
		// - Validate input
		// Convert to database params
		// var artistID sql.NullInt64
		// if req.ArtistID != nil {
		// 	artistID = sql.NullInt64{Int64: *req.ArtistID, Valid: true}
		// } else {
		// 	artistID = sql.NullInt64{Valid: false} // NULL
		// }
		// - Update record in database
		// - Return updated record row partial for HTMX
		id := r.PathValue("id")
		h.logger.Info("Update record handler called", slog.String("id", id))
		h.renderer.RenderPartial(w, "records-row", nil)
	}
}

// DELETE /records/{id}
func (h *Handler) DeleteRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete record
		// - Parse ID from path
		// - Delete record from database
		// - Return 200 OK for HTMX to remove row
		id := r.PathValue("id")
		h.logger.Info("Delete record handler called", slog.String("id", id))
		w.WriteHeader(http.StatusOK)
	}
}

// POST /records/{id}/play
func (h *Handler) PlayRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Increment play count
		// - Parse ID from path
		// - Call RecordPlayback query (updates play_count and last_played_at)
		// - Return updated record row or play count display for HTMX
		id := r.PathValue("id")
		h.logger.Info("Record play handler called", slog.String("id", id))
		h.renderer.RenderPartial(w, "record-play-count", nil)
	}
}

// API Handlers

// GET /api/v1/records
func (h *Handler) JsonGetRecords() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get records list
		// - Parse query params (page, limit, search, artist_id, location_id, condition, year, sort)
		// - Query database with filters and joins
		// - Return JSON array with pagination metadata
		h.logger.Info("API get records handler called")
		records, err := h.queries.ListRecordsWithDetails(r.Context())
		if err != nil {
			h.logger.Error("Failed to retrieve records", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Failed to retrieve records", http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, records, http.StatusOK)
	}
}

// POST /api/v1/records
func (h *Handler) JsonCreateRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Create record via API
		// - Parse JSON body
		// - Validate input
		// - Create record in database
		// - Return created record with 201 status
		h.logger.Info("API create record handler called")
		h.writeJSON(w, map[string]string{"message": "record created"}, http.StatusCreated)
	}
}

// GET /api/v1/records/{id}
func (h *Handler) JsonGetRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get single record
		// - Parse ID from path
		// - Query record with details from database
		// - Return JSON
		id := r.PathValue("id")
		h.logger.Info("API get record handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "record details"}, http.StatusOK)
	}
}

// PUT /api/v1/records/{id}
func (h *Handler) JsonUpdateRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update record via API
		// - Parse ID from path
		// - Parse JSON body
		// - Validate input
		// - Update record in database
		// - Return updated record
		id := r.PathValue("id")
		h.logger.Info("API update record handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "record updated"}, http.StatusOK)
	}
}

// DELETE /api/v1/records/{id}
func (h *Handler) JsonDeleteRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete record via API
		// - Parse ID from path
		// - Delete record from database
		// - Return 204 No Content
		id := r.PathValue("id")
		h.logger.Info("API delete record handler called", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)
	}
}

// POST /api/v1/records/{api}/play
func (h *Handler) JsonPlayRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Increment play count via API
		// - Parse ID from path
		// - Call RecordPlayback query
		// - Return updated record
		id := r.PathValue("id")
		h.logger.Info("API record play handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "play recorded"}, http.StatusOK)
	}
}

// GET /api/v1/records/recent
func (h *Handler) JsonGetRecordsByRecent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get recently played records
		// - Parse limit query param (default 10)
		// - Call GetRecentlyPlayedRecords query
		// - Return JSON array
		h.logger.Info("API get recent records handler called")
		h.writeJSON(w, []interface{}{}, http.StatusOK)
	}
}

// GET /api/v1/records/popular
func (h *Handler) JsonGetRecordsByPopular() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get most played records
		// - Parse limit query param (default 10)
		// - Call GetMostPlayedRecords query
		// - Return JSON array
		h.logger.Info("API get popular records handler called")
		h.writeJSON(w, []interface{}{}, http.StatusOK)
	}
}
