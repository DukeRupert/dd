package auth

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/store"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database with migrations
func setupTestDB(t *testing.T) (*sql.DB, *store.Queries) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		t.Fatalf("Failed to create migration provider: %v", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db, store.New(db)
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "mySecurePass123", false},
		{"empty password", "", false}, // bcrypt allows empty
		{"72 byte password", string(make([]byte, 72)), false},
		{"exceeds 72 bytes", string(make([]byte, 100)), true}, // bcrypt limit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
		})
	}
}

func TestComparePassword(t *testing.T) {
	password := "mySecurePass123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  bool
	}{
		{"correct password", hash, password, false},
		{"wrong password", hash, "wrongPassword", true},
		{"empty password", hash, "", true},
		{"invalid hash", "invalid", password, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePassword(tt.hash, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateJWT(t *testing.T) {
	secret := "test-secret-key"
	expiration := 1 * time.Hour

	tests := []struct {
		name    string
		userID  string
		email   string
		role    string
		wantErr bool
	}{
		{"valid user", "user123", "test@example.com", "user", false},
		{"admin user", "admin1", "admin@example.com", "admin", false},
		{"empty fields", "", "", "", false}, // JWT allows this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWT(tt.userID, tt.email, tt.role, secret, expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateJWT() returned empty token")
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret"
	userID := "user123"
	email := "test@example.com"
	role := "user"

	// Generate valid token
	validToken, err := GenerateJWT(userID, email, role, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Generate expired token
	expiredToken, err := GenerateJWT(userID, email, role, secret, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		secret    string
		wantError bool
		wantID    string
	}{
		{"valid token", validToken, secret, false, userID},
		{"wrong secret", validToken, wrongSecret, true, ""},
		{"expired token", expiredToken, secret, true, ""},
		{"invalid token", "invalid.token.string", secret, true, ""},
		{"empty token", "", secret, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateJWT(tt.token, tt.secret)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateJWT() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if claims.UserID != tt.wantID {
					t.Errorf("ValidateJWT() userID = %v, want %v", claims.UserID, tt.wantID)
				}
			}
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	tokens := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("GenerateSecureToken() error = %v", err)
		}
		if token == "" {
			t.Error("GenerateSecureToken() returned empty token")
		}
		if len(token) != 64 { // 32 bytes = 64 hex chars
			t.Errorf("GenerateSecureToken() length = %v, want 64", len(token))
		}
		if tokens[token] {
			t.Error("GenerateSecureToken() generated duplicate token")
		}
		tokens[token] = true
	}
}

func TestCreateSession(t *testing.T) {
	db, queries := setupTestDB(t)
	defer db.Close()

	// Create test user
	ctx := context.Background()
	user, err := queries.CreateUser(ctx, store.CreateUserParams{
		ID:           "user123",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hash",
		Role:         "user",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Mock request (nil is acceptable for this test)
	var _ *testing.T // placeholder - you'd use httptest.NewRequest() in real test

	tests := []struct {
		name     string
		userID   string
		duration time.Duration
		wantErr  bool
	}{
		{"valid session", user.ID, 24 * time.Hour, false},
		{"short duration", user.ID, 1 * time.Minute, false},
		{"invalid user", "nonexistent", 24 * time.Hour, false}, // DB allows this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock request
			req := httptest.NewRequest("GET", "/", nil)
			
			token, err := CreateSession(ctx, queries, tt.userID, req, tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if token == "" {
					t.Error("CreateSession() returned empty token")
				}
				// Verify session exists in DB
				session, err := queries.GetSessionByToken(ctx, token)
				if err != nil {
					t.Errorf("Session not found in database: %v", err)
				}
				if session.UserID != tt.userID {
					t.Errorf("Session userID = %v, want %v", session.UserID, tt.userID)
				}
			}
		})
	}
}