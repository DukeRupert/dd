package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/rs/zerolog"
	_ "modernc.org/sqlite"
)

func migrate(ctx context.Context, provider *goose.Provider) error {
	// List migration sources the provider is aware of.
	log.Println("\n=== migration list ===")
	sources := provider.ListSources()
	for _, s := range sources {
		log.Printf("%-3s %-2v %v\n", s.Type, s.Version, filepath.Base(s.Path))
	}

	// List status of migrations before applying them.
	stats, err := provider.Status(ctx)
	if err != nil {
		return err
	}
	log.Println("\n=== migration status ===")
	for _, s := range stats {
		log.Printf("%-3s %-2v %v\n", s.Source.Type, s.Source.Version, s.State)
	}

	log.Println("\n=== log migration output  ===")
	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}
	log.Println("\n=== migration results  ===")
	for _, r := range results {
		log.Printf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration)
	}
	return nil
}

func newServer(logger zerolog.Logger, queries *store.Queries) http.Handler {
	// Initialize the validator
	validate = validator.New()

	// Initialize the template renderer
	renderer := NewTemplateRenderer()
	if err := renderer.LoadTemplates(); err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	// Register routes
	addRoutes(mux, logger, queries, renderer)

	var handler http.Handler = mux
	// Apply middleware
	handler = LoggingMiddleware(handler, logger)
	handler = RequestIDMiddleware(handler)
	return handler
}

func run() error {
	// Load configuration
	_, appConfig, err := config.LoadConfig()
	if err != nil {
		return err
	}

	logger := setupLogger(appConfig.Environment, appConfig.LogLevel)

	ctx := context.Background()

	// open sqlite database
	db, err := sql.Open("sqlite", "sqlite.db")
	if err != nil {
		return err
	}

	// create goose provider to run migrations
	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		return err
	}

	// run migrations
	err = migrate(ctx, provider)
	if err != nil {
		return err
	}

	// sqlc generated queries
	queries := store.New(db)

	srv := newServer(logger, queries)

	logger.Info().
		Str("port", appConfig.Port).
		Msg("Starting server")

	err = http.ListenAndServe(net.JoinHostPort(appConfig.Host, appConfig.Port), srv)
	if err != nil {
		logger.Error().Err(err).Msg("HTTP Server error")
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
