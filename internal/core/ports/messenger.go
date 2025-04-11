package ports

import (
	"context"
	"net/http"

	"github.com/Tovli/chatops/internal/core/commands"
)

// MessengerPort defines the interface for any messaging platform
type MessengerPort interface {
	ValidateRequest(ctx context.Context, request *http.Request) error
	ParseCommand(ctx context.Context, request *http.Request) (*commands.Command, error)
	SendResponse(ctx context.Context, response *commands.CommandResult) error
}
