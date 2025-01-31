package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/dukerupert/dd/views"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/db"
	"github.com/labstack/echo-contrib/session"
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
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required"`
}

type resendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (app *application) registerUser(c echo.Context) error {
    var req registerUserRequest
    if err := c.Bind(&req); err != nil {
        if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
            return views.RegisterPage(views.RegisterData{ErrorMessage: "Invalid request"}).Render(c.Request().Context(), c.Response().Writer)
        }
        return api.NewBadRequestError("invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
            return views.RegisterPage(views.RegisterData{ErrorMessage: "Invalid input"}).Render(c.Request().Context(), c.Response().Writer)
        }
        return err
    }

    _, err := app.queries.GetUserByEmail(context.Background(), req.Email)
    if err == nil {
        if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
            return views.RegisterPage(views.RegisterData{ErrorMessage: "Email already registered"}).Render(c.Request().Context(), c.Response().Writer)
        }
        return api.NewBadRequestError("email already registered")
    } else if err != sql.ErrNoRows {
        return api.NewDatabaseError(err)
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        app.logger.Error().Err(err).Msg("Failed to hash password")
        return api.NewInternalError(err)
    }

    user, err := app.queries.CreateUser(context.Background(), db.CreateUserParams{
        Email:        req.Email,
        PasswordHash: string(hashedPassword),
        FirstName:    req.FirstName,
        LastName:     req.LastName,
    })
    if err != nil {
        return api.NewDatabaseError(err)
    }

    if app.mailer != nil {
        if err := app.sendVerificationEmail(user.ID, user.Email); err != nil {
            app.logger.Error().Err(err).Msg("Failed to send verification email")
        }
    }

    // Create session
    sess, _ := session.Get("session", c)
    sess.Values["user_id"] = user.ID
    sess.Values["email"] = user.Email
    if err := sess.Save(c.Request(), c.Response()); err != nil {
        return api.NewInternalError(err)
    }

    app.logger.Info().
        Int64("user_id", user.ID).
        Str("email", user.Email).
        Msg("User registered successfully")

    if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
        return c.Redirect(http.StatusSeeOther, "/records")
    }

    return c.JSON(http.StatusCreated, echo.Map{
        "user": userResponse{
            ID:        user.ID,
            Email:     user.Email,
            FirstName: user.FirstName,
            LastName:  user.LastName,
            CreatedAt: user.CreatedAt,
        },
    })
}

func (app *application) loginUser(c echo.Context) error {
    ip := c.RealIP()

    if !app.rateLimiter.Allow(ip) {
        remaining, duration := app.rateLimiter.GetRemainingAttempts(ip)
        app.logger.Warn().
            Str("ip", ip).
            Int("remaining_attempts", remaining).
            Dur("lockout_duration", duration).
            Msg("Rate limit exceeded for login attempts")

        errMsg := fmt.Sprintf("Too many login attempts. Try again in %v", duration.Round(time.Minute))
        if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
            return views.LoginPage(views.LoginData{ErrorMessage: errMsg}).Render(c.Request().Context(), c.Response().Writer)
        }
        return api.NewTooManyRequestsError(errMsg)
    }

    var req loginRequest
    if err := c.Bind(&req); err != nil {
        if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
            return views.LoginPage(views.LoginData{ErrorMessage: "Invalid request"}).Render(c.Request().Context(), c.Response().Writer)
        }
        return api.NewBadRequestError("invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
            return views.LoginPage(views.LoginData{ErrorMessage: "Invalid input"}).Render(c.Request().Context(), c.Response().Writer)
        }
        return err
    }

    user, err := app.queries.GetUserByEmail(context.Background(), req.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            app.logger.Info().
                Str("ip", ip).
                Str("email", req.Email).
                Msg("Failed login attempt - user not found")
            
            if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
                return views.LoginPage(views.LoginData{ErrorMessage: "Invalid credentials"}).Render(c.Request().Context(), c.Response().Writer)
            }
            return api.NewUnauthorizedError("invalid credentials")
        }
        return api.NewDatabaseError(err)
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
    if err != nil {
        app.logger.Info().
            Str("ip", ip).
            Str("email", req.Email).
            Msg("Failed login attempt - invalid password")
        
        if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
            return views.LoginPage(views.LoginData{ErrorMessage: "Invalid credentials"}).Render(c.Request().Context(), c.Response().Writer)
        }
        return api.NewUnauthorizedError("invalid credentials")
    }

    // Create session
    sess, _ := session.Get("session", c)
    sess.Values["user_id"] = user.ID
    sess.Values["email"] = user.Email
    if err := sess.Save(c.Request(), c.Response()); err != nil {
        return api.NewInternalError(err)
    }

    app.logger.Info().
        Str("ip", ip).
        Int64("user_id", user.ID).
        Str("email", user.Email).
        Msg("User logged in successfully")

	if isHtmx := c.Request().Header.Get("HX-Request") == "true"; isHtmx {
		c.Response().Header().Set("HX-Redirect", "/records")
		return nil
	}

    return c.JSON(http.StatusOK, echo.Map{
        "user": userResponse{
            ID:        user.ID,
            Email:     user.Email,
            FirstName: user.FirstName,
            LastName:  user.LastName,
            CreatedAt: user.CreatedAt,
        },
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

func (app *application) showLogin(c echo.Context) error {
	// Otherwise render the full page
	return views.LoginPage(views.LoginData{}).Render(c.Request().Context(), c.Response().Writer)
}

func (app *application) showRegister(c echo.Context) error {
	// Otherwise render the full page
	return views.RegisterPage(views.RegisterData{}).Render(c.Request().Context(), c.Response().Writer)
}
