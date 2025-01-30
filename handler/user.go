package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/db"
	"golang.org/x/crypto/bcrypt"

	"github.com/labstack/echo/v4"
)

type requestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type resetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// Helper function to generate secure random token
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (app *application) requestPasswordReset(c echo.Context) error {
	var req requestPasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user by email
	user, err := app.queries.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal whether the email exists
			app.logger.Info().Str("email", req.Email).Msg("Password reset requested for non-existent email")
			return c.NoContent(http.StatusAccepted)
		}
		return api.NewDatabaseError(err)
	}

	// Generate reset token
	token, err := generateResetToken()
	if err != nil {
		return api.NewInternalError(err)
	}

	// Store reset token
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = app.queries.CreatePasswordReset(context.Background(), db.CreatePasswordResetParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return api.NewDatabaseError(err)
	}

	// Create reset link
	resetLink := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	// Send password reset email
	if app.mailer == nil {
		app.logger.Error().
			Str("email", user.Email).
			Msg("Password reset requested but email service is not configured")
		return api.NewInternalError(fmt.Errorf("email service not configured"))
	}

	err = app.mailer.SendPasswordResetEmail(user.Email, resetLink)
	if err != nil {
		app.logger.Error().
			Err(err).
			Str("email", user.Email).
			Msg("Failed to send password reset email")
		return api.NewInternalError(fmt.Errorf("failed to send password reset email"))
	}

	app.logger.Info().
		Str("email", user.Email).
		Msg("Password reset email sent")

	// Don't reveal whether the email exists
	return c.NoContent(http.StatusAccepted)
}

func (app *application) resetPassword(c echo.Context) error {
	var req resetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get reset token info
	resetInfo, err := app.queries.GetPasswordResetByToken(context.Background(), req.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewBadRequestError("invalid or expired reset token")
		}
		return api.NewDatabaseError(err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return api.NewInternalError(err)
	}

	// Update user's password
	err = app.queries.UpdateUserPassword(context.Background(), db.UpdateUserPasswordParams{
		ID:           resetInfo.UserID,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return api.NewDatabaseError(err)
	}

	// Mark reset token as used
	err = app.queries.MarkPasswordResetUsed(context.Background(), req.Token)
	if err != nil {
		app.logger.Error().Err(err).Msg("Failed to mark reset token as used")
		// Continue since password was updated successfully
	}

	app.logger.Info().
		Int64("user_id", resetInfo.UserID).
		Msg("Password reset successfully")

	return c.NoContent(http.StatusOK)
}