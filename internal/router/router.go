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
	mux.HandleFunc("GET /dashboard", h.Dashboard())

	// Artists
	mux.HandleFunc("GET /artists", h.GetArtistsPage())
	mux.HandleFunc("GET /artists/new", h.GetArtistNewForm())
	mux.HandleFunc("POST /artists", h.PostArtist())
	mux.HandleFunc("GET /artists/{id}", h.GetArtist())
	mux.HandleFunc("GET /artists/{id}/edit", h.GetArtistEditForm())
	mux.HandleFunc("PUT /artists/{id}", h.PutArtist())
	mux.HandleFunc("DELETE /artists/{id}", h.DeleteArtist())

	// Records
	mux.HandleFunc("GET /records", h.GetRecordsPage())
	// ... etc
}

func addAPIRoutes(mux *http.ServeMux, h *handler.Handler) {
	// Public API
	mux.HandleFunc("POST /v1/auth/signup", h.APISignup())
	mux.HandleFunc("POST /v1/auth/login", h.APILogin())

	// Protected API
	mux.HandleFunc("GET /v1/artists", h.APIGetArtists())
	// ... etc
}