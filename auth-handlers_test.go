package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

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
		checkSession   bool
	}{
		{
			name:           "valid admin credentials",
			email:          "admin@example.com",
			password:       testPassword,
			wantStatusCode: http.StatusSeeOther,
			wantRedirect:   "/dashboard",
			checkSession:   true,
		},
		{
			name:           "invalid email",
			email:          "notfound@example.com",
			password:       "anypassword",
			wantStatusCode: http.StatusUnauthorized,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "invalid password",
			email:          "admin@example.com",
			password:       "wrongpassword",
			wantStatusCode: http.StatusUnauthorized,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "missing email",
			email:          "",
			password:       "password",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "missing password",
			email:          "admin@example.com",
			password:       "",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "invalid email format",
			email:          "notanemail",
			password:       "password",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
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

			// Check session was created
			if tt.checkSession {
				cookies := rr.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == SessionCookieName {
						sessionCookie = cookie
						break
					}
				}

				if sessionCookie == nil {
					t.Error("Session cookie was not set")
				} else {
					// Verify session exists in database
					session, err := queries.GetSessionByToken(req.Context(), sessionCookie.Value)
					if err != nil {
						t.Errorf("Session was not created in database: %v", err)
					}

					// Verify session belongs to the user
					user, _ := queries.GetUserByEmail(req.Context(), tt.email)
					if session.UserID != user.ID {
						t.Errorf("Session user_id mismatch: got %v want %v", session.UserID, user.ID)
					}

					// Verify session hasn't expired
					if session.ExpiresAt.Before(time.Now()) {
						t.Error("Session already expired")
					}
				}
			}
		})
	}
}

func TestHandleAPILogin(t *testing.T) {
	// Initialize validator
	validate = validator.New()

	// Setup
	db, queries := setupTestDB(t)
	defer db.Close()

	// Create a test user with a known password
	testEmail := "apiuser@example.com"
	testPassword := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Create test user
	userID := generateUUID()
	_, err = db.Exec(`
		INSERT INTO users (id, email, username, password_hash, role, is_active, email_verified)
		VALUES (?, ?, ?, ?, ?, 1, 1)
	`, userID, testEmail, "apiuser", string(hashedPassword), "user")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := handleAPILogin(logger, queries)

	tests := []struct {
		name           string
		email          string
		password       string
		wantStatusCode int
		checkToken     bool
	}{
		{
			name:           "valid credentials",
			email:          testEmail,
			password:       testPassword,
			wantStatusCode: http.StatusOK,
			checkToken:     true,
		},
		{
			name:           "invalid email",
			email:          "notfound@example.com",
			password:       "anypassword",
			wantStatusCode: http.StatusUnauthorized,
			checkToken:     false,
		},
		{
			name:           "invalid password",
			email:          testEmail,
			password:       "wrongpassword",
			wantStatusCode: http.StatusUnauthorized,
			checkToken:     false,
		},
		{
			name:           "missing email",
			email:          "",
			password:       "password",
			wantStatusCode: http.StatusBadRequest,
			checkToken:     false,
		},
		{
			name:           "missing password",
			email:          testEmail,
			password:       "",
			wantStatusCode: http.StatusBadRequest,
			checkToken:     false,
		},
		{
			name:           "invalid email format",
			email:          "notanemail",
			password:       "password",
			wantStatusCode: http.StatusBadRequest,
			checkToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JSON request body
			reqBody := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			jsonBody, _ := json.Marshal(reqBody)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(string(jsonBody)))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", rr.Code, tt.wantStatusCode, rr.Body.String())
			}

			// Check token was returned
			if tt.checkToken {
				var response struct {
					Token     string    `json:"token"`
					ExpiresAt time.Time `json:"expires_at"`
					User      struct {
						ID       string `json:"id"`
						Email    string `json:"email"`
						Username string `json:"username"`
						Role     string `json:"role"`
					} `json:"user"`
				}

				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Verify token is not empty
				if response.Token == "" {
					t.Error("Token was not returned")
				}

				// Verify token is valid
				claims, err := validateJWT(response.Token)
				if err != nil {
					t.Errorf("Token validation failed: %v", err)
				}

				// Verify claims
				if claims.UserID != userID {
					t.Errorf("Token UserID mismatch: got %v want %v", claims.UserID, userID)
				}
				if claims.Email != testEmail {
					t.Errorf("Token Email mismatch: got %v want %v", claims.Email, testEmail)
				}
				if claims.Role != "user" {
					t.Errorf("Token Role mismatch: got %v want %v", claims.Role, "user")
				}

				// Verify user info in response
				if response.User.ID != userID {
					t.Errorf("Response UserID mismatch: got %v want %v", response.User.ID, userID)
				}
				if response.User.Email != testEmail {
					t.Errorf("Response Email mismatch: got %v want %v", response.User.Email, testEmail)
				}
				if response.User.Username != "apiuser" {
					t.Errorf("Response Username mismatch: got %v want %v", response.User.Username, "apiuser")
				}
				if response.User.Role != "user" {
					t.Errorf("Response Role mismatch: got %v want %v", response.User.Role, "user")
				}

				// Verify expires_at is in the future
				if response.ExpiresAt.Before(time.Now()) {
					t.Error("ExpiresAt should be in the future")
				}
			}
		})
	}
}

