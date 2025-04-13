package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"crypto/hmac"
	"crypto/sha256"

	"github.com/Tovli/chatops/internal/adapters/github"
	"github.com/Tovli/chatops/internal/adapters/slack"
	"github.com/Tovli/chatops/internal/core/services"
	"github.com/Tovli/chatops/internal/infrastructure/config"
	"github.com/Tovli/chatops/internal/infrastructure/env"
	"github.com/Tovli/chatops/internal/infrastructure/router"
	storagepg "github.com/Tovli/chatops/internal/infrastructure/storage/postgres"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	// Set test environment
	if env.GetEnvWithDefault("APP_ENV", "") == "" {
		os.Setenv("APP_ENV", "test")
	}

	// Load environment variables in correct order
	environment := env.GetEnvWithDefault("APP_ENV", "test")
	godotenv.Load(".env." + environment + ".local") // .env.test.local
	if environment != "test" {
		godotenv.Load(".env.local") // Skip in test environment
	}
	godotenv.Load(".env." + environment) // .env.test
	godotenv.Load()                      // The Original .env

	// Set default test values if not provided
	setDefaultTestEnv("CHATOPS_GITHUB_TOKEN", "test_github_token")
	setDefaultTestEnv("CHATOPS_SLACK_BOT_TOKEN", "test_slack_bot_token")
	setDefaultTestEnv("CHATOPS_SLACK_SIGNING_KEY", "test_slack_signing_key")
}

// setDefaultTestEnv sets a default value for an environment variable if it's not already set
func setDefaultTestEnv(key, defaultValue string) {
	if env.GetEnvWithDefault(key, "") == "" {
		os.Setenv(key, defaultValue)
	}
}

type testServer struct {
	router     *router.Config
	httpServer *httptest.Server
	storage    *storagepg.PostgresStorage
	cleanup    func()
}

func setupTestServer(t *testing.T) *testServer {
	// Initialize logger with test configuration
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := loggerConfig.Build()
	require.NoError(t, err)

	// Setup test database
	db, cleanup := setupTestDB(t)
	storage := storagepg.NewPostgresStorage(db)

	// Get GitHub token from environment (will use test token if not set)
	githubToken := env.GetEnvWithDefault("CHATOPS_GITHUB_TOKEN", "")
	githubConfig := &config.GitHubConfig{
		Token: githubToken,
	}

	// Initialize GitHub adapter with test configuration
	githubAdapter, err := github.NewGitHubAdapter(logger, githubConfig)
	require.NoError(t, err)

	// Initialize repository service
	repoService, err := services.NewRepositoryService(services.RepositoryServiceOptions{
		Logger:     logger,
		GitHubPort: githubAdapter,
		Storage:    storage,
	})
	require.NoError(t, err)

	// Initialize command processor
	cmdProcessor, err := services.NewCommandProcessor(logger, repoService, githubAdapter)
	require.NoError(t, err)

	// Initialize Slack adapter with test configuration
	slackConfig := &config.SlackConfig{
		BotToken:   env.GetEnvWithDefault("CHATOPS_SLACK_BOT_TOKEN", ""),
		SigningKey: env.GetEnvWithDefault("CHATOPS_SLACK_SIGNING_KEY", ""),
	}
	slackAdapter, err := slack.NewSlackAdapter(logger, slackConfig, cmdProcessor)
	require.NoError(t, err)

	// Initialize router
	routerConfig := &router.Config{
		Logger:        logger,
		SlackAdapter:  slackAdapter,
		HealthHandler: nil, // Not needed for this test
	}

	return &testServer{
		router:  routerConfig,
		storage: storage,
		cleanup: cleanup,
	}
}

func TestSlackCommandsEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.cleanup()

	appRouter := router.NewRouter(server.router)
	testServer := httptest.NewServer(appRouter)
	defer testServer.Close()

	t.Run("Add GitHub Repository Command", func(t *testing.T) {
		// Use a real public GitHub repository for testing
		repoURL := "https://github.com/Tovli/ChatOps"

		// Prepare the Slack slash command payload
		form := url.Values{}
		form.Add("token", "test-signing-key") // Using signing key as verification token
		form.Add("team_id", "T123456")
		form.Add("team_domain", "test-team")
		form.Add("channel_id", "C123456")
		form.Add("channel_name", "test-channel")
		form.Add("user_id", "U123456")
		form.Add("user_name", "testuser")
		form.Add("command", "/chatops")
		form.Add("text", "manage "+repoURL)
		form.Add("response_url", "https://hooks.slack.com/commands/123456")
		form.Add("trigger_id", "123456.123456")

		// Send request to the endpoint
		req, err := http.NewRequest(
			"POST",
			testServer.URL+"/api/v1/slack/commands",
			strings.NewReader(form.Encode()),
		)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Perform the request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Decode response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// Verify response contains expected fields
		assert.Contains(t, response, "status")
		assert.Contains(t, response, "message")

		// Verify repository was added to storage
		repo, err := server.storage.GetRepository(context.Background(), "ChatOps")
		require.NoError(t, err)
		assert.Equal(t, repoURL, repo.URL)
		assert.Equal(t, "U123456", repo.AddedBy)
	})

	t.Run("Invalid Command Format", func(t *testing.T) {
		form := url.Values{}
		form.Add("token", "test-signing-key") // Using signing key as verification token
		form.Add("team_id", "T123456")
		form.Add("user_id", "U123456")
		form.Add("command", "/chatops")
		form.Add("text", "invalid") // Invalid command format

		req, err := http.NewRequest(
			"POST",
			testServer.URL+"/api/v1/slack/commands",
			strings.NewReader(form.Encode()),
		)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Decode error response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// Verify error response format
		assert.Equal(t, "error", response["status"])
		assert.Contains(t, response["message"], "invalid command format")
	})

	t.Run("Invalid Token", func(t *testing.T) {
		form := url.Values{}
		form.Add("token", "wrong-token")
		form.Add("team_id", "T123456")
		form.Add("user_id", "U123456")
		form.Add("command", "/chatops")
		form.Add("text", "manage https://github.com/test/test-repo")

		req, err := http.NewRequest(
			"POST",
			testServer.URL+"/api/v1/slack/commands",
			strings.NewReader(form.Encode()),
		)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// Decode error response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// Verify error response format
		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "Invalid token", response["message"])
	})
}

func TestSlackWebhookEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.cleanup()

	appRouter := router.NewRouter(server.router)
	testServer := httptest.NewServer(appRouter)
	defer testServer.Close()

	t.Run("Valid Webhook Request", func(t *testing.T) {
		// Prepare webhook payload
		payload := map[string]interface{}{
			"type": "workflow_step_execute",
			"event": map[string]interface{}{
				"workflow_step": map[string]interface{}{
					"workflow_id": "W123456",
					"step_id":     "S123456",
				},
				"inputs": map[string]interface{}{
					"repository": "Tovli/ChatOps",
					"action":     "verify",
				},
			},
			"team_id":    "T123456",
			"api_app_id": "A123456",
			"token":      env.GetEnvWithDefault("CHATOPS_SLACK_SIGNING_KEY", ""),
		}

		// Convert payload to JSON
		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		// Create timestamp for signature
		timestamp := fmt.Sprintf("%d", time.Now().Unix())

		// Create Slack signature
		signingSecret := env.GetEnvWithDefault("CHATOPS_SLACK_SIGNING_KEY", "")
		baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(payloadBytes))
		mac := hmac.New(sha256.New, []byte(signingSecret))
		mac.Write([]byte(baseString))
		signature := fmt.Sprintf("v0=%x", mac.Sum(nil))

		// Create request
		req, err := http.NewRequest(
			"POST",
			testServer.URL+"/api/v1/slack/webhooks",
			bytes.NewReader(payloadBytes),
		)
		require.NoError(t, err)

		// Add required headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Slack-Request-Timestamp", timestamp)
		req.Header.Set("X-Slack-Signature", signature)

		// Perform request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Contains(t, response["message"], "Webhook processed successfully")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		payload := map[string]interface{}{
			"type": "workflow_step_execute",
			"event": map[string]interface{}{
				"workflow_step": map[string]interface{}{
					"workflow_id": "W123456",
					"step_id":     "S123456",
				},
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest(
			"POST",
			testServer.URL+"/api/v1/slack/webhooks",
			bytes.NewReader(payloadBytes),
		)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Slack-Request-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
		req.Header.Set("X-Slack-Signature", "invalid_signature")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "Invalid signature", response["message"])
	})

	t.Run("Expired Timestamp", func(t *testing.T) {
		payload := map[string]interface{}{
			"type": "workflow_step_execute",
			"event": map[string]interface{}{
				"workflow_step": map[string]interface{}{
					"workflow_id": "W123456",
					"step_id":     "S123456",
				},
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		// Use a timestamp from 6 minutes ago (Slack requires within 5 minutes)
		oldTimestamp := fmt.Sprintf("%d", time.Now().Add(-6*time.Minute).Unix())

		// Create signature with old timestamp
		signingSecret := env.GetEnvWithDefault("CHATOPS_SLACK_SIGNING_KEY", "")
		baseString := fmt.Sprintf("v0:%s:%s", oldTimestamp, string(payloadBytes))
		mac := hmac.New(sha256.New, []byte(signingSecret))
		mac.Write([]byte(baseString))
		signature := fmt.Sprintf("v0=%x", mac.Sum(nil))

		req, err := http.NewRequest(
			"POST",
			testServer.URL+"/api/v1/slack/webhooks",
			bytes.NewReader(payloadBytes),
		)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Slack-Request-Timestamp", oldTimestamp)
		req.Header.Set("X-Slack-Signature", signature)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "Request timestamp is too old", response["message"])
	})
}
