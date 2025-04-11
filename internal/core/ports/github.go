package ports

import (
	"context"

	"github.com/Tovli/chatops/internal/core/domain"
)

// GitHubPort defines the interface for GitHub operations
type GitHubPort interface {
	// GetRepositoryDetails fetches repository details from GitHub
	GetRepositoryDetails(ctx context.Context, url string) (*domain.Repository, error)
	// TriggerWorkflow triggers a GitHub Actions workflow
	TriggerWorkflow(ctx context.Context, trigger *domain.WorkflowTrigger) (*domain.CommandResult, error)
}

// AuditService defines the interface for audit logging
type AuditService interface {
	LogAction(ctx context.Context, action string, details map[string]interface{}) error
}
