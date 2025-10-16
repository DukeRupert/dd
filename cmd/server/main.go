package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/config"
	"github.com/dukerupert/dd/internal/handler"
	"github.com/dukerupert/dd/internal/renderer"
	"github.com/dukerupert/dd/internal/router"
	"github.com/dukerupert/dd/internal/store"
	"github.com/dukerupert/dd/templates"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Set this handler as the default for slog
	logger := slog.New(cfg.Logging.Handler)
	slog.SetDefault(logger)
	logger.Info("logger initialized", slog.String("level", cfg.Logging.Level.String()))

	// Open database
	db, err := sql.Open("sqlite", cfg.Database.Path)
	if err != nil {
		return err
	}
	defer db.Close()

	// CRITICAL: Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if _, err := provider.Up(ctx); err != nil {
		return err
	}

	// Create queries
	queries := store.New(db)

	// Create renderer
	templateRenderer := renderer.New(templates.FS)
	if err := templateRenderer.LoadTemplates(); err != nil {
		return err
	}

	// Create handler
	h := handler.New(logger, queries, templateRenderer, cfg)

	// Create router
	srv := router.New(h, queries, cfg.Session.CookieName)

	slog.Info("Starting server", slog.String("port", strconv.Itoa(cfg.Server.Port)))

	return http.ListenAndServe(net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port)), srv)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
