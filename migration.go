package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB) error {
	// Ensure migrations directory exists
	if err := os.MkdirAll("migrations", 0755); err != nil {
		return err
	}

	// Initialize sqlite driver for migrations
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	// Initialize migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return err
	}

	// Run migrations
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		log.Println("No migrations to run")
		return nil
	}

	log.Println("Migrations completed successfully")
	return nil
}

// Add this to your main.go initialization
func InitializeDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "vinyl.db")
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}