package mocks

import (
	"context"

	"github.com/Tovli/chatops/internal/core/domain"
)

// MockGitHubAdapter is a mock implementation of the GitHubPort interface for testing
type MockGitHubAdapter struct {
	GetRepositoryDetailsFn func(ctx context.Context, url string) (*domain.Repository, error)
	TriggerWorkflowFn      func(ctx context.Context, trigger *domain.WorkflowTrigger) (*domain.CommandResult, error)
}

func (m *MockGitHubAdapter) GetRepositoryDetails(ctx context.Context, url string) (*domain.Repository, error) {
	if m.GetRepositoryDetailsFn != nil {
		return m.GetRepositoryDetailsFn(ctx, url)
	}
	return &domain.Repository{
		ID:            "mock-repo-id",
		Name:          "ChatOps",
		URL:           url,
		DefaultBranch: "main",
		Pipelines:     []domain.Pipeline{},
	}, nil
}

func (m *MockGitHubAdapter) TriggerWorkflow(ctx context.Context, trigger *domain.WorkflowTrigger) (*domain.CommandResult, error) {
	if m.TriggerWorkflowFn != nil {
		return m.TriggerWorkflowFn(ctx, trigger)
	}
	return &domain.CommandResult{
		Status:  "success",
		Message: "Workflow triggered successfully",
	}, nil
}
