package github

import (
	"context"
	"strings"

	"github.com/Tovli/chatops/internal/platform/config"
	"github.com/google/go-github/v45/github"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	config *config.GitHubConfig
	client *github.Client
}

// TriggerWorkflow initiates a GitHub Actions workflow
func (s *Service) TriggerWorkflow(ctx context.Context, repo, workflow, ref string, inputs map[string]interface{}) error {
	// 1. Create GitHub client with app authentication
	// 2. Trigger workflow
	owner, name := splitRepo(repo)
	_, err := s.client.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, name, workflow, github.CreateWorkflowDispatchEventRequest{
		Ref:    ref,
		Inputs: inputs,
	})
	if err != nil {
		s.logger.Error("Failed to trigger workflow", zap.Error(err))
		return err
	}

	return nil
}

func splitRepo(fullName string) (owner, repo string) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
