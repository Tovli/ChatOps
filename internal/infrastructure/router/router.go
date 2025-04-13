package router

import (
	"github.com/Tovli/chatops/internal/adapters/slack"
	"github.com/Tovli/chatops/internal/infrastructure/health"
	"github.com/Tovli/chatops/internal/infrastructure/middleware"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Config holds the configuration for router dependencies
type Config struct {
	Logger        *zap.Logger
	SlackAdapter  *slack.SlackAdapter
	HealthHandler *health.Handler
}

// NewRouter creates and configures a new router with all application routes
func NewRouter(cfg *Config) *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.LoggingMiddleware(cfg.Logger))

	// Health check endpoints
	router.HandleFunc("/health", cfg.HealthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/health/live", cfg.HealthHandler.LivenessCheck).Methods("GET")

	// API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/slack/commands", cfg.SlackAdapter.HandleSlashCommand).Methods("POST")

	return router
}