func TestHandleAPISignup(t *testing.T) {
	// Initialize validator
	validate = validator.New()

	// Setup
	db, queries := setupTestDB(t)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := handleAPISignup(logger, queries)

	tests := []struct {
		name           string
		email          string
		username       string
		password       string
		wantStatusCode int
		checkUser      bool
	}{
		{
			name:           "valid signup",
			email:          "newuser@example.com",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusCreated,
			checkUser:      true,
		},
		{
			name:           "duplicate email",
			email:          "admin@example.com", // Already exists from migration
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusConflict,
			checkUser:      false,
		},
		{
			name:           "missing email",
			email:          "",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			checkUser:      false,
		},
		{
			name:           "invalid email format",
			email:          "notanemail",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			checkUser:      false,
		},
		{
			name:           "missing username",
			email:          "test@example.com",
			username:       "",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			checkUser:      false,
		},
		{
			name:           "username too short",
			email:          "test@example.com",
			username:       "ab",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			checkUser:      false,
		},
		{
			name:           "missing password",
			email:          "test@example.com",
			username:       "newuser",
			password:       "",
			wantStatusCode: http.StatusBadRequest,
			checkUser:      false,
		},
		{
			name:           "password too short",
			email:          "test@example.com",
			username:       "newuser",
			password:       "short",
			wantStatusCode: http.StatusBadRequest,
			checkUser:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JSON request body
			reqBody := map[string]string{
				"email":    tt.email,
				"username": tt.username,
				"password": tt.password,
			}
			jsonBody, _ := json.Marshal(reqBody)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(string(jsonBody)))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", rr.Code, tt.wantStatusCode, rr.Body.String())
			}

			// Verify user was created and check response
			if tt.checkUser {
				var response struct {
					Token     string    `json:"token"`
					ExpiresAt time.Time `json:"expires_at"`
					User      struct {
						ID       string `json:"id"`
						Email    string `json:"email"`
						Username string `json:"username"`
						Role     string `json:"role"`
					} `json:"user"`
				}

				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Failed to decode response: %v, body: %s", err, rr.Body.String())
				}

				// Verify token is not empty
				if response.Token == "" {
					t.Error("Token was not returned")
				}

				// Verify token is valid
				claims, err := validateJWT(response.Token)
				if err != nil {
					t.Errorf("Token validation failed: %v", err)
				}

				// Verify claims match user
				if claims.Email != tt.email {
					t.Errorf("Token Email mismatch: got %v want %v", claims.Email, tt.email)
				}
				if claims.Role != "user" {
					t.Errorf("Token Role mismatch: got %v want %v", claims.Role, "user")
				}

				// Verify user info in response
				if response.User.ID == "" {
					t.Error("Response UserID is empty")
				}
				if response.User.Email != tt.email {
					t.Errorf("Response Email mismatch: got %v want %v", response.User.Email, tt.email)
				}
				if response.User.Username != tt.username {
					t.Errorf("Response Username mismatch: got %v want %v", response.User.Username, tt.username)
				}
				if response.User.Role != "user" {
					t.Errorf("Response Role mismatch: got %v want %v", response.User.Role, "user")
				}

				// Verify expires_at is in the future
				if response.ExpiresAt.Before(time.Now()) {
					t.Error("ExpiresAt should be in the future")
				}

				// Verify user exists in database
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

	handler := handleSignup(logger, queries)

	tests := []struct {
		name           string
		email          string
		username       string
		password       string
		wantStatusCode int
		wantRedirect   string
		checkSession   bool
	}{
		{
			name:           "valid signup",
			email:          "newuser@example.com",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusSeeOther,
			wantRedirect:   "/dashboard",
			checkSession:   true,
		},
		{
			name:           "duplicate email",
			email:          "admin@example.com", // Already exists from migration
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusConflict,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "missing email",
			email:          "",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "invalid email format",
			email:          "notanemail",
			username:       "newuser",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "missing username",
			email:          "test@example.com",
			username:       "",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "username too short",
			email:          "test@example.com",
			username:       "ab",
			password:       "password123",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "missing password",
			email:          "test@example.com",
			username:       "newuser",
			password:       "",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
		},
		{
			name:           "password too short",
			email:          "test@example.com",
			username:       "newuser",
			password:       "short",
			wantStatusCode: http.StatusBadRequest,
			wantRedirect:   "",
			checkSession:   false,
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

			// Check session was created
			if tt.checkSession {
				cookies := rr.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == SessionCookieName {
						sessionCookie = cookie
						break
					}
				}

				if sessionCookie == nil {
					t.Error("Session cookie was not set")
				} else {
					// Verify session exists in database
					session, err := queries.GetSessionByToken(req.Context(), sessionCookie.Value)
					if err != nil {
						t.Errorf("Session was not created in database: %v", err)
					}

					// Verify session belongs to the user
					user, _ := queries.GetUserByEmail(req.Context(), tt.email)
					if session.UserID != user.ID {
						t.Errorf("Session user_id mismatch: got %v want %v", session.UserID, user.ID)
					}

					// Verify session hasn't expired
					if session.ExpiresAt.Before(time.Now()) {
						t.Error("Session already expired")
					}
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

	handler := handleSignup(logger, queries)

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

func TestHandleLogout(t *testing.T) {
	// Initialize validator
	validate = validator.New()

	// Setup
	db, queries := setupTestDB(t)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	handler := handleLogout(logger, queries)

	// Create a test session
	testUserID := "test-user-id"
	testToken := "test-session-token"
	sessionID := generateUUID()
	
	_, err := queries.CreateSession(context.Background(), store.CreateSessionParams{
		ID:        sessionID,
		UserID:    testUserID,
		Token:     testToken,
		IpAddress: toNullString("127.0.0.1"),
		UserAgent: toNullString("test-agent"),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Create request with session cookie
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: testToken,
	})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rr, req)

	// Check status code
	if rr.Code != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// Check redirect location
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("handler returned wrong redirect: got %v want %v", location, "/")
	}

	// Verify session was deleted from database
	_, err = queries.GetSessionByToken(context.Background(), testToken)
	if err == nil {
		t.Error("Session was not deleted from database")
	}

	// Verify cookie was cleared
	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Error("Session cookie was not set in response")
	} else if sessionCookie.MaxAge != -1 {
		t.Errorf("Session cookie MaxAge should be -1, got %v", sessionCookie.MaxAge)
	}
}