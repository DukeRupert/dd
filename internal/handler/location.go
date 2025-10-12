package handler

import (
	"log/slog"
	"net/http"
)

// HTML Handlers

// GET /locations
func (h *Handler) GetLocations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all locations
		// - Query database
		// - Render locations list page
		locations, err := h.queries.ListLocations(r.Context())
		if err != nil {
			h.logger.Error("Failed to retrieve locations", slog.String("error", err.Error()))
			http.Error(w, "Failed to retrieve locations", http.StatusInternalServerError)
			return
		}

		h.renderer.Render(w, "locations", map[string]interface{}{
			"Title":     "Locations",
			"Locations": locations,
		})
	}
}

// GET /locations/new
func (h *Handler) GetCreateLocationForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render new location form
		h.renderer.Render(w, "create-location-form", nil)
	}
}

// POST /locations
func (h *Handler) CreateLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Create new location
		// - Parse form data
		// - Validate input
		// - Create location in database
		// - Return location row partial for HTMX
		h.logger.Info("Create location handler called")
		h.renderer.Render(w, "locations-row", nil)
	}
}

// GET /locations/{id}
func (h *Handler) GetLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get location details
		// - Parse ID from path
		// - Query location and records at that location
		// - Render location detail page
		id := r.PathValue("id")
		h.logger.Info("Get location handler called", slog.String("id", id))
		h.renderer.Render(w, "location-detail", nil)
	}
}

// GET /locations/{id}/edit
func (h *Handler) GetUpdateLocationForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render edit location form
		// - Parse ID from path
		// - Query location
		// - Render edit form with current data
		id := r.PathValue("id")
		h.logger.Info("Get location edit form handler called", slog.String("id", id))
		h.renderer.Render(w, "update-location-form", nil)
	}
}

// PUT /locations/{id}
func (h *Handler) UpdateLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update location
		// - Parse ID from path
		// - Parse form data
		// - Validate input
		// - Update location in database
		// - Return updated location row partial for HTMX
		id := r.PathValue("id")
		h.logger.Info("Update location handler called", slog.String("id", id))
		h.renderer.Render(w, "locations-row", nil)
	}
}

// DELETE /locations/{id}
func (h *Handler) DeleteLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete location
		// - Parse ID from path
		// - Check if location has records (maybe prevent deletion or set to null)
		// - Delete location from database
		// - Return 200 OK for HTMX to remove row
		id := r.PathValue("id")
		h.logger.Info("Delete location handler called", slog.String("id", id))
		w.WriteHeader(http.StatusOK)
	}
}

// POST /locations/default/{id}
func (h *Handler) SetDefaultLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Set location as default
		// - Parse ID from path
		// - Call SetDefaultLocation query (sets this location to default, others to false)
		// - Return updated locations list or row for HTMX
		id := r.PathValue("id")
		h.logger.Info("Set default location handler called", slog.String("id", id))
		h.renderer.Render(w, "locations-list", nil)
	}
}

// API Handlers

// GET /v1/locations
func (h *Handler) JsonGetLocations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get locations list
		// - Parse query params (page, limit, search)
		// - Query database with filters
		// - Return JSON array
		locations, err := h.queries.ListLocations(r.Context())
		if err != nil {
			h.logger.Error("Failed to retrieve locations", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Failed to retrieve locations", http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, locations, http.StatusOK)
	}
}

// POST /v1/locations
func (h *Handler) JsonCreateLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Create location via API
		// - Parse JSON body
		// - Validate input
		// - Create location in database
		// - Return created location with 201 status
		h.logger.Info("API create location handler called")
		h.writeJSON(w, map[string]string{"message": "location created"}, http.StatusCreated)
	}
}

// GET /v1/records/{id}
func (h *Handler) JsonGetLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get single location
		// - Parse ID from path
		// - Query location from database
		// - Return JSON
		id := r.PathValue("id")
		h.logger.Info("API get location handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "location details"}, http.StatusOK)
	}
}
// PUT /v1/locations/{id}
func (h *Handler) JsonUpdateLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update location via API
		// - Parse ID from path
		// - Parse JSON body
		// - Validate input
		// - Update location in database
		// - Return updated location
		id := r.PathValue("id")
		h.logger.Info("API update location handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "location updated"}, http.StatusOK)
	}
}

// DELETE /v1/locations/{id}
func (h *Handler) JsonDeleteLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete location via API
		// - Parse ID from path
		// - Delete location from database
		// - Return 204 No Content
		id := r.PathValue("id")
		h.logger.Info("API delete location handler called", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)
	}
}

// POST /v1/locations/default/{id}
func (h *Handler) JsonSetDefaultLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Set location as default via API
		// - Parse ID from path
		// - Call SetDefaultLocation query
		// - Return updated location
		id := r.PathValue("id")
		h.logger.Info("API set default location handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "default location set"}, http.StatusOK)
	}
}

// GET /v1/locations/{id}/records
func (h *Handler) JsonGetRecordsByLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all records at location
		// - Parse ID from path
		// - Query records at location
		// - Return JSON array
		id := r.PathValue("id")
		h.logger.Info("API get location records handler called", slog.String("id", id))
		h.writeJSON(w, []interface{}{}, http.StatusOK)
	}
}