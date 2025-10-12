package handler

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/middleware"
)

// HTML Handlers

func (h *Handler) handleGetProfile() http.HandlerFunc {
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

func (h *Handler) handleGetProfileEditForm() http.HandlerFunc {
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

func (h *Handler) handlePutProfile() http.HandlerFunc {
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

func (h *Handler) handleGetPasswordForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render change password form
		h.renderer.Render(w, "password-form", nil)
	}
}

func (h *Handler) handlePutPassword() http.HandlerFunc {
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

func (h *Handler) handleAPIGetMe() http.HandlerFunc {
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

func (h *Handler) handleAPIGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get user by ID
		// - Parse ID from path
		// - Check permissions (admin or own profile)
		// - Query user from database
		// - Return user JSON (without password)
		id := r.PathValue("id")
		h.logger.Info("API get user handler called", slog.String("id", id))
		h.writeJSON(w, map[string]string{"message": "user details"}, http.StatusOK)
	}
}

func (h *Handler) handleAPIPutUser() http.HandlerFunc {
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

func (h *Handler) handleAPIDeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete user (admin only)
		// - Parse ID from path
		// - Check admin permissions
		// - Prevent self-deletion
		// - Delete user from database (cascade sessions/tokens)
		// - Return 204 No Content
		id := r.PathValue("id")
		h.logger.Info("API delete user handler called", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) handleAPIPutUserPassword() http.HandlerFunc {
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