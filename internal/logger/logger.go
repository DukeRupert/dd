// logger/logger.go
package logger

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds logger configuration
type Config struct {
	Development bool
	Level       string
}

// Setup configures the global zerolog logger
func Setup(config Config) {
	// Parse log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if config.Development {
		// Pretty console output for development
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		})
	} else {
		// Structured JSON output for production
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Caller().
			Logger()
	}

	log.Info().
		Bool("development", config.Development).
		Str("level", level.String()).
		Msg("Logger initialized")
}

// Middleware returns Echo middleware for request logging
func Middleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogError:     true,
		LogMethod:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			var event *zerolog.Event
			switch {
			case v.Status >= 500:
				event = log.Error()
			case v.Status >= 400:
				event = log.Warn()
			case v.Status >= 300:
				event = log.Info()
			default:
				event = log.Info()
			}
			
			event.
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Str("remote_ip", v.RemoteIP).
				Str("bytes_in", v.ContentLength).
				Int64("bytes_out", v.ResponseSize)

			if v.Error != nil {
				event.Err(v.Error)
			}

			event.Msg("HTTP request")
			return nil
		},
	})
}

// ErrorHandler returns a custom error handler that logs with zerolog
func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := 500
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if msg, ok := he.Message.(string); ok {
				message = msg
			}
		}

		// Log the error
		log.Error().
			Err(err).
			Int("status", code).
			Str("method", c.Request().Method).
			Str("path", c.Request().URL.Path).
			Str("remote_ip", c.RealIP()).
			Msg("HTTP error")

		// Send response if not already committed
		if !c.Response().Committed {
			if c.Request().Method == "HEAD" {
				err = c.NoContent(code)
			} else {
				err = c.JSON(code, map[string]interface{}{
					"error":  message,
					"status": code,
				})
			}
			if err != nil {
				log.Error().Err(err).Msg("Failed to send error response")
			}
		}
	}
}

// Get returns a logger with additional context
func Get() zerolog.Logger {
	return log.Logger
}

// WithContext returns a logger with request context
func WithContext(c echo.Context) zerolog.Logger {
	return log.With().
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Logger()
}