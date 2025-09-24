package main

import (
	"net/http"

	"github.com/dukerupert/dd/internal/store"
	"github.com/rs/zerolog"
)

func addRoutes(mux *http.ServeMux, l zerolog.Logger, q *store.Queries, r *TemplateRenderer) {
	// Register routes
	mux.Handle("GET /", homeHandler(r))
	mux.Handle("GET /artists", getArtistsHandler(l, q, r))
	mux.Handle("POST /artists", postArtistHandler(l, q))
	mux.Handle("GET /artists/add-form", getCreateArtistForm(r))
	mux.Handle("GET /artists/{id}", getArtistHandler(l, q))
	mux.Handle("PUT /artists/{id}", putArtistHandler(l, q))
	mux.Handle("DELETE /artists/{id}", deleteArtistHandler(l, q))
	mux.Handle("GET /artists/{id}/update-form", getUpdateArtistForm(l, q, r))
	mux.Handle("GET /locations", getLocationsHandler(l, q, r))
	mux.Handle("GET /albums", getRecordsHandler(l, q, r))
	mux.HandleFunc("GET /static", staticHandler)
}