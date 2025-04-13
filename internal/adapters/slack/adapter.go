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
	w.Header().Set("Content-Type", "application/json")

	cmd, err := slack.SlashCommandParse(r)
	if err != nil {
		a.sendErrorResponse(w, "Invalid slash command", http.StatusBadRequest)
		return
	}

	// Verify the command token matches our signing key
	if cmd.Token != a.config.SigningKey {
		a.sendErrorResponse(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	domainCmd, err := a.parseCommand(cmd)
	if err != nil {
		a.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := a.processor.ProcessCommand(r.Context(), domainCmd)
	if err != nil {
		a.sendErrorResponse(w, fmt.Sprintf("Failed to process command: %v", err), http.StatusInternalServerError)
		return
	}

	response := a.buildSlackResponse(result)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (a *SlackAdapter) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"status":  "error",
		"message": message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.logger.Error("failed to encode error response", zap.Error(err))
	}
}

func (a *SlackAdapter) parseCommand(cmd slack.SlashCommand) (*domain.Command, error) {
	parts := strings.Fields(cmd.Text)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid command format: expected at least 2 parts, got %d", len(parts))
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
