package handler

import (
	"net/http"

	"github.com/dukerupert/dd/api"
	"github.com/dukerupert/dd/auth"
	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/db"
	"github.com/dukerupert/dd/email"
	"github.com/dukerupert/dd/ratelimit"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type Config struct {
	Config		config.Config
	Queries     *db.Queries
	Logger      zerolog.Logger
	RateLimiter *ratelimit.RateLimiter
	Mailer      *email.Client
}

type application struct {
	config		config.Config
	queries     *db.Queries
	logger      zerolog.Logger
	rateLimiter *ratelimit.RateLimiter
	mailer      *email.Client
}

func NewHandler(cfg Config) *application {
	return &application{
		config:		 cfg.Config,
		queries:     cfg.Queries,
		logger:      cfg.Logger,
		rateLimiter: cfg.RateLimiter,
		mailer:      cfg.Mailer,
	}
}

func (app *application) ApplyMiddleware(e *echo.Echo) {
	// Middleware - order is important
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(api.ErrorHandlerMiddleware(app.logger))
	e.Use(middleware.CORS())
}

func (app *application) CreateRoutes(e *echo.Echo) {
	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Vinyl Collection API")
	})
	e.POST("/register", app.registerUser)
	e.GET("/login", app.showLogin)
	e.POST("/login", app.loginUser)
	e.POST("/forgot-password", app.requestPasswordReset)
	e.POST("/reset-password", app.resetPassword)
	e.GET("/verify-email", app.verifyEmail)
	e.POST("/resend-verification", app.resendVerificationEmail)

	// Protected routes
	protected := e.Group("")
	protected.Use(auth.Middleware())

	// Record endpoints
	protected.GET("/records", app.getAllRecords)
	protected.GET("/records/:id", app.getRecord)
	protected.POST("/records", app.createRecord)
	protected.PUT("/records/:id", app.updateRecord)
	protected.DELETE("/records/:id", app.deleteRecord) // Middleware - order is important

	// Development routes
	if app.config.AppEnv == "development" {
		e.DELETE("/dev/users", app.deleteUserByEmail)
	}

}
