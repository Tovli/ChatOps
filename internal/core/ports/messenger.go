package ports

import (
	"context"
	"net/http"

	"github.com/your-project/domain"
)

// MessengerPort defines the interface for any messaging platform
type MessengerPort interface {
	ValidateRequest(ctx context.Context, request *http.Request) error
	ParseCommand(ctx context.Context, request *http.Request) (*domain.Command, error)
	SendResponse(ctx context.Context, response *domain.CommandResult) error
} 