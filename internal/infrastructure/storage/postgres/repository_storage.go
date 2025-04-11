package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/yourusername/chatops/internal/core/domain"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) AddRepository(ctx context.Context, repo *domain.Repository) error {
	pipelines, err := json.Marshal(repo.Pipelines)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO repositories (name, url, default_branch, added_by, added_at, pipelines)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = s.db.ExecContext(ctx, query,
		repo.Name,
		repo.URL,
		repo.DefaultBranch,
		repo.AddedBy,
		repo.AddedAt,
		pipelines,
	)

	return err
}

func (s *PostgresStorage) GetRepository(ctx context.Context, name string) (*domain.Repository, error) {
	query := `
		SELECT name, url, default_branch, added_by, added_at, pipelines
		FROM repositories
		WHERE name = $1
	`

	var repo domain.Repository
	var pipelinesJSON []byte

	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&repo.Name,
		&repo.URL,
		&repo.DefaultBranch,
		&repo.AddedBy,
		&repo.AddedAt,
		&pipelinesJSON,
	)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(pipelinesJSON, &repo.Pipelines)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}
