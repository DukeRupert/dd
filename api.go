package main

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/store"
)

func handleGetArtists(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do the records exist?
		artists, err := queries.ListArtists(r.Context())
		if err != nil {
			logger.Error("Failed to retrieve artists from database", slog.String("error", err.Error()))
			http.Error(w, "Artists not found", http.StatusInternalServerError)
			return
		}
		// great success!
		logger.Info("List artists", slog.Int("artists", len(artists)))
		writeJSON(w, artists, http.StatusOK)
	})
}
