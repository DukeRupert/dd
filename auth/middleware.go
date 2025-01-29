package auth

import (
	"strings"

	"github.com/dukerupert/dd/api"
	"github.com/labstack/echo/v4"
)

// Middleware creates a new auth middleware
func (m *Manager) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return api.NewUnauthorizedError("missing authorization header")
			}

			// Check if the header has the correct format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return api.NewUnauthorizedError("invalid authorization header format")
			}

			// Validate the token
			claims, err := m.ValidateToken(parts[1])
			if err != nil {
				switch err {
				case ErrExpiredToken:
					return api.NewUnauthorizedError("token has expired")
				case ErrInvalidToken:
					return api.NewUnauthorizedError("invalid token")
				default:
					return api.NewUnauthorizedError("authentication failed")
				}
			}

			// Set user information in context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c echo.Context) (int64, error) {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return 0, api.NewUnauthorizedError("user id not found in context")
	}
	return userID, nil
}

// GetUserEmail retrieves the user email from the context
func GetUserEmail(c echo.Context) (string, error) {
	email, ok := c.Get("email").(string)
	if !ok {
		return "", api.NewUnauthorizedError("user email not found in context")
	}
	return email, nil
}