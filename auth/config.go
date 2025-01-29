package auth

import (
	"time"
)

// Config holds the JWT configuration
type Config struct {
	SecretKey           string
	TokenExpiration     time.Duration
	RefreshTokenExpiration time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		SecretKey:             "your-secret-key-change-this-in-production",
		TokenExpiration:       24 * time.Hour,    // 24 hours
		RefreshTokenExpiration: 7 * 24 * time.Hour, // 7 days
	}
}