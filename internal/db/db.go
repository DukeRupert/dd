package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/dukerupert/dd/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// DB wraps the sql.DB connection
type DB struct {
	DB *sql.DB
}

// NewDB creates a new database connection and returns a DB instance
func NewDB(dbConfig *config.DatabaseConfig) (*DB, error) {
	// Connect to database
	db, err := sql.Open("postgres", dbConfig.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(dbConfig.MaxConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database!")

	// Run migrations
	goose.SetBaseFS(embedMigrations)
	if err = runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations", err)
	}

	log.Println("Successfully executed migrations!")

	// Return the wrapped DB instance
	return &DB{DB: db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

func (d *DB) Ping() error {
	return d.DB.Ping()
}

func runMigrations(db *sql.DB) error {
	// Implementation for running migrations
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	return nil
}
