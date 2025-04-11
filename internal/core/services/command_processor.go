package services

import (
	"context"
	"fmt"

	"github.com/your-project/domain"
	"github.com/your-project/ports"
	"github.com/your-project/rbac"
	"github.com/your-project/audit"
	"github.com/your-project/zap"
)

type CommandProcessor struct {
	logger      *zap.Logger
	rbac        *rbac.Service
	workflow    ports.WorkflowPort
	audit       *audit.Service
	repoService ports.RepositoryService
	githubPort  ports.GitHubPort
}

func (cp *CommandProcessor) ProcessCommand(ctx context.Context, cmd *domain.Command) (*domain.CommandResult, error) {
	switch cmd.Type {
	case domain.CommandTypeManageRepo:
		return cp.handleManageRepository(ctx, cmd)
	case domain.CommandTypeVerifyRepo:
		return cp.handleVerifyRepository(ctx, cmd)
	default:
		return nil, fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

func (cp *CommandProcessor) handleManageRepository(ctx context.Context, cmd *domain.Command) (*domain.CommandResult, error) {
	repoCmd, ok := cmd.Parameters["repository_url"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid repository URL")
	}

	repo := &domain.Repository{
		URL:     repoCmd,
		AddedBy: cmd.User.ID,
		AddedAt: cmd.Timestamp,
	}

	if err := cp.repoService.AddRepository(ctx, repo); err != nil {
		return nil, fmt.Errorf("failed to add repository: %w", err)
	}

	return &domain.CommandResult{
		Status:  "success",
		Message: fmt.Sprintf("Repository %s has been added successfully", repo.Name),
	}, nil
}

func (cp *CommandProcessor) handleVerifyRepository(ctx context.Context, cmd *domain.Command) (*domain.CommandResult, error) {
	repoName, ok := cmd.Parameters["repository_name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid repository name")
	}

	repo, err := cp.repoService.GetRepository(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	if len(repo.Pipelines) == 0 {
		return &domain.CommandResult{
			Status:  "error",
			Message: "No pipelines found for this repository",
		}, nil
	}

	var defaultPipeline *domain.Pipeline
	for _, p := range repo.Pipelines {
		if p.IsDefault {
			defaultPipeline = &p
			break
		}
	}

	if defaultPipeline == nil {
		// Return available pipelines for selection
		return &domain.CommandResult{
			Status:  "select_pipeline",
			Message: "Please select a pipeline to run",
			Details: repo.Pipelines,
		}, nil
	}

	// Trigger default pipeline
	return cp.githubPort.TriggerWorkflow(ctx, &domain.WorkflowTrigger{
		Repository: repo.Name,
		Workflow:   defaultPipeline.Path,
		Type:      "verification",
	})
} 