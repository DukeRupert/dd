package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/dukerupert/dd/internal/store"
)

const (
	SessionCookieName = "session_token"
	SessionDuration   = 24 * time.Hour * 7 // 7 days
)

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// createSession creates a new session for the user and returns the session token
func createSession(ctx context.Context, queries *store.Queries, userID string, r *http.Request) (string, error) {
	// Generate session token
	token, err := generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate session ID
	sessionID := generateUUID()

	// Get IP address and user agent
	ipAddress := getIP(r)
	userAgent := r.UserAgent()

	// Create session in database
	expiresAt := time.Now().Add(SessionDuration)
	_, err = queries.CreateSession(ctx, store.CreateSessionParams{
		ID:        sessionID,
		UserID:    userID,
		Token:     token,
		IpAddress: toNullString(ipAddress),
		UserAgent: toNullString(userAgent),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

// setSessionCookie sets the session cookie in the response
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie removes the session cookie
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// toNullString converts a string to sql.NullString
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}