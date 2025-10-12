package handler

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/middleware"
)

// HTML Handlers

// GET /profile
func (h *Handler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get current user's profile
		// - Get user ID from context
		// - Query user from database
		// - Render profile page
		userID, _ := middleware.GetUserID(r.Context())
		h.logger.Info("Get profile handler called", slog.String("userID", userID))
		h.renderer.Render(w, "profile", nil)
	}
}

// GET /profile/edit
func (h *Handler) GetUpdateProfileForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render edit profile form
		// - Get user ID from context
		// - Query user from database
		// - Render edit form with current data
		userID, _ := middleware.GetUserID(r.Context())
		h.logger.Info("Get profile edit form handler called", slog.String("userID", userID))
		h.renderer.Render(w, "profile-edit-form", nil)
	}
}

// PUT /users/{id}
func (h *Handler) UpdateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update profile
		// - Get user ID from context
		// - Parse form data
		// - Validate input
		// - Update user in database
		// - Redirect to profile page
		userID, _ := middleware.GetUserID(r.Context())
		h.logger.Info("Update profile handler called", slog.String("userID", userID))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}

// GET /profile/password
func (h *Handler) GetUpdatePasswordForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render change password form
		h.renderer.Render(w, "password-form", nil)
	}
}

// PUT /profile/password
func (h *Handler) UpdatePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update password
		// - Get user ID from context
		// - Parse form data (current password, new password, confirm password)
		// - Verify current password
		// - Validate new password
		// - Hash new password
		// - Update user in database
		// - Redirect to profile with success message
		userID, _ := middleware.GetUserID(r.Context())
		h.logger.Info("Update password handler called", slog.String("userID", userID))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}

// API Handlers

// GET /v1/profile
func (h *Handler) JsonGetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get current user info
		// - Get user ID from context
		// - Query user from database
		// - Return user JSON (without password)
		userID, _ := middleware.GetUserID(r.Context())
		h.logger.Info("API get me handler called", slog.String("userID", userID))
		h.writeJSON(w, map[string]string{"message": "current user info"}, http.StatusOK)
	}
}

// PUT /v1/profile
func (h *Handler) JsonUpdateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update user
		// - Parse ID from path
		// - Check permissions (admin or own profile)
		// - Parse JSON body
		// - Validate input
		// - Update user in database
		// - Return updated user (without password)
		id := r.PathValue("id")
		h.logger.Info("API update user handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "user updated"}, http.StatusOK)
	}
}

// PUT /v1/profile/password
func (h *Handler) JsonUpdatePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Change password via API
		// - Parse ID from path
		// - Check permissions (admin or own profile)
		// - Parse JSON body (current password, new password)
		// - Verify current password (unless admin)
		// - Validate new password
		// - Hash new password
		// - Update user in database
		// - Return success message
		id := r.PathValue("id")
		h.logger.Info("API update password handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "password updated"}, http.StatusOK)
	}
}
