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
	mux.Handle("GET /", handleLanding(renderer))
	mux.Handle("GET /signup", handleSignupPage(renderer))
	mux.Handle("POST /signup", handleSignup(logger, queries, renderer))
	mux.Handle("GET /login", handleLoginPage(renderer))
	mux.Handle("POST /login", handleLogin(logger, queries, renderer))
	mux.Handle("GET /forgot-password", handleForgotPasswordPage(renderer))
	mux.Handle("POST /logout", handleLogout(logger, queries))

	// Protected routes - register directly with methods (no sub-mux needed for now)
	
	// Dashboard
	mux.Handle("GET /dashboard", handleDashboard(logger, queries, renderer))
	
	// Artists
	mux.Handle("GET /artists", handleGetArtistsPage(logger, queries, renderer))
	mux.Handle("GET /artists/new", handleGetArtistNewForm(renderer))
	mux.Handle("POST /artists", handlePostArtist(logger, queries, renderer))
	mux.Handle("GET /artists/{id}", handleGetArtist(logger, queries, renderer))
	mux.Handle("GET /artists/{id}/edit", handleGetArtistEditForm(logger, queries, renderer))
	mux.Handle("PUT /artists/{id}", handlePutArtist(logger, queries, renderer))
	mux.Handle("DELETE /artists/{id}", handleDeleteArtist(logger, queries))
	
	// Records
	mux.Handle("GET /records", handleGetRecordsPage(logger, queries, renderer))
	mux.Handle("GET /records/new", handleGetRecordNewForm(logger, queries, renderer))
	mux.Handle("POST /records", handlePostRecord(logger, queries, renderer))
	mux.Handle("GET /records/{id}", handleGetRecord(logger, queries, renderer))
	mux.Handle("GET /records/{id}/edit", handleGetRecordEditForm(logger, queries, renderer))
	mux.Handle("PUT /records/{id}", handlePutRecord(logger, queries, renderer))
	mux.Handle("DELETE /records/{id}", handleDeleteRecord(logger, queries))
	mux.Handle("POST /records/{id}/play", handlePostRecordPlay(logger, queries, renderer))
	
	// Locations
	mux.Handle("GET /locations", handleGetLocationsPage(logger, queries, renderer))
	mux.Handle("GET /locations/new", handleGetLocationNewForm(renderer))
	mux.Handle("POST /locations", handlePostLocation(logger, queries, renderer))
	mux.Handle("GET /locations/{id}", handleGetLocation(logger, queries, renderer))
	mux.Handle("GET /locations/{id}/edit", handleGetLocationEditForm(logger, queries, renderer))
	mux.Handle("PUT /locations/{id}", handlePutLocation(logger, queries, renderer))
	mux.Handle("DELETE /locations/{id}", handleDeleteLocation(logger, queries))
	mux.Handle("POST /locations/{id}/set-default", handlePostLocationSetDefault(logger, queries, renderer))
	
	// Profile/User
	mux.Handle("GET /profile", handleGetProfile(logger, queries, renderer))
	mux.Handle("GET /profile/edit", handleGetProfileEditForm(logger, queries, renderer))
	mux.Handle("PUT /profile", handlePutProfile(logger, queries, renderer))
	mux.Handle("GET /profile/password", handleGetPasswordForm(renderer))
	mux.Handle("PUT /profile/password", handlePutPassword(logger, queries))
}

func addAPIRoutes(mux *http.ServeMux, logger *slog.Logger, queries *store.Queries) {
	// Public API routes
	mux.Handle("POST /v1/auth/signup", handleAPISignup(logger, queries))
	mux.Handle("POST /v1/auth/login", handleAPILogin(logger, queries))
	mux.Handle("POST /v1/auth/logout", handleAPILogout(logger, queries))

	// Protected API routes
	protectedMux := http.NewServeMux()
	
	// Me/Current user
	protectedMux.Handle("GET /v1/me", handleAPIGetMe(logger, queries))
	
	// Artists
	protectedMux.Handle("GET /v1/artists", handleAPIGetArtists(logger, queries))
	protectedMux.Handle("POST /v1/artists", handleAPIPostArtist(logger, queries))
	protectedMux.Handle("GET /v1/artists/{id}", handleAPIGetArtist(logger, queries))
	protectedMux.Handle("PUT /v1/artists/{id}", handleAPIPutArtist(logger, queries))
	protectedMux.Handle("DELETE /v1/artists/{id}", handleAPIDeleteArtist(logger, queries))
	protectedMux.Handle("GET /v1/artists/{id}/records", handleAPIGetArtistRecords(logger, queries))
	
	// Records
	protectedMux.Handle("GET /v1/records", handleAPIGetRecords(logger, queries))
	protectedMux.Handle("POST /v1/records", handleAPIPostRecord(logger, queries))
	protectedMux.Handle("GET /v1/records/{id}", handleAPIGetRecord(logger, queries))
	protectedMux.Handle("PUT /v1/records/{id}", handleAPIPutRecord(logger, queries))
	protectedMux.Handle("DELETE /v1/records/{id}", handleAPIDeleteRecord(logger, queries))
	protectedMux.Handle("POST /v1/records/{id}/play", handleAPIPostRecordPlay(logger, queries))
	protectedMux.Handle("GET /v1/records/recent", handleAPIGetRecentRecords(logger, queries))
	protectedMux.Handle("GET /v1/records/popular", handleAPIGetPopularRecords(logger, queries))
	
	// Locations
	protectedMux.Handle("GET /v1/locations", handleAPIGetLocations(logger, queries))
	protectedMux.Handle("POST /v1/locations", handleAPIPostLocation(logger, queries))
	protectedMux.Handle("GET /v1/locations/{id}", handleAPIGetLocation(logger, queries))
	protectedMux.Handle("PUT /v1/locations/{id}", handleAPIPutLocation(logger, queries))
	protectedMux.Handle("DELETE /v1/locations/{id}", handleAPIDeleteLocation(logger, queries))
	protectedMux.Handle("POST /v1/locations/{id}/set-default", handleAPIPostLocationSetDefault(logger, queries))
	protectedMux.Handle("GET /v1/locations/{id}/records", handleAPIGetLocationRecords(logger, queries))
	
	// Users
	protectedMux.Handle("GET /v1/users/{id}", handleAPIGetUser(logger, queries))
	protectedMux.Handle("PUT /v1/users/{id}", handleAPIPutUser(logger, queries))
	protectedMux.Handle("DELETE /v1/users/{id}", handleAPIDeleteUser(logger, queries))
	protectedMux.Handle("PUT /v1/users/{id}/password", handleAPIPutUserPassword(logger, queries))
	
	// Stats
	protectedMux.Handle("GET /v1/stats", handleAPIGetStats(logger, queries))

	// Mount protected routes (no auth for now during development)
	// TODO: Uncomment this line when ready to enable authentication
	// mux.Handle("/v1/", RequireAPIAuthMiddleware(protectedMux))
	
	// For now, mount without auth for easy testing
	mux.Handle("/v1/", protectedMux)
}