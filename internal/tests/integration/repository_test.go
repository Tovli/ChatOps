package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Tovli/chatops/internal/core/domain"
	storagepg "github.com/Tovli/chatops/internal/infrastructure/storage/postgres"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateUniqueDatabaseName() string {
	return fmt.Sprintf("chatops_test_%d", time.Now().UnixNano())
}

func terminateConnections(db *sql.DB, dbName string) error {
	query := fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
		AND pid <> pg_backend_pid()`, dbName)

	_, err := db.Exec(query)
	return err
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	dbName := generateUniqueDatabaseName()

	// Build connection string from environment variables
	host := getEnvOrDefault("CHATOPS_DB_HOST", "localhost")
	port := getEnvOrDefault("CHATOPS_DB_PORT", "5432")
	user := getEnvOrDefault("CHATOPS_DB_USER", "chatops")
	password := getEnvOrDefault("CHATOPS_DB_PASSWORD", "chatops")
	sslmode := getEnvOrDefault("CHATOPS_DB_SSLMODE", "disable")

	// Connection string for the postgres database
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/postgres?sslmode=%s",
		user, password, host, port, sslmode)

	// First connect to the postgres database to create our test database
	rootDB, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer rootDB.Close()

	// Ensure the connection works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, rootDB.PingContext(ctx))

	// Drop the test database if it exists and create it fresh
	_, err = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	require.NoError(t, err)
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err)

	// Now connect to our test database
	testDBURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslmode)
	db, err := sql.Open("postgres", testDBURL)
	require.NoError(t, err)

	// Ensure the connection works
	require.NoError(t, db.PingContext(ctx))

	// Run migrations
	workDir, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Join(filepath.Dir(workDir), "..", "..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	// Read and execute migrations manually
	files, err := os.ReadDir(migrationsPath)
	require.NoError(t, err)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			content, err := os.ReadFile(filepath.Join(migrationsPath, file.Name()))
			require.NoError(t, err)

			_, err = db.Exec(string(content))
			require.NoError(t, err, "Failed to execute migration: "+file.Name())
		}
	}

	return db, func() {
		db.Close()

		// Connect back to postgres database to clean up
		rootDB, err := sql.Open("postgres", dbURL)
		if err != nil {
			t.Logf("Warning: Failed to connect to postgres database for cleanup: %v", err)
			return
		}
		defer rootDB.Close()

		// Terminate all connections to the test database
		err = terminateConnections(rootDB, dbName)
		if err != nil {
			t.Logf("Warning: Failed to terminate connections to test database: %v", err)
		}

		// Drop the test database with retry logic
		for i := 0; i < 3; i++ {
			_, err = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
			if err == nil {
				break
			}
			time.Sleep(time.Second) // Wait before retrying
		}
		if err != nil {
			t.Logf("Warning: Failed to drop test database after retries: %v", err)
		}
	}
}

func TestRepositoryStorage(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := storagepg.NewPostgresStorage(db)

	t.Run("AddAndGetRepository", func(t *testing.T) {
		repo := &domain.Repository{
			Name:          "test-repo",
			URL:           "https://github.com/test/test-repo",
			DefaultBranch: "main",
			AddedBy:       "user123",
			AddedAt:       time.Now(),
			Pipelines: []domain.Pipeline{
				{
					Name:      "CI",
					Path:      ".github/workflows/ci.yml",
					IsDefault: true,
				},
			},
		}

		err := storage.AddRepository(context.Background(), repo)
		assert.NoError(t, err)

		fetched, err := storage.GetRepository(context.Background(), repo.Name)
		assert.NoError(t, err)
		assert.Equal(t, repo.Name, fetched.Name)
		assert.Equal(t, repo.URL, fetched.URL)
		assert.Len(t, fetched.Pipelines, 1)
	})
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
