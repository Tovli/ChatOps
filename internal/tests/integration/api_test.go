package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Tovli/chatops/internal/adapters/github"
	"github.com/Tovli/chatops/internal/adapters/slack"
	"github.com/Tovli/chatops/internal/core/services"
	"github.com/Tovli/chatops/internal/infrastructure/config"
	"github.com/Tovli/chatops/internal/infrastructure/router"
	storagepg "github.com/Tovli/chatops/internal/infrastructure/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testServer struct {
	router     *router.Config
	httpServer *httptest.Server
	storage    *storagepg.PostgresStorage
	cleanup    func()
}

func setupTestServer(t *testing.T) *testServer {
	// Initialize logger
	logger := zap.NewNop()

	// Setup test database
	db, cleanup := setupTestDB(t)
	storage := storagepg.NewPostgresStorage(db)

	// Get GitHub token from environment
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		t.Skip("GITHUB_TOKEN environment variable not set")
	}

	// Initialize GitHub adapter with test configuration
	githubConfig := &config.GitHubConfig{
		Token: githubToken,
	}
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

	// Initialize Slack adapter
	slackConfig := &config.SlackConfig{
		BotToken:   "test-bot-token",
		SigningKey: "test-signing-key",
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
