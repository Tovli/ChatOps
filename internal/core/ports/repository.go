package ports

import (
	"context"
	"github.com/yourusername/chatops/internal/core/domain"
)

type RepositoryService interface {
	AddRepository(ctx context.Context, repo *domain.Repository) error
	GetRepository(ctx context.Context, name string) (*domain.Repository, error)
	ListRepositories(ctx context.Context) ([]*domain.Repository, error)
	GetRepositoryPipelines(ctx context.Context, name string) ([]domain.Pipeline, error)
	SetDefaultPipeline(ctx context.Context, repoName, pipelineName string) error
} 