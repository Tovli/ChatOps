package commands

import (
	"errors"
	"strings"
	"time"
)

type Command struct {
	Action string            // e.g., "deploy"
	Target string            // e.g., "production"
	Params map[string]string // Additional parameters
}

type CommandResult struct {
	Status  string
	Message string
	Data    interface{}
}

type Repository struct {
	Name          string
	URL           string
	DefaultBranch string
	Pipelines     []Pipeline
	AddedBy       string
	AddedAt       time.Time
}

type Pipeline struct {
	Name      string
	Path      string
	IsDefault bool
}

func ParseCommand(input string) (*Command, error) {
	if input == "" {
		return nil, errors.New("input cannot be empty")
	}

	parts := strings.Fields(input)
	if len(parts) < 2 {
		return nil, errors.New("input must contain at least an action and a target")
	}

	command := &Command{
		Action: parts[0],
		Target: parts[1],
		Params: make(map[string]string),
	}

	for _, part := range parts[2:] {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			return nil, errors.New("invalid parameter format, expected key=value")
		}
		command.Params[keyValue[0]] = keyValue[1]
	}

	return command, nil
}
