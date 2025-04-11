package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/chatops/internal/adapters/github"
	"github.com/yourusername/chatops/internal/adapters/slack"
	"github.com/yourusername/chatops/internal/core/services"
	"github.com/yourusername/chatops/internal/infrastructure/config"
	"github.com/yourusername/chatops/internal/infrastructure/health"
	"github.com/yourusername/chatops/internal/infrastructure/middleware"
	"github.com/yourusername/chatops/internal/infrastructure/storage/postgres"
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
	defer db.Close()

	// Initialize storage
	storage := postgres.NewPostgresStorage(db)

	// Initialize GitHub adapter
	githubAdapter, err := github.NewGitHubAdapter(logger, &github.Config{
		Token: cfg.GitHub.Token,
	})
	if err != nil {
		logger.Fatal("failed to create GitHub adapter", zap.Error(err))
	}

	// Initialize repository service
	repoService := services.NewRepositoryService(logger, githubAdapter, storage)

	// Initialize command processor
	cmdProcessor := services.NewCommandProcessor(logger, repoService, githubAdapter)

	// Initialize Slack adapter
	slackAdapter := slack.NewSlackAdapter(logger, cfg.Slack, cmdProcessor)

	// Initialize router with middleware
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware(logger))

	// Health check endpoints
	healthHandler := health.NewHandler(logger, db)
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/health/live", healthHandler.LivenessCheck).Methods("GET")

	// API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/slack/commands", slackAdapter.HandleSlashCommand).Methods("POST")

	// Initialize HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
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
