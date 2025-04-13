package messenger

import "context"

// Message represents a generic message from any chat platform
type Message struct {
	ID        string
	Platform  string
	ChannelID string
	UserID    string
	Content   string
	RawData   interface{}
}

// Response represents a generic response to be sent back
type Response struct {
	Content     string
	Attachments []Attachment
}

// Attachment represents additional content for responses
type Attachment struct {
	Type    string
	Content interface{}
}

// Service defines the interface that all messenger implementations must satisfy
type Service interface {
	// Initialize sets up the messenger service
	Initialize(ctx context.Context) error

	// HandleMessage processes incoming messages
	HandleMessage(ctx context.Context, msg Message) error

	// SendResponse sends a response back to the platform
	SendResponse(ctx context.Context, channelID string, response Response) error
}
