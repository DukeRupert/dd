package auth

import (
	"github.com/dukerupert/dd/api"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Middleware creates a new auth middleware
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get("session", c)
			if err != nil {
				return api.NewUnauthorizedError("invalid session")
			}

			userID, ok := sess.Values["user_id"].(int64)
			if !ok {
				return api.NewUnauthorizedError("unauthorized")
			}

			email, ok := sess.Values["email"].(string)
			if !ok {
				return api.NewUnauthorizedError("unauthorized")
			}

			c.Set("user_id", userID)
			c.Set("email", email)

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
