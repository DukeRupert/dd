package api

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"net/http"
	"time"
	"github.com/go-playground/validator/v10"
)

// ErrorHandlerMiddleware creates a middleware for handling errors
func ErrorHandlerMiddleware(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Start time for request timing
			start := time.Now()
			
			err := next(c)
			if err == nil {
				return nil
			}

			// Get or generate request ID
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = c.Request().Header.Get(echo.HeaderXRequestID)
			}

			// Create the error response
			errorResponse := ErrorResponse{
				StatusCode: http.StatusInternalServerError, // Default status code
				Timestamp:  time.Now(),
				RequestID:  requestID,
			}

			// Determine the appropriate status code and error details
			switch e := err.(type) {
			case *AppError:
				errorResponse.StatusCode = e.Status
				errorResponse.Error = Error{
					Code:    e.Code,
					Message: e.Message,
				}
				if e.Err != nil {
					logger.Error().
						Str("request_id", requestID).
						Str("error_code", e.Code).
						Err(e.Err).
						Msg(e.Message)
				}

			case *echo.HTTPError:
				errorResponse.StatusCode = e.Code
				errorResponse.Error = Error{
					Code:    "HTTP_ERROR",
					Message: e.Message.(string),
				}

			case validator.ValidationErrors:
				errorResponse.StatusCode = http.StatusBadRequest
				errorResponse.Error = Error{
					Code:    ErrCodeValidation,
					Message: e.Error(),
				}

			default:
				if err == sql.ErrNoRows {
					errorResponse.StatusCode = http.StatusNotFound
					errorResponse.Error = Error{
						Code:    ErrCodeNotFound,
						Message: "Resource not found",
					}
				} else {
					errorResponse.Error = Error{
						Code:    ErrCodeInternal,
						Message: "Internal server error",
					}
					logger.Error().
						Str("request_id", requestID).
						Err(err).
						Msg("Unhandled error")
				}
			}

			// Log the error with request details
			logger.Error().
				Str("request_id", requestID).
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Int("status", errorResponse.StatusCode).
				Dur("duration", time.Since(start)).
				Msg("Request error")

			// Send error response
			if !c.Response().Committed {
				return c.JSON(errorResponse.StatusCode, errorResponse)
			}
			return nil
		}
	}
}

// RequestIDMiddleware adds a request ID to the context
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("requestTime", time.Now())
			return next(c)
		}
	}
}