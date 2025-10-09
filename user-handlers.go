package main

import (
	"log/slog"
	"net/http"

	"github.com/dukerupert/dd/internal/store"
)

// HTML Handlers

func handleGetProfile(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get current user's profile
		// - Get user ID from context
		// - Query user from database
		// - Render profile page
		userID, _ := GetUserID(r.Context())
		logger.Info("Get profile handler called", slog.String("userID", userID))
		renderer.Render(w, "profile", nil)
	})
}

func handleGetProfileEditForm(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render edit profile form
		// - Get user ID from context
		// - Query user from database
		// - Render edit form with current data
		userID, _ := GetUserID(r.Context())
		logger.Info("Get profile edit form handler called", slog.String("userID", userID))
		renderer.Render(w, "profile-edit-form", nil)
	})
}

func handlePutProfile(logger *slog.Logger, queries *store.Queries, renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update profile
		// - Get user ID from context
		// - Parse form data
		// - Validate input
		// - Update user in database
		// - Redirect to profile page
		userID, _ := GetUserID(r.Context())
		logger.Info("Update profile handler called", slog.String("userID", userID))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	})
}

func handleGetPasswordForm(renderer *TemplateRenderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Render change password form
		renderer.Render(w, "password-form", nil)
	})
}

func handlePutPassword(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update password
		// - Get user ID from context
		// - Parse form data (current password, new password, confirm password)
		// - Verify current password
		// - Validate new password
		// - Hash new password
		// - Update user in database
		// - Redirect to profile with success message
		userID, _ := GetUserID(r.Context())
		logger.Info("Update password handler called", slog.String("userID", userID))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	})
}

// API Handlers

func handleAPIGetMe(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get current user info
		// - Get user ID from context
		// - Query user from database
		// - Return user JSON (without password)
		userID, _ := GetUserID(r.Context())
		logger.Info("API get me handler called", slog.String("userID", userID))
		writeJSON(w, map[string]string{"message": "current user info"}, http.StatusOK)
	})
}

func handleAPIGetUser(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Get user by ID
		// - Parse ID from path
		// - Check permissions (admin or own profile)
		// - Query user from database
		// - Return user JSON (without password)
		id := r.PathValue("id")
		logger.Info("API get user handler called", slog.String("id", id))
		writeJSON(w, map[string]string{"message": "user details"}, http.StatusOK)
	})
}

func handleAPIPutUser(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Update user
		// - Parse ID from path
		// - Check permissions (admin or own profile)
		// - Parse JSON body
		// - Validate input
		// - Update user in database
		// - Return updated user (without password)
		id := r.PathValue("id")
		logger.Info("API update user handler called", slog.String("id", id))
		writeJSON(w, map[string]string{"message": "user updated"}, http.StatusOK)
	})
}

func handleAPIDeleteUser(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete user (admin only)
		// - Parse ID from path
		// - Check admin permissions
		// - Prevent self-deletion
		// - Delete user from database (cascade sessions/tokens)
		// - Return 204 No Content
		id := r.PathValue("id")
		logger.Info("API delete user handler called", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)
	})
}

func handleAPIPutUserPassword(logger *slog.Logger, queries *store.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		logger.Info("API update password handler called", slog.String("id", id))
		writeJSON(w, map[string]string{"message": "password updated"}, http.StatusOK)
	})
}