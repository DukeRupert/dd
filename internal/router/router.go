package router

import (
	"net/http"

	"github.com/dukerupert/dd/internal/handler"
	"github.com/dukerupert/dd/internal/middleware"
	"github.com/dukerupert/dd/internal/store"
)

// New creates and configures the application router
func New(h *handler.Handler, queries *store.Queries, sessionCookieName string) http.Handler {
	mux := http.NewServeMux()

	// API routes
	apiMux := http.NewServeMux()
	addAPIRoutes(apiMux, h)

	apiHandler := http.StripPrefix("/api", apiMux)
	apiHandler = middleware.RateLimit(apiHandler, 100)
	apiHandler = middleware.MaxBytes(1 << 20)(apiHandler)
	apiHandler = middleware.Auth(queries, sessionCookieName)(apiHandler)
	apiHandler = middleware.Logging(apiHandler, h.Logger())
	apiHandler = middleware.RequestID(apiHandler)
	mux.Handle("/api/", apiHandler)

	// HTML routes
	htmlMux := http.NewServeMux()
	addHTMLRoutes(htmlMux, h)

	htmlHandler := http.Handler(htmlMux)
	htmlHandler = middleware.RateLimit(htmlHandler, 1000)
	htmlHandler = middleware.MaxBytes(10 << 20)(htmlHandler)
	htmlHandler = middleware.Auth(queries, sessionCookieName)(htmlHandler)
	htmlHandler = middleware.Logging(htmlHandler, h.Logger())
	htmlHandler = middleware.RequestID(htmlHandler)
	mux.Handle("/", htmlHandler)

	return mux
}

func addHTMLRoutes(mux *http.ServeMux, h *handler.Handler) {
	// Public routes
	mux.HandleFunc("GET /", h.Landing())
	mux.HandleFunc("GET /signup", h.SignupPage())
	mux.HandleFunc("POST /signup", h.Signup())
	mux.HandleFunc("GET /login", h.LoginPage())
	mux.HandleFunc("POST /login", h.Login())
	mux.HandleFunc("POST /logout", h.Logout())
	mux.HandleFunc("GET /forgot-password", h.ForgotPassword())

	// Protected routes
	// TODO: Protect these routes in production

	// Artists
	mux.HandleFunc("GET /artists", h.GetArtists())
	mux.HandleFunc("GET /artists/new", h.GetCreateArtistForm())
	mux.HandleFunc("POST /artists", h.CreateArtist())
	mux.HandleFunc("GET /artists/{id}", h.GetArtist())
	mux.HandleFunc("PUT /artists/{id}", h.UpdateArtist())
	mux.HandleFunc("GET /artists/{id}/edit", h.GetUpdateArtistForm())
	mux.HandleFunc("DELETE /artists/{id}", h.DeleteArtist())

	// Records
	mux.HandleFunc("GET /records", h.GetRecords())
	mux.HandleFunc("GET /records/new", h.GetCreateRecordForm())
	mux.HandleFunc("POST /records", h.CreateRecord())
	mux.HandleFunc("GET /records/{id}", h.GetRecord())
	mux.HandleFunc("PUT /records/{id}", h.UpdateRecord())
	mux.HandleFunc("GET /records/{id}/edit", h.GetUpdateRecordForm())
	mux.HandleFunc("DELETE /records/{id}", h.DeleteRecord())
	mux.HandleFunc("POST /records/{id}/play", h.PlayRecord())

	// Locations
	mux.HandleFunc("GET /locations", h.GetLocations())
	mux.HandleFunc("GET /locations/new", h.GetCreateLocationForm())
	mux.HandleFunc("POST /locations", h.CreateLocation())
	mux.HandleFunc("GET /locations/{id}", h.GetLocation())
	mux.HandleFunc("PUT /locations/{id}", h.UpdateLocation())
	mux.HandleFunc("GET /locations/{id}/edit", h.GetUpdateLocationForm())
	mux.HandleFunc("DELETE /locations/{id}", h.DeleteLocation())
	mux.HandleFunc("POST /locations/default/{id}", h.SetDefaultLocation())

	// Profile
	mux.HandleFunc("GET /profile", h.GetProfile())
	mux.HandleFunc("PUT /profile", h.UpdateProfile())
	mux.HandleFunc("GET /profile/new", h.GetCreateArtistForm())
	mux.HandleFunc("GET /profile/edit", h.GetUpdateProfileForm())
	mux.HandleFunc("GET /profile/password", h.GetUpdatePasswordForm())
	mux.HandleFunc("PUT /profile/password", h.UpdatePassword())
}

func addAPIRoutes(mux *http.ServeMux, h *handler.Handler) {
	// Public API
	mux.HandleFunc("POST /v1/auth/signup", h.JsonSignup())
	mux.HandleFunc("POST /v1/auth/login", h.JsonLogin())
	mux.HandleFunc("POST /v1/auth/logout", h.JsonLogout())

	// Protected Api

	// Artists
	mux.HandleFunc("GET /v1/artists", h.JsonGetArtists())
	mux.HandleFunc("POST /v1/artists", h.JsonCreateArtist())
	mux.HandleFunc("GET /v1/artists/{id}", h.JsonGetArtist())
	mux.HandleFunc("PUT /v1/artists/{id}", h.JsonUpdateArtist())
	mux.HandleFunc("DELETE /v1/artists/{id}", h.JsonDeleteArtist())
	mux.HandleFunc("GET /v1/artists/{id}/records", h.JsonGetRecordsByArtist())

	// Records
	mux.HandleFunc("GET /v1/records", h.JsonGetRecords())
	mux.HandleFunc("POST /v1/records", h.JsonCreateRecord())
	mux.HandleFunc("GET /v1/records/{id}", h.JsonGetRecord())
	mux.HandleFunc("DELETE /v1/records/{id}", h.JsonDeleteRecord())
	mux.HandleFunc("PUT /v1/records/{id}", h.JsonUpdateRecord())
	mux.HandleFunc("GET /v1/records/{id}/play", h.JsonPlayRecord())
	mux.HandleFunc("GET /v1/records/recent", h.JsonGetRecordsByRecent())
	mux.HandleFunc("GET /v1/records/popular", h.JsonGetRecordsByPopular())

	// Locations
	mux.HandleFunc("GET /v1/locations", h.JsonGetLocations())
	mux.HandleFunc("POST /v1/locations", h.JsonCreateLocation())
	mux.HandleFunc("GET /v1/locations/{id}", h.JsonGetLocation())
	mux.HandleFunc("PUT /v1/locations/{id}", h.JsonUpdateLocation())
	mux.HandleFunc("DELETE /v1/locations/{id}", h.JsonDeleteLocation())
	mux.HandleFunc("GET /v1/locations/{id}/records", h.JsonGetRecordsByLocation())
	mux.HandleFunc("POST /v1/locations/default/{id}", h.JsonSetDefaultLocation())

	// User
	mux.HandleFunc("GET /v1/profile", h.JsonGetProfile())
	mux.HandleFunc("PUT /v1/profile", h.JsonUpdateProfile())
	mux.HandleFunc("PUT /v1/profile/password", h.JsonUpdatePassword())
}
