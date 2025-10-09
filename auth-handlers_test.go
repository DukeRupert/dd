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
	_ "modernc.org/sqlite"
	"golang.org/x/crypto/bcrypt"
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
	// Initialize validator (required by bind function)
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

func TestHandleSignup(t *testing.T) {
	// Initialize validator
	validate = validator.New()

	// Setup
	db, queries := setupTestDB(t)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	renderer := NewTemplateRenderer()
	if err := renderer.LoadTemplates(); err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	handler := handleSignup(logger, queries, renderer)

	tests := []struct {
		name           string
		email          string
		username       string
		password       string
		wantStatusCode int
		wantRedirect   string
	}{
		{
			name:           "valid signup",
			email:          "newuser@example.com",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusSeeOther,
			wantRedirect:   "/dashboard",
		},
		{
			name:           "duplicate email",
			email:          "admin@example.com", // Already exists from migration
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusConflict,
			wantRedirect:   "",
		},
		{
			name:           "missing email",
			email:          "",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
		{
			name:           "invalid email format",
			email:          "notanemail",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
		{
			name:           "missing username",
			email:          "test@example.com",
			username:       "",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
		{
			name:           "username too short",
			email:          "test@example.com",
			username:       "ab",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
		{
			name:           "missing password",
			email:          "test@example.com",
			username:       "newuser",
			password:       "",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
		{
			name:           "password too short",
			email:          "test@example.com",
			username:       "newuser",
			password:       "short",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create form data
			form := url.Values{}
			form.Add("email", tt.email)
			form.Add("username", tt.username)
			form.Add("password", tt.password)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
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

			// Verify user was created if successful
			if tt.wantStatusCode == http.StatusSeeOther {
				user, err := queries.GetUserByEmail(req.Context(), tt.email)
				if err != nil {
					t.Errorf("User was not created in database: %v", err)
				}

				// Verify password was hashed
				if user.PasswordHash == tt.password {
					t.Error("Password was not hashed")
				}

				// Verify password hash is valid
				err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tt.password))
				if err != nil {
					t.Error("Password hash is invalid")
				}

				// Verify username
				if user.Username != tt.username {
					t.Errorf("Username mismatch: got %v want %v", user.Username, tt.username)
				}

				// Verify role is set to user
				if user.Role != "user" {
					t.Errorf("Role should be 'user', got %v", user.Role)
				}
			}
		})
	}
}

func TestHandleSignupDuplicateUsername(t *testing.T) {
	// Initialize validator
	validate = validator.New()

	// Setup
	db, queries := setupTestDB(t)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	renderer := NewTemplateRenderer()
	if err := renderer.LoadTemplates(); err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	handler := handleSignup(logger, queries, renderer)

	// First signup - should succeed
	form := url.Values{}
	form.Add("email", "user1@example.com")
	form.Add("username", "testuser")
	form.Add("password", "password123")

	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("First signup failed: got status %v", rr.Code)
	}

	// Second signup with same username but different email - should fail
	form = url.Values{}
	form.Add("email", "user2@example.com")
	form.Add("username", "testuser") // Same username
	form.Add("password", "password123")

	req = httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// This should fail due to unique constraint on username
	// The exact status code depends on how you handle database constraint violations
	if rr.Code == http.StatusSeeOther {
		t.Error("Second signup should have failed due to duplicate username")
	}
}