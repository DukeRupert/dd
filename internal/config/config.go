package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Session  SessionConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Host string
	Port int
	Env  string
}

type DatabaseConfig struct {
	Path string
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
}

type SessionConfig struct {
	CookieName string
	Duration   time.Duration
	Secure     bool
}

type LoggingConfig struct {
	Level   string
	Handler slog.Handler
}

// Load loads configuration from environment variables and command-line flags
func Load() (*Config, error) {
	// Command-line flags
	var flagHost = flag.String("host", getEnv("SERVER_HOST", "localhost"), "server host")
	var flagPort = flag.Int("port", getEnvInt("SERVER_PORT", 8080), "server port")
	var flagEnv = flag.String("env", getEnv("ENVIRONMENT", "prod"), "environment: prod, dev")
	var flagLogLevel = flag.String("log_level", getEnv("LOG_LEVEL", "info"), "log level: debug, info, warn, error")
	var flagDatabase = flag.String("database", getEnv("DATABASE_PATH", "sqlite.db"), "sqlite database file path")
	flag.Parse()

	cfg := &Config{
		Server: ServerConfig{
			Host: *flagHost,
			Port: *flagPort,
			Env:  *flagEnv,
		},
		Database: DatabaseConfig{
			Path: *flagDatabase,
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			JWTExpiration: 24 * time.Hour * 7, // 7 days
		},
		Session: SessionConfig{
			CookieName: "session_token",
			Duration:   24 * time.Hour * 7, // 7 days
			Secure:     *flagEnv == "prod" || *flagEnv == "production",
		},
	}

	// Set up logging
	var programLevel = new(slog.LevelVar) // Info by default
	switch *flagLogLevel {
	case "error":
		programLevel.Set(slog.LevelError)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "debug":
		programLevel.Set(slog.LevelDebug)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	var handler slog.Handler
	switch *flagEnv {
	case "prod", "production":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	case "dev", "development":
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	default:
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	}

	cfg.Logging = LoggingConfig{
		Level:   *flagLogLevel,
		Handler: handler,
	}

	// Validate JWT secret in production
	if (cfg.Server.Env == "prod" || cfg.Server.Env == "production") && cfg.Auth.JWTSecret == "your-secret-key-change-in-production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production environment")
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
