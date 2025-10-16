package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dukerupert/dd/internal/auth"
	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
)

const (
	SessionCookieName = "session_token"
	SessionDuration   = 24 * time.Hour * 7 // 7 days
)

type UserInfo struct {
	ID       string `form:"id" json:"id"`
	Email    string `form:"email" json:"email"`
	Username string `form:"username" json:"username"`
	Role     string `form:"role" json:"role"`
}

type LoginRequest struct {
	Email    string `form:"email" json:"email" validate:"required,email"`
	Password string `form:"password" json:"password" validate:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

type SignupRequest struct {
	Email    string `form:"email" validate:"required,email"`
	Username string `form:"username" validate:"required,min=3,max=50"`
	Password string `form:"password" validate:"required,min=8"`
}

type SignupResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// HTML Handlers

// GET /
func (h *Handler) Landing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.renderer.Render(w, "home", nil)
	}
}

// GET /signup
func (h *Handler) SignupPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.renderer.Render(w, "signup", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// POST /signup
func (h *Handler) Signup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignupRequest
		if err := h.bind(r, &req); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(h.formatValidationErrorsHTML(validationErrs)))
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if email already exists
		_, err := h.queries.GetUserByEmail(r.Context(), req.Email)
		if err == nil {
			// User already exists
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}

		// Hash password
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			h.logger.Error("Failed to hash password", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Generate UUID for user
		userID := generateUUID()

		// Create user
		ctx := r.Context()
		_, err = h.queries.CreateUser(ctx, store.CreateUserParams{
			ID:           userID,
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Role:         "user",
		})
		if err != nil {
			h.logger.Error("Failed to create user", slog.String("error", err.Error()))
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}

		// Create session, default length 2 hours
		token, err := auth.CreateSession(r.Context(), h.queries, userID, r, SessionDuration)
		if err != nil {
			h.logger.Error("Failed to create session", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		secure := h.config.Server.Env == "prod" // Set to true in production with HTTPS
		auth.SetSessionCookie(w, token, SessionCookieName, SessionDuration, secure)

		h.logger.Info("User created successfully", slog.String("userID", userID), slog.String("email", req.Email))

		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// GET /login
func (h *Handler) LoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.renderer.Render(w, "login", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// POST /login
func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := h.bind(r, &req); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(h.formatValidationErrorsHTML(validationErrs)))
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get user by email
		user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			h.logger.Warn("Login attempt for non-existent user", slog.String("email", req.Email))
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Verify password
		err = auth.ComparePassword(user.PasswordHash, req.Password)
		if err != nil {
			h.logger.Warn("Invalid password attempt", slog.String("email", req.Email))
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Create session
		token, err := auth.CreateSession(r.Context(), h.queries, user.ID, r, SessionDuration)
		if err != nil {
			h.logger.Error("Failed to create session", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		secure := h.config.Server.Env == "prod" // Set to true in production with HTTPS
		auth.SetSessionCookie(w, token, SessionCookieName, SessionDuration, secure)

		h.logger.Info("User logged in successfully", slog.String("userID", user.ID), slog.String("email", user.Email))

		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

//POST /logout
func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token from cookie
		cookie, err := r.Cookie(SessionCookieName)
		if err == nil && cookie.Value != "" {
			// Delete session from database
			err = h.queries.DeleteSession(r.Context(), cookie.Value)
			if err != nil {
				h.logger.Error("Failed to delete session", slog.String("error", err.Error()))
			}
		}

		// Clear cookie
		secure := h.config.Server.Env == "prod" // Set to true in production with HTTPS
		auth.ClearSessionCookie(w, SessionCookieName, secure)

		h.logger.Info("User logged out")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// GET /forgot-password
func (h *Handler) ForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.renderer.Render(w, "forgot-password", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// API Handlers

// POST /v1/auth/signup
func (h *Handler) JsonSignup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req SignupRequest
		if err := h.bind(r, &req); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ValidationErrorResponse{
					Error:   "Validation failed",
					Message: "Please check your input",
					Details: h.getValidationErrors(validationErrs),
				})
				return
			}
			h.writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if email already exists
		_, err := h.queries.GetUserByEmail(r.Context(), req.Email)
		if err == nil {
			// User already exists
			h.writeErrorJSON(w, "Email already registered", http.StatusConflict)
			return
		}

		// Hash password
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			h.logger.Error("Failed to hash password", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Generate UUID for user
		userID := generateUUID()

		// Create user
		ctx := r.Context()
		user, err := h.queries.CreateUser(ctx, store.CreateUserParams{
			ID:           userID,
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Role:         "user",
		})
		if err != nil {
			h.logger.Error("Failed to create user", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Failed to create account", http.StatusInternalServerError)
			return
		}

		h.logger.Info("User created via API", slog.String("userID", userID), slog.String("email", req.Email))

		// Generate JWT token
		token, err := auth.GenerateJWT(user.ID, user.Email, user.Role, h.config.Auth.JWTSecret, h.config.Auth.JWTExpiration)
		if err != nil {
			h.logger.Error("Failed to generate JWT", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return token and user info
		response := SignupResponse{
			Token:     token,
			ExpiresAt: time.Now().Add(h.config.Auth.JWTExpiration),
			User: UserInfo{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
				Role:     user.Role,
			},
		}

		h.writeJSON(w, response, http.StatusCreated)
	}
}

// POST /v1/auth/login
func (h *Handler) JsonLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := h.bind(r, &req); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ValidationErrorResponse{
					Error:   "Validation failed",
					Message: "Please check your input",
					Details: h.getValidationErrors(validationErrs),
				})
				return
			}
			h.writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get user by email
		user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			h.logger.Warn("API login attempt for non-existent user", slog.String("email", req.Email))
			h.writeErrorJSON(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Verify password
		err = auth.ComparePassword(user.PasswordHash, req.Password)
		if err != nil {
			h.logger.Warn("API invalid password attempt", slog.String("email", req.Email))
			h.writeErrorJSON(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Generate JWT token
		token, err := auth.GenerateJWT(user.ID, user.Email, user.Role, h.config.Auth.JWTSecret, h.config.Auth.JWTExpiration)
		if err != nil {
			h.logger.Error("Failed to generate JWT", slog.String("error", err.Error()))
			h.writeErrorJSON(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		h.logger.Info("User logged in via API", slog.String("userID", user.ID), slog.String("email", user.Email))

		// Return token and user info
		response := LoginResponse{
			Token:     token,
			ExpiresAt: time.Now().Add(h.config.Auth.JWTExpiration),
			User: UserInfo{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
				Role:     user.Role,
			},
		}

		h.writeJSON(w, response, http.StatusOK)
	}
}

// POST /v1/auth/logout
func (h *Handler) JsonLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: API logout
		// - Get token from Authorization header
		// - Delete token from database
		// - Return success
		h.logger.Info("API logout handler called")
		h.writeJSON(w, map[string]string{"message": "logged out"}, http.StatusOK)
	}
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
