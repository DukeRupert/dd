package main

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
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
		type SignupRequest struct {
			Email    string `form:"email" validate:"required,email"`
			Username string `form:"username" validate:"required,min=3,max=50"`
			Password string `form:"password" validate:"required,min=8"`
		}

		var req SignupRequest
		if err := bind(r, &req); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(formatValidationErrorsHTML(validationErrs)))
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if email already exists
		_, err := queries.GetUserByEmail(r.Context(), req.Email)
		if err == nil {
			// User already exists
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Failed to hash password", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Generate UUID for user
		userID := generateUUID()

		// Create user
		ctx := r.Context()
		_, err = queries.CreateUser(ctx, store.CreateUserParams{
			ID:           userID,
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Role:         "user",
		})
		if err != nil {
			logger.Error("Failed to create user", slog.String("error", err.Error()))
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}

		// Create session
		token, err := createSession(r.Context(), queries, userID, r)
		if err != nil {
			logger.Error("Failed to create session", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		setSessionCookie(w, token)

		logger.Info("User created successfully", slog.String("userID", userID), slog.String("email", req.Email))

		// Redirect to dashboard
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

		// Verify password using bcrypt
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			logger.Warn("Invalid password attempt", slog.String("email", req.Email))
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Create session
		token, err := createSession(r.Context(), queries, user.ID, r)
		if err != nil {
			logger.Error("Failed to create session", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		setSessionCookie(w, token)
		
		logger.Info("User logged in successfully", slog.String("userID", user.ID), slog.String("email", user.Email))
		
		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})
}

func handleLogout(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session token from cookie
		cookie, err := r.Cookie(SessionCookieName)
		if err == nil && cookie.Value != "" {
			// Delete session from database
			err = queries.DeleteSession(r.Context(), cookie.Value)
			if err != nil {
				logger.Error("Failed to delete session", slog.String("error", err.Error()))
			}
		}

		// Clear cookie
		clearSessionCookie(w)

		logger.Info("User logged out")
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

// Helpers

// generateUUID generates a simple UUID v4
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}