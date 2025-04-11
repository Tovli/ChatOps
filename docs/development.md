# Development Guide

## Table of Contents
1. [Getting Started](#getting-started)
2. [Development Environment](#development-environment)
3. [Project Structure](#project-structure)
4. [Development Workflow](#development-workflow)
5. [Testing](#testing)
6. [Code Style and Standards](#code-style-and-standards)
7. [Adding New Features](#adding-new-features)
8. [Debugging](#debugging)
9. [Common Development Tasks](#common-development-tasks)
10. [Troubleshooting](#troubleshooting)

## Getting Started

### Prerequisites
- Go 1.22 or later
- Docker and Docker Compose
- Git
- Visual Studio Code (recommended) or another Go-compatible IDE
- PostgreSQL 15 (for local development)

### Initial Setup

1. Clone the repository:
```bash
git clone https://github.com/Tovli/chatops.git
cd chatops
```

2. Install development tools:
```bash
# Install golang-migrate for database migrations
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install golangci-lint for code linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install mockgen for generating mocks
go install github.com/golang/mock/mockgen@latest
```

3. Set up your environment:
```bash
# Copy example configuration
cp config/config.yaml.example config/config.yaml

# Create .env file for local development
cat > .env << EOL
GITHUB_TOKEN=your_github_token
SLACK_BOT_TOKEN=your_slack_bot_token
SLACK_SIGNING_KEY=your_slack_signing_key
EOL
```

## Development Environment

### IDE Setup (VS Code)

1. Install recommended extensions:
```json
{
    "recommendations": [
        "golang.go",
        "eamodio.gitlens",
        "davidanson.vscode-markdownlint",
        "ms-azuretools.vscode-docker"
    ]
}
```

2. VS Code settings for Go development:
```json
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintFlags": [
        "--fast"
    ],
    "editor.formatOnSave": true,
    "[go]": {
        "editor.codeActionsOnSave": {
            "source.organizeImports": true
        }
    }
}
```

### Local Development Environment

1. Start PostgreSQL using Docker:

   > **Note for Windows Git Bash users**: Make sure Docker Desktop is running before executing these commands.

   ```bash
   # If using Windows Git Bash, ensure Docker Desktop is running first
   docker run --name mypostgres \
     -e POSTGRES_USER=chatops \
     -e POSTGRES_PASSWORD=chatops \
     -e POSTGRES_DB=chatops_test \
     -p 5432:5432 \
     -d postgres:latest

   # Verify the container is running
   docker ps | grep mypostgres
   ```

   If you see the error `docker: error during connect: ... dockerDesktopLinuxEngine: The system cannot find the file specified`, please:
   1. Open Docker Desktop
   2. Wait for it to fully start
   3. Try the command again

2. Run database migrations:
   ```bash
   # For the test database
   migrate -path migrations -database "postgresql://chatops:chatops@localhost:5432/chatops_test?sslmode=disable" up

   # Verify migration status
   migrate -path migrations -database "postgresql://chatops:chatops@localhost:5432/chatops_test?sslmode=disable" version
   ```

   > **Note**: If you get a "migrate: not found" error, install it first:
   > ```bash
   > go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   > ```

3. Start development services:
```bash
make docker-up
```

4. Start the application:
```bash
make run
```

## Project Structure 