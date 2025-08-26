package database

import (
	"context"
	"embed"
	"fmt"	
	"time"

	"github.com/dukerupert/dd/config"
	"github.com/dukerupert/dd/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Database wraps the pgxpool.Pool connection
type Database struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

// NewDatabase creates a new database connection pool and returns a Database instance
func New(DatabaseConfig *config.DatabaseConfig) (*Database, error) {
	ctx := context.Background()

	// Configure pgxpool
	config, err := pgxpool.ParseConfig(DatabaseConfig.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = int32(DatabaseConfig.MaxConns)
	config.MinConns = int32(DatabaseConfig.MaxIdleConns)
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Debug().Msg("Successfully connected to database!")

	// Run migrations using goose with pgx
	if err = runMigrations(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Debug().Msg("Successfully executed migrations!")

	// Create queries instance
    queries := db.New(pool)

    return &Database{
        Pool:    pool,
        Queries: queries,
    }, nil
}

// Close closes the database connection pool
func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

// Ping tests the database connection
func (d *Database) Ping(ctx context.Context) error {
	return d.Pool.Ping(ctx)
}

// GetConn gets a single connection from the pool (useful for transactions)
func (d *Database) GetConn(ctx context.Context) (*pgxpool.Conn, error) {
	return d.Pool.Acquire(ctx)
}

// BeginTx starts a new transaction
func (d *Database) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return d.Pool.Begin(ctx)
}

// runMigrations runs database migrations using goose
func runMigrations(pool *pgxpool.Pool) error {
	// Convert pgxpool to sql.Database for goose compatibility
	Database := stdlib.OpenDBFromPool(pool)
	defer Database.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Error().Err(err).Msg("failed to set goose dialect")
		return err
	}

	if err := goose.Up(Database, "migrations"); err != nil {
		log.Error().Err(err).Msg("failed to run goose migrations")
		return err
	}

	return nil
}
