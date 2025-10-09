package main

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/store"
)

// HTML Handlers

func handleLanding(renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := renderer.Render(w, "landing", nil)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
	})
}

func handleSignupPage(renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := renderer.Render(w, "signup", nil)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
	})
}

func handleSignup(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Process signup form
		// - Parse form data
		// - Validate input
		// - Hash password
		// - Create user
		// - Create session
		// - Redirect to dashboard
		logger.Info("Signup handler called")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})
}

func handleLoginPage(renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        err := renderer.Render(w, "login", nil)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    })
}

func handleLogin(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Process login form
		// - Parse form data
		// - Validate credentials
		// - Create session
		// - Set cookie
		// - Redirect to dashboard or original URL
		logger.Info("Login handler called")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})
}

func handleLogout(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Process logout
		// - Get session token from cookie
		// - Delete session from database
		// - Clear cookie
		// - Redirect to landing page
		logger.Info("Logout handler called")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

// API Handlers

func handleAPISignup(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: API signup
		// - Parse JSON body
		// - Validate input
		// - Hash password
		// - Create user
		// - Return user data (without password)
		logger.Info("API signup handler called")
		writeJSON(w, map[string]string{"message": "signup endpoint"}, http.StatusCreated)
	})
}

func handleAPILogin(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: API login
		// - Parse JSON body
		// - Validate credentials
		// - Create API token
		// - Return token and user data
		logger.Info("API login handler called")
		writeJSON(w, map[string]string{"token": "fake-token"}, http.StatusOK)
	})
}

func handleAPILogout(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: API logout
		// - Get token from Authorization header
		// - Delete token from database
		// - Return success
		logger.Info("API logout handler called")
		writeJSON(w, map[string]string{"message": "logged out"}, http.StatusOK)
	})
}