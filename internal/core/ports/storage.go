package ports

import (
	"context"

	"github.com/Tovli/chatops/internal/core/domain"
)

type RepositoryStorage interface {
	AddRepository(ctx context.Context, repo *domain.Repository) error
	GetRepository(ctx context.Context, name string) (*domain.Repository, error)
	ListRepositories(ctx context.Context) ([]*domain.Repository, error)
	UpdateRepository(ctx context.Context, repo *domain.Repository) error
}
