package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tovli/chatops/internal/adapters/github"
	"github.com/Tovli/chatops/internal/adapters/slack"
	"github.com/Tovli/chatops/internal/core/ports"
	"github.com/Tovli/chatops/internal/core/services"
	"github.com/Tovli/chatops/internal/infrastructure/config"
	"github.com/Tovli/chatops/internal/infrastructure/health"
	"github.com/Tovli/chatops/internal/infrastructure/router"
	"github.com/Tovli/chatops/internal/infrastructure/storage/postgres"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection", zap.Error(err))
		}
	}()

	// Initialize storage
	storage := postgres.NewPostgresStorage(db)

	// Initialize repository service with optional GitHub integration
	var githubPort ports.GitHubPort
	if cfg.GitHub.Token != "" {
		githubAdapter, err := github.NewGitHubAdapter(logger, &cfg.GitHub)
		if err != nil {
			logger.Fatal("failed to create GitHub adapter", zap.Error(err))
		}
		githubPort = githubAdapter
	}

	repoService, err := services.NewRepositoryService(services.RepositoryServiceOptions{
		Logger:     logger,
		GitHubPort: githubPort, // May be nil if GitHub is not configured
		Storage:    storage,
	})
	if err != nil {
		logger.Fatal("failed to create repository service", zap.Error(err))
	}

	// Initialize command processor
	cmdProcessor, err := services.NewCommandProcessor(logger, repoService, githubPort)
	if err != nil {
		logger.Fatal("failed to create command processor", zap.Error(err))
	}

	// Initialize Slack adapter
	slackAdapter, err := slack.NewSlackAdapter(logger, &cfg.Slack, cmdProcessor)
	if err != nil {
		logger.Fatal("failed to create Slack adapter", zap.Error(err))
	}

	// Initialize health handler
	healthHandler := health.NewHandler(logger, db)

	// Initialize router
	routerConfig := &router.Config{
		Logger:        logger,
		SlackAdapter:  slackAdapter,
		HealthHandler: healthHandler,
	}
	appRouter := router.NewRouter(routerConfig)

	// Initialize HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      appRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("starting server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown failed", zap.Error(err))
	}

	logger.Info("server stopped")
}
