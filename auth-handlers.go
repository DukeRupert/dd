package main

import (
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
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
		type LoginRequest struct {
			Email    string `form:"email" validate:"required,email"`
			Password string `form:"password" validate:"required"`
		}

		var req LoginRequest
		if err := bind(r, &req); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(formatValidationErrorsHTML(validationErrs)))
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get user by email
		user, err := queries.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			logger.Warn("Login attempt for non-existent user", slog.String("email", req.Email))
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// For now, just check if password matches the hash
		// TODO: Implement proper password hashing with bcrypt
		if user.PasswordHash != req.Password {
			logger.Warn("Invalid password attempt", slog.String("email", req.Email))
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// TODO: Create session
		// TODO: Set session cookie
		
		logger.Info("User logged in successfully", slog.String("userID", user.ID), slog.String("email", user.Email))
		
		// Redirect to dashboard
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

func handleForgotPasswordPage(t *TemplateRenderer) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        err := t.Render(w, "forgot-password", nil)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
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