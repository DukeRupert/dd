package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// DatabaseConfig holds all database configuration
type DatabaseConfig struct {
	Host         string
	Port         int
	Name         string
	User         string
	Password     string
	SSLMode      string
	MaxConns     int
	MaxIdleConns int
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host        string
	Port        string
	LogLevel    string
	Environment string
}

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() (*DatabaseConfig, *ServerConfig, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we only log the error
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load database configuration
	dbConfig, err := loadDatabaseConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load database config: %w", err)
	}

	// Load server configuration
	serverConfig := loadServerConfig()

	return dbConfig, serverConfig, nil
}

func loadDatabaseConfig() (*DatabaseConfig, error) {
	config := &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Name:     getEnv("DB_NAME", "doxie_discs"),
		User:     getEnv("DB_USER", "doxie_user"),
		Password: getEnv("DB_PASSWORD", "serenity"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Parse integer values
	var err error

	config.Port, err = getEnvAsInt("DB_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	config.MaxConns, err = getEnvAsInt("DB_MAXCONNS", 10)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAXCONNS: %w", err)
	}

	config.MaxIdleConns, err = getEnvAsInt("DB_MAXIDLECONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAXIDLECONNS: %w", err)
	}

	return config, nil
}

func loadServerConfig() *ServerConfig {
	config := &ServerConfig{
		Host:        getEnv("HOST", "0.0.0.0"),
		Port:        getEnv("PORT", "1234"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENV", "development"),
	}

	return config
}

// Helper functions

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) (int, error) {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	}
	return defaultValue, nil
}

// GetConnectionString builds a PostgreSQL connection string
func (db *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
	)
}

// GetConnectionStringForLogging returns a connection string without the password
func (db *DatabaseConfig) GetConnectionStringForLogging() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Name, db.SSLMode,
	)
}

// PrintConfig prints the loaded configuration (without sensitive data)
func PrintConfig(dbConfig *DatabaseConfig, serverConfig *ServerConfig) {
	fmt.Println("=== Loaded Configuration ===")
	fmt.Printf("Database:\n")
	fmt.Printf("  Host: %s\n", dbConfig.Host)
	fmt.Printf("  Port: %d\n", dbConfig.Port)
	fmt.Printf("  Name: %s\n", dbConfig.Name)
	fmt.Printf("  User: %s\n", dbConfig.User)
	fmt.Printf("  Password: ********\n")
	fmt.Printf("  SSL Mode: %s\n", dbConfig.SSLMode)
	fmt.Printf("  Max Conns: %d\n", dbConfig.MaxConns)
	fmt.Printf("  Max Idle Conns: %d\n", dbConfig.MaxIdleConns)

	fmt.Printf("\nServer:\n")
	fmt.Printf("  Host: %s\n", serverConfig.Host)
	fmt.Printf("  Port: %d\n", serverConfig.Port)
	fmt.Printf("  Environment: %t\n", serverConfig.Environment)
	fmt.Printf("  Log Level: %s\n", serverConfig.LogLevel)
	fmt.Println("==========================")
}
