package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DatabaseURL string

	// Authentication
	SessionSecret string

	// Email
	PostmarkServerToken string
	FromEmail           string

	// Application
	AppEnv  string
	BaseURL string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Only return error if .env file exists but couldn't be loaded
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
		fmt.Println("No .env file found, using environment variables")
	} else {
		fmt.Println(".env file loaded successfully")
	}

	config := &Config{
		// Authentication
		DatabaseURL: os.Getenv("DB_URL"),

		// Authentication
		SessionSecret: os.Getenv("SESSION_SECRET"),

		// Email
		PostmarkServerToken: os.Getenv("POSTMARK_SERVER_TOKEN"),
		FromEmail:           os.Getenv("FROM_EMAIL"),

		// Application
		AppEnv:  os.Getenv("APP_ENV"),
		BaseURL: os.Getenv("BASE_URL"),
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	fmt.Println("\nValidating Configuration:")

	// Required configurations
	if c.DatabaseURL == "" {
		fmt.Println("❌ DB_URL is missing")
		return fmt.Errorf("DB_URL is required")
	}
	fmt.Println("✓ DB_URL is set")

	if c.SessionSecret == "" {
		fmt.Println("❌ SESSION_SECRET is missing")
		return fmt.Errorf("SESSION_SECRET is required")
	}
	fmt.Println("✓ SESSION_SECRET is set")

	// Optional configurations with defaults
	if c.AppEnv == "" {
		c.AppEnv = "development"
		fmt.Println("ℹ️ Using default APP_ENV: development")
	} else {
		fmt.Println("✓ APP_ENV is set:", c.AppEnv)
	}

	if c.BaseURL == "" {
		c.BaseURL = "http://localhost:8080"
		fmt.Println("ℹ️ Using default BASE_URL: http://localhost:8080")
	} else {
		fmt.Println("✓ BASE_URL is set:", c.BaseURL)
	}

	// Email configuration validation
	fmt.Printf("Email Configuration: POSTMARK_TOKEN=%v, FROM_EMAIL=%v\n",
		c.PostmarkServerToken != "", c.FromEmail != "")

	// Email configuration is optional, but both fields must be set if one is
	if (c.PostmarkServerToken == "") != (c.FromEmail == "") {
		return fmt.Errorf("both POSTMARK_SERVER_TOKEN and FROM_EMAIL must be set if using email")
	}

	return nil
}
