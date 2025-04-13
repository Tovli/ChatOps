package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha256"

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

// verifySlackSignature verifies the request signature from Slack
func (a *SlackAdapter) verifySlackSignature(r *http.Request, body []byte) error {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	slackSignature := r.Header.Get("X-Slack-Signature")

	// Verify timestamp is within 5 minutes
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}

	now := time.Now().Unix()
	if math.Abs(float64(now-ts)) > 300 {
		return fmt.Errorf("request timestamp is too old")
	}

	// Create signature
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(a.config.SigningKey))
	mac.Write([]byte(baseString))
	expectedSignature := fmt.Sprintf("v0=%x", mac.Sum(nil))

	if !hmac.Equal([]byte(expectedSignature), []byte(slackSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// HandleWebhook processes incoming Slack webhooks
func (a *SlackAdapter) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.sendErrorResponse(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Replace the body for further reading

	// Verify signature
	if err := a.verifySlackSignature(r, body); err != nil {
		a.sendErrorResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse webhook payload
	var payload struct {
		Type      string                 `json:"type"`
		Event     map[string]interface{} `json:"event"`
		TeamID    string                 `json:"team_id"`
		APIAppID  string                 `json:"api_app_id"`
		Token     string                 `json:"token"`
		Challenge string                 `json:"challenge"` // For URL verification
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		a.sendErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Handle URL verification challenge
	if payload.Type == "url_verification" {
		response := map[string]string{"challenge": payload.Challenge}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			a.logger.Error("failed to encode challenge response", zap.Error(err))
		}
		return
	}

	// Process webhook event
	result, err := a.processWebhookEvent(&payload)
	if err != nil {
		a.sendErrorResponse(w, fmt.Sprintf("Failed to process webhook: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	response := map[string]interface{}{
		"status":  "success",
		"message": "Webhook processed successfully",
		"details": result,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.logger.Error("failed to encode webhook response", zap.Error(err))
	}
}

func (a *SlackAdapter) processWebhookEvent(payload *struct {
	Type      string                 `json:"type"`
	Event     map[string]interface{} `json:"event"`
	TeamID    string                 `json:"team_id"`
	APIAppID  string                 `json:"api_app_id"`
	Token     string                 `json:"token"`
	Challenge string                 `json:"challenge"`
}) (*domain.CommandResult, error) {
	// Extract workflow step if present
	if payload.Type == "workflow_step_execute" {
		if event, ok := payload.Event["workflow_step"].(map[string]interface{}); ok {
			inputs, ok := event["inputs"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid workflow step inputs")
			}

			// Convert workflow step to domain command
			cmd := &domain.Command{
				Type:       domain.CommandTypeVerifyRepo,
				Parameters: inputs,
				Source: domain.CommandSource{
					Platform:  "slack",
					ChannelID: payload.TeamID, // Using team ID as channel for workflow steps
				},
				User: domain.User{
					ID:       payload.APIAppID,
					Platform: "slack",
				},
				Timestamp: time.Now(),
			}

			// Process the command
			return a.processor.ProcessCommand(context.Background(), cmd)
		}
	}

	return &domain.CommandResult{
		Status:  "success",
		Message: "Event acknowledged",
	}, nil
}
