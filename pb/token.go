package pocketbase

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TokenStore interface for storing and retrieving auth tokens
type TokenStore interface {
	SetToken(token string)
	GetToken() string
	ClearToken()
	IsValid() bool
}

// MemoryTokenStore stores the token in memory
type MemoryTokenStore struct {
	token string
}

// NewMemoryTokenStore creates a new in-memory token store
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{}
}

// SetToken stores the token
func (s *MemoryTokenStore) SetToken(token string) {
	s.token = token
}

// GetToken retrieves the stored token
func (s *MemoryTokenStore) GetToken() string {
	return s.token
}

// ClearToken removes the stored token
func (s *MemoryTokenStore) ClearToken() {
	s.token = ""
}

// IsValid checks if the token exists and is not expired
func (s *MemoryTokenStore) IsValid() bool {
	if s.token == "" {
		return false
	}

	// Parse the JWT to check if it's expired
	parts := strings.Split(s.token, ".")
	if len(parts) != 3 {
		return false
	}

	// Parse without verifying signature - we just want to check the expiration
	token, _ := jwt.Parse(s.token, func(token *jwt.Token) (interface{}, error) {
		return []byte("dummy-key-for-parsing-only"), nil
	})

	if token == nil {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// Check if token is expired
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}

	expTime := time.Unix(int64(exp), 0)
	return time.Now().Before(expTime)
}