package main

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/store"
)

// HTML Handler (Dashboard already exists in http.go, but keeping for reference)

func handleDashboard(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get dashboard statistics
		// - Count total artists
		// - Count total records
		// - Count total locations
		// - Get recently played records
		// - Get most played records
		// - Calculate estimated value (if tracking prices)
		// - Render dashboard page

		type DashboardStats struct {
			TotalArtists      int64
			TotalRecords      int64
			TotalLocations    int64
			RecentlyPlayed    []store.Record
			MostPlayed        []store.Record
			EstimatedValue    string
		}

		type DashboardData struct {
			Title string
			Stats DashboardStats
		}

		// Example of how to get counts
		artistCount, _ := queries.CountArtists(r.Context())
		recordCount, _ := queries.CountRecords(r.Context())
		locationCount, _ := queries.CountLocations(r.Context())
		
		data := DashboardData{
			Title: "Dashboard",
			Stats: DashboardStats{
				TotalArtists:   artistCount,
				TotalRecords:   recordCount,
				TotalLocations: locationCount,
				EstimatedValue: "$0.00",
			},
		}

		logger.Info("Dashboard handler called")
		err := renderer.Render(w, "home", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// API Handler

func handleAPIGetStats(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get collection statistics via API
		// - Count total artists
		// - Count total records
		// - Count total locations
		// - Count records by condition
		// - Count records by decade
		// - Get recently played count
		// - Calculate average play count
		// - Return JSON with all stats

		type CollectionStats struct {
			TotalArtists      int64            `json:"total_artists"`
			TotalRecords      int64            `json:"total_records"`
			TotalLocations    int64            `json:"total_locations"`
			RecordsByCondition map[string]int64 `json:"records_by_condition"`
			RecordsByDecade   map[string]int64 `json:"records_by_decade"`
			RecentlyPlayed    int              `json:"recently_played_count"`
			AveragePlayCount  float64          `json:"average_play_count"`
		}

		// Example counts
		artistCount, _ := queries.CountArtists(r.Context())
		recordCount, _ := queries.CountRecords(r.Context())
		locationCount, _ := queries.CountLocations(r.Context())

		stats := CollectionStats{
			TotalArtists:      artistCount,
			TotalRecords:      recordCount,
			TotalLocations:    locationCount,
			RecordsByCondition: map[string]int64{},
			RecordsByDecade:   map[string]int64{},
			RecentlyPlayed:    0,
			AveragePlayCount:  0.0,
		}

		logger.Info("API get stats handler called")
		writeJSON(w, stats, http.StatusOK)
	})
}