package ports

// WorkflowPort defines the interface for workflow engines
type WorkflowPort interface {
    TriggerWorkflow(ctx context.Context, trigger *domain.WorkflowTrigger) (*domain.CommandResult, error)
    GetWorkflowStatus(ctx context.Context, id string) (*domain.CommandResult, error)
} 