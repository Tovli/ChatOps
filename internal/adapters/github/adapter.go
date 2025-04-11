package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tovli/chatops/internal/core/domain"
	"github.com/Tovli/chatops/internal/infrastructure/config"
	"github.com/google/go-github/v45/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type GitHubAdapter struct {
	logger *zap.Logger
	config *config.GitHubConfig
	client *github.Client
}

func NewGitHubAdapter(logger *zap.Logger, config *config.GitHubConfig) (*GitHubAdapter, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &GitHubAdapter{
		logger: logger,
		client: client,
		config: config,
	}, nil
}

func (a *GitHubAdapter) GetRepositoryDetails(ctx context.Context, url string) (*domain.Repository, error) {
	owner, repo := parseGitHubURL(url)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	repository, _, err := a.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	pipelines, err := a.getRepositoryWorkflows(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workflows: %w", err)
	}

	return &domain.Repository{
		Name:          repository.GetName(),
		URL:           repository.GetHTMLURL(),
		DefaultBranch: repository.GetDefaultBranch(),
		Pipelines:     pipelines,
	}, nil
}

func (a *GitHubAdapter) getRepositoryWorkflows(ctx context.Context, owner, repo string) ([]domain.Pipeline, error) {
	workflows, _, err := a.client.Actions.ListWorkflows(ctx, owner, repo, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	var pipelines []domain.Pipeline
	for _, workflow := range workflows.Workflows {
		pipelines = append(pipelines, domain.Pipeline{
			Name: workflow.GetName(),
			Path: workflow.GetPath(),
		})
	}

	return pipelines, nil
}

func (a *GitHubAdapter) TriggerWorkflow(ctx context.Context, trigger *domain.WorkflowTrigger) (*domain.CommandResult, error) {
	owner, repo := parseGitHubURL(trigger.Repository)

	resp, err := a.client.Actions.CreateWorkflowDispatchEventByFileName(
		ctx,
		owner,
		repo,
		trigger.Workflow,
		github.CreateWorkflowDispatchEventRequest{
			Ref:    "main",
			Inputs: trigger.Parameters,
		},
	)

	if err != nil {
		return &domain.CommandResult{
			Status:  "error",
			Message: fmt.Sprintf("Failed to trigger workflow: %v", err),
		}, nil
	}

	if resp.StatusCode >= 400 {
		return &domain.CommandResult{
			Status:  "error",
			Message: fmt.Sprintf("Failed to trigger workflow: HTTP %d", resp.StatusCode),
		}, nil
	}

	return &domain.CommandResult{
		Status:  "success",
		Message: "Workflow triggered successfully",
	}, nil
}

func parseGitHubURL(url string) (owner, repo string) {
	// Handle both HTTPS and SSH URLs
	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[len(parts)-2], parts[len(parts)-1]
}
