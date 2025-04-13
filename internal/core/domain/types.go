package domain

import "time"

// Command represents a platform-agnostic command
type Command struct {
	ID         string
	Type       string
	Parameters map[string]interface{}
	Source     CommandSource
	User       User
	Timestamp  time.Time
}

type CommandSource struct {
	Platform   string // "slack", "teams", etc.
	ChannelID  string
	MessageID  string
	WorkflowID string // ID of the workflow if command is from a workflow
	StepID     string // ID of the workflow step if command is from a workflow
}

type User struct {
	ID          string
	Platform    string
	Permissions []string
}

type WorkflowTrigger struct {
	Type       string
	Repository string
	Workflow   string
	Parameters map[string]interface{}
}

type CommandResult struct {
	Status  string
	Message string
	Details interface{}
	Error   error
}

type Workflow struct {
	ID         string
	Name       string
	Repository string
	Path       string
	Parameters map[string]interface{}
	Status     string
	CreatedAt  time.Time
}

type WorkflowStatus struct {
	ID        string
	Status    string
	Progress  int
	Error     string
	UpdatedAt time.Time
}
