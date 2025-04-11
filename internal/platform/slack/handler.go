package slack

import (
	"github.com/Tovli/chatops/internal/infrastructure/config"
	"github.com/Tovli/chatops/internal/platform/github"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

type CommandHandler struct {
	logger    *zap.Logger
	config    *config.SlackConfig
	githubSvc *github.Service
}

type Response struct {
	Text         string
	ResponseType string
}

// HandleCommand processes incoming Slack slash commands
func (h *CommandHandler) HandleCommand(cmd *slack.SlashCommand) (*Response, error) {
	// 1. Validate Slack signature
	// 2. Parse command and arguments
	// 3. Trigger appropriate GitHub workflow
	// 4. Return immediate acknowledgment
	return &Response{
		Text:         "Command received",
		ResponseType: "ephemeral",
	}, nil
}
