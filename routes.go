package main

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/store"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) {
	// API routes - stricter rate limiting, smaller body size, no template renderer
	apiMux := http.NewServeMux()
	addAPIRoutes(apiMux, logger, queries)

	// Chain API middleware (innermost to outermost)
	apiHandler := http.StripPrefix("/api", apiMux)
	apiHandler = RateLimitMiddleware(apiHandler, 100)
	apiHandler = MaxBytesMiddleware(1 << 20)(apiHandler)
	apiHandler = AuthMiddleware(queries)(apiHandler)
	apiHandler = LoggingMiddleware(apiHandler, logger)
	apiHandler = RequestIDMiddleware(apiHandler)
	mux.Handle("/api/", apiHandler)

	// HTML routes - more relaxed limits, larger body size for file uploads
	htmlMux := http.NewServeMux()
	addHTMLRoutes(htmlMux, logger, queries, renderer)

	// Chain HTML middleware (innermost to outermost)
	htmlHandler := http.Handler(htmlMux)
	htmlHandler = RateLimitMiddleware(htmlHandler, 1000)
	htmlHandler = MaxBytesMiddleware(10 << 20)(htmlHandler)
	htmlHandler = AuthMiddleware(queries)(htmlHandler)
	htmlHandler = LoggingMiddleware(htmlHandler, logger)
	htmlHandler = RequestIDMiddleware(htmlHandler)
	mux.Handle("/", htmlHandler)
}

func addHTMLRoutes(mux *http.ServeMux, logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) {
	// Public routes
	// mux.HandleFunc("GET /login", handleLogin(renderer))
	// mux.Handle("GET /", handleHome(renderer))

	// Protected routes - require authentication
	protectedMux := http.NewServeMux()
	protectedMux.Handle("GET /dashboard", handleDashboard(renderer))
	// protectedMux.HandleFunc("POST /artists", handleCreateArtist(queries, renderer))

	mux.Handle("/dashboard", RequireAuthMiddleware(protectedMux))
	
	// Artists route with CSRF protection
	artistsHandler := http.Handler(protectedMux)
	artistsHandler = CSRFMiddleware(artistsHandler)
	artistsHandler = RequireAuthMiddleware(artistsHandler)
	mux.Handle("/artists", artistsHandler)

	// Admin only routes
	// adminMux := http.NewServeMux()
	// adminMux.HandleFunc("GET /admin", handleAdmin(queries, renderer))
	
	// Admin route with role check
	// adminHandler := http.Handler(adminMux)
	// adminHandler = RequireRoleMiddleware(queries, "admin")(adminHandler)
	// adminHandler = RequireAuthMiddleware(adminHandler)
	// mux.Handle("/admin/", adminHandler)
}

// func addHTMLRoutes(mux *http.ServeMux, logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) {
// 	mux.Handle("GET /", homeHandler(renderer))
// 	mux.Handle("GET /artists", handleGetArtistsPage(logger, queries, renderer))
// 	mux.Handle("POST /artists", postArtistHandler(logger, queries, renderer))
// 	mux.Handle("GET /artists/add-form", getCreateArtistForm(renderer))
// 	mux.Handle("GET /artists/{id}", getArtistHandler(logger, queries))
// 	mux.Handle("PUT /artists/{id}", putArtistHandler(logger, queries))
// 	mux.Handle("DELETE /artists/{id}", deleteArtistHandler(logger, queries))
// 	mux.Handle("GET /artists/{id}/update-form", getUpdateArtistForm(logger, queries, renderer))
// 	mux.Handle("GET /locations", getLocationsHandler(logger, queries, renderer))
// 	mux.Handle("GET /albums", getRecordsHandler(logger, queries, renderer))
// 	mux.HandleFunc("GET /static", staticHandler)
// }

func addAPIRoutes(mux *http.ServeMux, logger *slog.Logger, queries *store.Queries) {
	mux.Handle("GET /artists", handleGetArtists(logger, queries))
}
