package main

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/store"
)

// HTML Handlers

func handleGetRecordsPage(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get all records
		// - Query database with joins for artist/location names
		// - Handle search and filter query params
		// - Render records list page
		records, err := queries.ListRecordsWithDetails(r.Context())
		if err != nil {
			logger.Error("Failed to retrieve records", slog.String("error", err.Error()))
			http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
			return
		}

		renderer.Render(w, "albums", map[string]interface{}{
			"Title":   "Records",
			"Records": records,
		})
	})
}

func handleGetRecordNewForm(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render new record form
		// - Query artists for dropdown
		// - Query locations for dropdown
		// - Render form
		logger.Info("Get new record form handler called")
		renderer.Render(w, "create-record-form", nil)
	})
}

func handlePostRecord(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Create new record
		// - Parse form data
		// - Validate input
		// - Create record in database
		// - Return record row partial for HTMX
		logger.Info("Create record handler called")
		renderer.Render(w, "records-row", nil)
	})
}

func handleGetRecord(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get record details
		// - Parse ID from path
		// - Query record with artist/location details
		// - Render record detail page
		id := r.PathValue("id")
		logger.Info("Get record handler called", slog.String("id", id))
		renderer.Render(w, "record-detail", nil)
	})
}

func handleGetRecordEditForm(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render edit record form
		// - Parse ID from path
		// - Query record
		// - Query artists for dropdown
		// - Query locations for dropdown
		// - Render edit form with current data
		id := r.PathValue("id")
		logger.Info("Get record edit form handler called", slog.String("id", id))
		renderer.Render(w, "update-record-form", nil)
	})
}

func handlePutRecord(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update record
		// - Parse ID from path
		// - Parse form data
		// - Validate input
		// - Update record in database
		// - Return updated record row partial for HTMX
		id := r.PathValue("id")
		logger.Info("Update record handler called", slog.String("id", id))
		renderer.Render(w, "records-row", nil)
	})
}

func handleDeleteRecord(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete record
		// - Parse ID from path
		// - Delete record from database
		// - Return 200 OK for HTMX to remove row
		id := r.PathValue("id")
		logger.Info("Delete record handler called", slog.String("id", id))
		w.WriteHeader(http.StatusOK)
	})
}

func handlePostRecordPlay(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Increment play count
		// - Parse ID from path
		// - Call RecordPlayback query (updates play_count and last_played_at)
		// - Return updated record row or play count display for HTMX
		id := r.PathValue("id")
		logger.Info("Record play handler called", slog.String("id", id))
		renderer.Render(w, "record-play-count", nil)
	})
}

// API Handlers

func handleAPIGetRecords(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get records list
		// - Parse query params (page, limit, search, artist_id, location_id, condition, year, sort)
		// - Query database with filters and joins
		// - Return JSON array with pagination metadata
		logger.Info("API get records handler called")
		records, err := queries.ListRecordsWithDetails(r.Context())
		if err != nil {
			logger.Error("Failed to retrieve records", slog.String("error", err.Error()))
			writeErrorJSON(w, "Failed to retrieve records", http.StatusInternalServerError)
			return
		}
		writeJSON(w, records, http.StatusOK)
	})
}

func handleAPIPostRecord(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Create record via API
		// - Parse JSON body
		// - Validate input
		// - Create record in database
		// - Return created record with 201 status
		logger.Info("API create record handler called")
		writeJSON(w, map[string]string{"message": "record created"}, http.StatusCreated)
	})
}

func handleAPIGetRecord(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get single record
		// - Parse ID from path
		// - Query record with details from database
		// - Return JSON
		id := r.PathValue("id")
		logger.Info("API get record handler called", slog.String("id", id))
		writeJSON(w, map[string]string{"message": "record details"}, http.StatusOK)
	})
}

func handleAPIPutRecord(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update record via API
		// - Parse ID from path
		// - Parse JSON body
		// - Validate input
		// - Update record in database
		// - Return updated record
		id := r.PathValue("id")
		logger.Info("API update record handler called", slog.String("id", id))
		writeJSON(w, map[string]string{"message": "record updated"}, http.StatusOK)
	})
}

func handleAPIDeleteRecord(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete record via API
		// - Parse ID from path
		// - Delete record from database
		// - Return 204 No Content
		id := r.PathValue("id")
		logger.Info("API delete record handler called", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)
	})
}

func handleAPIPostRecordPlay(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Increment play count via API
		// - Parse ID from path
		// - Call RecordPlayback query
		// - Return updated record
		id := r.PathValue("id")
		logger.Info("API record play handler called", slog.String("id", id))
		writeJSON(w, map[string]string{"message": "play recorded"}, http.StatusOK)
	})
}

func handleAPIGetRecentRecords(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get recently played records
		// - Parse limit query param (default 10)
		// - Call GetRecentlyPlayedRecords query
		// - Return JSON array
		logger.Info("API get recent records handler called")
		writeJSON(w, []interface{}{}, http.StatusOK)
	})
}

func handleAPIGetPopularRecords(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get most played records
		// - Parse limit query param (default 10)
		// - Call GetMostPlayedRecords query
		// - Return JSON array
		logger.Info("API get popular records handler called")
		writeJSON(w, []interface{}{}, http.StatusOK)
	})
}