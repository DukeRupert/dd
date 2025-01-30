package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/db"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type registerUserRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	FirstName string `json:"first_name" validate:"required,max=50"`
	LastName  string `json:"last_name" validate:"required,max=50"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type tokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type resendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (app *application) registerUser(c echo.Context) error {
	var req registerUserRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Check if user already exists
	_, err := app.queries.GetUserByEmail(context.Background(), req.Email)
	if err == nil {
		return api.NewBadRequestError("email already registered")
	} else if err != sql.ErrNoRows {
		return api.NewDatabaseError(err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		app.logger.Error().Err(err).Msg("Failed to hash password")
		return api.NewInternalError(err)
	}

	// Create user
	user, err := app.queries.CreateUser(context.Background(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	})
	if err != nil {
		return api.NewDatabaseError(err)
	}

	// Send verification email
	if app.mailer != nil {
		if err := app.sendVerificationEmail(user.ID, user.Email); err != nil {
			app.logger.Error().Err(err).Msg("Failed to send verification email")
			// Continue registration process even if email fails
		}
	}

	// Generate JWT token
	token, err := app.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return api.NewInternalError(err)
	}

	app.logger.Info().
		Int64("user_id", user.ID).
		Str("email", user.Email).
		Msg("User registered successfully")

	return c.JSON(http.StatusCreated, echo.Map{
		"user": userResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
		},
		"token": token,
	})
}

func (app *application) loginUser(c echo.Context) error {
	// Get IP address for rate limiting
	ip := c.RealIP()

	// Check rate limit
	if !app.rateLimiter.Allow(ip) {
		remaining, duration := app.rateLimiter.GetRemainingAttempts(ip)
		app.logger.Warn().
			Str("ip", ip).
			Int("remaining_attempts", remaining).
			Dur("lockout_duration", duration).
			Msg("Rate limit exceeded for login attempts")

		return api.NewTooManyRequestsError(fmt.Sprintf("Too many login attempts. Try again in %v", duration.Round(time.Minute)))
	}

	var req loginRequest
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
			app.logger.Info().
				Str("ip", ip).
				Str("email", req.Email).
				Msg("Failed login attempt - user not found")
			return api.NewUnauthorizedError("invalid credentials")
		}
		return api.NewDatabaseError(err)
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		app.logger.Info().
			Str("ip", ip).
			Str("email", req.Email).
			Msg("Failed login attempt - invalid password")
		return api.NewUnauthorizedError("invalid credentials")
	}

	// Generate token
	token, err := app.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return api.NewInternalError(err)
	}

	// Generate refresh token
	refreshToken, err := app.auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return api.NewInternalError(err)
	}

	app.logger.Info().
		Str("ip", ip).
		Int64("user_id", user.ID).
		Str("email", user.Email).
		Msg("User logged in successfully")

	return c.JSON(http.StatusOK, echo.Map{
		"user": userResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
		},
		"token":         token,
		"refresh_token": refreshToken,
	})
}

func (app *application) refreshToken(c echo.Context) error {
	var req refreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return api.NewBadRequestError("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate refresh token
	userID, err := app.auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		app.logger.Info().
			Str("error", err.Error()).
			Msg("Invalid refresh token")
		return api.NewUnauthorizedError("invalid refresh token")
	}

	// Get user details
	user, err := app.queries.GetUserByID(context.Background(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewUnauthorizedError("user not found")
		}
		return api.NewDatabaseError(err)
	}

	// Generate new tokens
	newToken, err := app.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return api.NewInternalError(err)
	}

	newRefreshToken, err := app.auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return api.NewInternalError(err)
	}

	app.logger.Info().
		Int64("user_id", user.ID).
		Msg("Tokens refreshed successfully")

	return c.JSON(http.StatusOK, tokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	})
}

func (app *application) sendVerificationEmail(userID int64, email string) error {
	// Generate verification token
	token, err := generateResetToken() // Using the same secure token generator we use for password resets
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store verification token
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = app.queries.CreateEmailVerification(context.Background(), db.CreateEmailVerificationParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}

	// Create verification link
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", app.config.BaseURL, token)

	// Send verification email
	if err := app.mailer.SendVerificationEmail(email, verificationLink); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (app *application) resendVerificationEmail(c echo.Context) error {
	var req resendVerificationRequest
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
			return c.NoContent(http.StatusAccepted)
		}
		return api.NewDatabaseError(err)
	}

	// Check if already verified
	isVerified, err := app.queries.IsEmailVerified(context.Background(), user.ID)
	if err != nil {
		return api.NewDatabaseError(err)
	}

	if isVerified {
		return api.NewBadRequestError("email already verified")
	}

	// Send verification email
	if err := app.sendVerificationEmail(user.ID, user.Email); err != nil {
		app.logger.Error().Err(err).Msg("Failed to send verification email")
		return api.NewInternalError(err)
	}

	app.logger.Info().
		Int64("user_id", user.ID).
		Str("email", user.Email).
		Msg("Verification email resent")

	return c.NoContent(http.StatusAccepted)
}

func (app *application) verifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return api.NewBadRequestError("missing verification token")
	}

	// Get verification by token
	verification, err := app.queries.GetEmailVerificationByToken(context.Background(), token)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NewBadRequestError("invalid or expired verification token")
		}
		return api.NewDatabaseError(err)
	}

	// Mark email as verified
	if err := app.queries.MarkEmailVerified(context.Background(), verification.UserID); err != nil {
		return api.NewDatabaseError(err)
	}

	// Mark verification token as used
	if err := app.queries.MarkEmailVerificationUsed(context.Background(), token); err != nil {
		app.logger.Error().Err(err).Msg("Failed to mark verification token as used")
		// Continue since email was verified successfully
	}

	app.logger.Info().
		Int64("user_id", verification.UserID).
		Msg("Email verified successfully")

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Email verified successfully",
	})
}
