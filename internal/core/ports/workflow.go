package ports

import (
	"context"

	"github.com/Tovli/chatops/internal/core/domain"
)

// WorkflowPort defines the interface for workflow operations
type WorkflowPort interface {
	ExecuteWorkflow(ctx context.Context, workflow *domain.Workflow) error
	GetWorkflowStatus(ctx context.Context, workflowID string) (*domain.WorkflowStatus, error)
}
