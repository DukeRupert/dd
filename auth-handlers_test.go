package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/store"
	"github.com/go-playground/validator/v10"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *store.Queries) {
	t.Helper()

	// Create in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		t.Fatalf("Failed to create goose provider: %v", err)
	}

	ctx := context.Background()
	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	queries := store.New(db)
	return db, queries
}

func TestHandleLogin(t *testing.T) {
	validate = validator.New()

	// Setup
	db, queries := setupTestDB(t)
	defer db.Close()

	// Create a test user with a known password
	testPassword := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Update the admin user's password to our test password
	_, err = db.Exec("UPDATE users SET password_hash = ? WHERE email = ?", string(hashedPassword), "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to update admin password: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	renderer := NewTemplateRenderer()
	if err := renderer.LoadTemplates(); err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	handler := handleLogin(logger, queries, renderer)

	tests := []struct {
		name           string
		email          string
		password       string
		wantStatusCode int
		wantRedirect   string
	}{
		{
			name:           "valid admin credentials",
			email:          "admin@example.com",
			password:       testPassword,
			wantStatusCode: http.StatusSeeOther,
			wantRedirect:   "/dashboard",
		},
		{
			name:           "invalid email",
			email:          "notfound@example.com",
			password:       "anypassword",
			wantStatusCode: http.StatusUnauthorized,
			wantRedirect:   "",
		},
		{
			name:           "invalid password",
			email:          "admin@example.com",
			password:       "wrongpassword",
			wantStatusCode: http.StatusUnauthorized,
			wantRedirect:   "",
		},
		{
			name:           "missing email",
			email:          "",
			password:       "password",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
		{
			name:           "missing password",
			email:          "admin@example.com",
			password:       "",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
		{
			name:           "invalid email format",
			email:          "notanemail",
			password:       "password",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create form data
			form := url.Values{}
			form.Add("email", tt.email)
			form.Add("password", tt.password)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}

			// Check redirect location if expecting redirect
			if tt.wantRedirect != "" {
				location := rr.Header().Get("Location")
				if location != tt.wantRedirect {
					t.Errorf("handler returned wrong redirect: got %v want %v", location, tt.wantRedirect)
				}
			}
		})
	}
}