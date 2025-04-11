# Contributing to ChatOps

## Getting Started
1. Fork the repository
2. Clone your fork
3. Create a new branch
4. Make your changes
5. Write or update tests
6. Update documentation
7. Submit a pull request

## Development Setup
1. Install Go 1.22 or later
2. Install required tools:
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```
3. Copy config.yaml.example to config.yaml and fill in your values

## Testing
- Run tests: `go test ./...`
- Run linter: `golangci-lint run`

## Commit Messages
We follow the Conventional Commits specification:
- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- style: Code style changes
- refactor: Code refactoring
- test: Test updates
- chore: Maintenance tasks

[More detailed content can be provided if needed] 