package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

func migrate(ctx context.Context, provider *goose.Provider, logger *slog.Logger) error {
	// List migration sources the provider is aware of.
	sources := provider.ListSources()
	for _, s := range sources {
		logger.Debug("migration_list", slog.String("source_type", string(s.Type)), slog.Int64("version", s.Version), slog.String("path", filepath.Base(s.Path)))
	}

	// List status of migrations before applying them.
	stats, err := provider.Status(ctx)
	if err != nil {
		return err
	}
	for _, s := range stats {
		logger.Debug("migration_status", slog.String("source_type", string(s.Source.Type)), slog.Int64("version", s.Source.Version),slog.String("state", string(s.State)))
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}
	for _, r := range results {
		logger.Debug("migration_result",slog.String("source_type", string(r.Source.Type)),slog.Int64("version", r.Source.Version),slog.Int64("duration", int64(r.Duration)))
	}
	return nil
}

func newServer(logger *slog.Logger, queries *store.Queries) http.Handler {
	// Initialize the validator
	validate = validator.New()

	// Initialize the template renderer
	renderer := NewTemplateRenderer()
	if err := renderer.LoadTemplates(); err != nil {
		panic(err)
	}

	// Register routes
	mux := http.NewServeMux()
	addRoutes(mux, logger, queries, renderer)
	return mux
}

func run() error {
	var flagHost = flag.String("host", "localhost", "app host")
	var flagPort = flag.Int("port", 8080, "port, app port")
	var flagEnv = flag.String("env", "prod", "values: prod, dev")
	var flagLogLevel = flag.String("log_level", "info", "values: debug, info, warn, error")
	var flagDatabase = flag.String("sqlite file", "sqlite.db", "sqlite filename")
	flag.Parse()
	
	// logger level
	var programLevel = new(slog.LevelVar) // Info by default
	switch *flagLogLevel {
	case "error":
		programLevel.Set(slog.LevelError)
	case "warn":
		programLevel.Set(slog.LevelError)
	case "debug":
		programLevel.Set(slog.LevelDebug)
	default:
		break
	}
	
	// logger handler
	var handler slog.Handler
	switch *flagEnv {
	case "prod", "production":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	case "dev", "development":
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	case "default":
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// open sqlite database
	db, err := sql.Open("sqlite", *flagDatabase)
	if err != nil {
		return err
	}

	// create goose provider to run migrations
	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		return err
	}

	ctx := context.Background()
	// run migrations
	err = migrate(ctx, provider, logger)
	if err != nil {
		return err
	}

	// sqlc generated queries
	queries := store.New(db)

	srv := newServer(logger, queries)

	logger.Info("Starting server",slog.String("port", strconv.Itoa(*flagPort)))

	err = http.ListenAndServe(net.JoinHostPort(*flagHost, strconv.Itoa(*flagPort)), srv) 
	if err != nil { logger.Error("HTTP Server error", "error", err)
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
