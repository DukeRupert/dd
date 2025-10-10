package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/dukerupert/dd/internal/store"
	"github.com/google/uuid"
)

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession creates a new session for the user and returns the session token
func CreateSession(ctx context.Context, queries *store.Queries, userID string, r *http.Request, duration time.Duration) (string, error) {
	// Generate session token
	token, err := GenerateSecureToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Get IP address and user agent
	ipAddress := getIP(r)
	userAgent := r.UserAgent()

	// Create session in database
	expiresAt := time.Now().Add(duration)
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

// SetSessionCookie sets the session cookie in the response
func SetSessionCookie(w http.ResponseWriter, token, cookieName string, duration time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(duration.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie removes the session cookie
func ClearSessionCookie(w http.ResponseWriter, cookieName string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// Helper functions
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
