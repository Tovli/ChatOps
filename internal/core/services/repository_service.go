package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tovli/chatops/internal/core/domain"
	"github.com/Tovli/chatops/internal/core/ports"
	"go.uber.org/zap"
)

type repositoryService struct {
	logger     *zap.Logger
	githubPort ports.GitHubPort // Optional: only needed for GitHub repositories
	storage    ports.RepositoryStorage
}

type RepositoryServiceOptions struct {
	Logger     *zap.Logger
	GitHubPort ports.GitHubPort // Optional
	Storage    ports.RepositoryStorage
}

func NewRepositoryService(opts RepositoryServiceOptions) (ports.RepositoryService, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if opts.Storage == nil {
		return nil, fmt.Errorf("storage is required")
	}

	return &repositoryService{
		logger:     opts.Logger,
		githubPort: opts.GitHubPort, // May be nil if GitHub integration is not needed
		storage:    opts.Storage,
	}, nil
}

func (s *repositoryService) AddRepository(ctx context.Context, repo *domain.Repository) error {
	// If it's a GitHub repository and we have GitHub integration
	if isGitHubURL(repo.URL) {
		if s.githubPort == nil {
			return fmt.Errorf("GitHub integration is not configured")
		}

		// Fetch repository details from GitHub
		details, err := s.githubPort.GetRepositoryDetails(ctx, repo.URL)
		if err != nil {
			return fmt.Errorf("failed to fetch GitHub repository details: %w", err)
		}

		// Update repository with GitHub details
		repo.Name = details.Name
		repo.DefaultBranch = details.DefaultBranch
		repo.Pipelines = details.Pipelines
	}

	// Store repository information
	return s.storage.AddRepository(ctx, repo)
}

func (s *repositoryService) GetRepository(ctx context.Context, name string) (*domain.Repository, error) {
	return s.storage.GetRepository(ctx, name)
}

// isGitHubURL checks if the given URL is a GitHub repository URL
func isGitHubURL(url string) bool {
	return strings.Contains(url, "github.com")
}

func (s *repositoryService) GetRepositoryPipelines(ctx context.Context, name string) ([]domain.Pipeline, error) {
	repo, err := s.GetRepository(ctx, name)
	if err != nil {
		return nil, err
	}
	return repo.Pipelines, nil
}

func (s *repositoryService) SetDefaultPipeline(ctx context.Context, repoName, pipelineName string) error {
	repo, err := s.GetRepository(ctx, repoName)
	if err != nil {
		return err
	}

	found := false
	for i := range repo.Pipelines {
		if repo.Pipelines[i].Name == pipelineName {
			repo.Pipelines[i].IsDefault = true
			found = true
		} else {
			repo.Pipelines[i].IsDefault = false
		}
	}

	if !found {
		return fmt.Errorf("pipeline %s not found in repository %s", pipelineName, repoName)
	}

	return s.storage.UpdateRepository(ctx, repo)
}

func (s *repositoryService) ListRepositories(ctx context.Context) ([]*domain.Repository, error) {
	return s.storage.ListRepositories(ctx)
}

// ... implement other interface methods
