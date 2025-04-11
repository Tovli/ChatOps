package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tovli/chatops/internal/core/domain"
	"github.com/Tovli/chatops/internal/core/services"
	"github.com/Tovli/chatops/internal/infrastructure/config"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

type SlackAdapter struct {
	logger    *zap.Logger
	config    *config.SlackConfig
	processor *services.CommandProcessor
	client    *slack.Client
}

// NewSlackAdapter creates a new instance of SlackAdapter
func NewSlackAdapter(logger *zap.Logger, config *config.SlackConfig, processor *services.CommandProcessor) (*SlackAdapter, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if processor == nil {
		return nil, fmt.Errorf("command processor is required")
	}
	if config.BotToken == "" {
		return nil, fmt.Errorf("slack bot token is required")
	}

	client := slack.New(config.BotToken)

	return &SlackAdapter{
		logger:    logger,
		config:    config,
		processor: processor,
		client:    client,
	}, nil
}

func (a *SlackAdapter) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	cmd, err := slack.SlashCommandParse(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	domainCmd, err := a.parseCommand(cmd)
	if err != nil {
		// Return error message to Slack
		return
	}

	result, err := a.processor.ProcessCommand(r.Context(), domainCmd)
	if err != nil {
		// Handle error
		return
	}

	response := a.buildSlackResponse(result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *SlackAdapter) parseCommand(cmd slack.SlashCommand) (*domain.Command, error) {
	parts := strings.Fields(cmd.Text)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid command format")
	}

	action := parts[0]
	switch action {
	case "manage":
		return &domain.Command{
			Type: domain.CommandTypeManageRepo,
			Parameters: map[string]interface{}{
				"repository_url": parts[1],
			},
			User: domain.User{
				ID:       cmd.UserID,
				Platform: "slack",
			},
			Source: domain.CommandSource{
				Platform:  "slack",
				ChannelID: cmd.ChannelID,
			},
			Timestamp: time.Now(),
		}, nil
	case "verify":
		return &domain.Command{
			Type: domain.CommandTypeVerifyRepo,
			Parameters: map[string]interface{}{
				"repository_name": parts[1],
			},
			User: domain.User{
				ID:       cmd.UserID,
				Platform: "slack",
			},
			Source: domain.CommandSource{
				Platform:  "slack",
				ChannelID: cmd.ChannelID,
			},
			Timestamp: time.Now(),
		}, nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (a *SlackAdapter) buildSlackResponse(result *domain.CommandResult) map[string]interface{} {
	response := map[string]interface{}{
		"status":  result.Status,
		"message": result.Message,
	}

	if result.Status == "select_pipeline" {
		response["details"] = result.Details
	}

	return response
}
