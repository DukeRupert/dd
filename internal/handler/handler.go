package handler

import (
	"log/slog"

	"github.com/dukerupert/dd/internal/config"
	"github.com/dukerupert/dd/internal/renderer"
	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
)

// Handler holds dependencies for all HTTP handlers
type Handler struct {
	logger   *slog.Logger
	queries  *store.Queries
	renderer *renderer.Renderer
	validate *validator.Validate
	config   *config.Config
}

// New creates a new Handler with all dependencies
func New(logger *slog.Logger, queries *store.Queries, renderer *renderer.Renderer, cfg *config.Config) *Handler {
	return &Handler{
		logger:   logger,
		queries:  queries,
		renderer: renderer,
		validate: validator.New(),
		config:   cfg,
	}
}
