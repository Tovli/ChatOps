package domain

import "time"

type Repository struct {
	ID            string
	Name          string    // Extracted from URL
	URL           string    // Full repository URL
	DefaultBranch string    // Usually 'main' or 'master'
	AddedBy       string    // User ID who added the repo
	AddedAt       time.Time
	Pipelines     []Pipeline
}

type Pipeline struct {
	Name     string
	Path     string // Path to the workflow file
	IsDefault bool
} 