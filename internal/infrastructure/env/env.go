package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Environment represents the current running environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

// Config holds environment configuration
type Config struct {
	Logger      *zap.Logger
	Environment Environment
	EnvFiles    []string
}

// LoadEnv loads environment variables from .env files with proper precedence
// Order of precedence (highest to lowest):
// 1. OS environment variables
// 2. .env.{environment}.local
// 3. .env.local (except in test)
// 4. .env.{environment}
// 5. .env
func LoadEnv(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}

	if cfg.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	// Get environment from APP_ENV, default to development
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Get project root directory for file paths
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "..", "..", "..")

	// Load files in order of precedence (lowest to highest)
	files := []string{
		filepath.Join(projectRoot, ".env"),               // The Original .env
		filepath.Join(projectRoot, ".env."+env),          // .env.{environment}
		filepath.Join(projectRoot, ".env.local"),         // .env.local (skipped in test)
		filepath.Join(projectRoot, ".env."+env+".local"), // .env.{environment}.local
	}

	// Load each file in reverse order (lowest to highest precedence)
	for _, file := range files {
		// Skip .env.local in test environment
		if strings.HasSuffix(file, ".env.local") && env == "test" {
			cfg.Logger.Debug("skipping .env.local in test environment", zap.String("file", file))
			continue
		}

		if err := loadEnvFile(file, cfg.Logger); err != nil {
			// Log error but continue - missing env files are often expected
			cfg.Logger.Debug("failed to load env file",
				zap.String("file", file),
				zap.Error(err))
		}
	}

	return nil
}

// loadEnvFile loads a single env file
func loadEnvFile(filename string, logger *zap.Logger) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // File doesn't exist, skip silently
	}

	err := godotenv.Load(filename)
	if err != nil {
		return fmt.Errorf("error loading env file %s: %w", filename, err)
	}

	logger.Debug("loaded env file", zap.String("file", filename))
	return nil
}

// GetEnvWithDefault gets an environment variable with a default value
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RequireEnv gets an environment variable and panics if it's not set
func RequireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}
