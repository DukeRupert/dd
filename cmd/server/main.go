package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/dukerupert/dd/templates"
	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/config"
	"github.com/dukerupert/dd/internal/handler"
	"github.com/dukerupert/dd/internal/renderer"
	"github.com/dukerupert/dd/internal/router"
	"github.com/dukerupert/dd/internal/store"
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

	// Setup logger
	logger := slog.New(cfg.Logging.Handler)
	slog.SetDefault(logger)

	// Open database
	db, err := sql.Open("sqlite", cfg.Database.Path)
	if err != nil {
		return err
	}
	defer db.Close()

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
	r := renderer.New(templates.FS)
	if err := r.LoadTemplates(); err != nil {
		return err
	}

	// Create handler
	h := handler.New(logger, queries, r, cfg)

	// Create router
	srv := router.New(h, queries, cfg.Session.CookieName)

	logger.Info("Starting server", slog.String("port", strconv.Itoa(cfg.Server.Port)))

	return http.ListenAndServe(net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port)), srv)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}