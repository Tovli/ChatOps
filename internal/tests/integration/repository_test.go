package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Tovli/chatops/internal/core/domain"
	"github.com/Tovli/chatops/internal/infrastructure/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://chatops:chatops@localhost:5432/chatops_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	// Run migrations
	// ... migration logic here ...

	return db, func() {
		db.Close()
	}
}

func TestRepositoryStorage(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	storage := postgres.NewPostgresStorage(db)

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
