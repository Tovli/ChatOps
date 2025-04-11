# ChatOps

A unified chat-based interface for managing internal tools and workflows, designed with multi-messenger support and open-source standards.

## Features

- Repository Management via Slack commands
- GitHub Actions workflow triggering
- Extensible architecture for multiple messaging platforms
- Comprehensive audit logging
- Health monitoring endpoints
- Role-based access control (coming soon)

## Quick Start

### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- PostgreSQL 15
- Slack App with appropriate permissions
- GitHub Personal Access Token

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/chatops.git
cd chatops
```

2. Copy the example configuration:
```bash
cp config/config.yaml.example config/config.yaml
```

3. Set up your environment variables:
```bash
export GITHUB_TOKEN=your_github_token
export SLACK_BOT_TOKEN=your_slack_bot_token
export SLACK_SIGNING_KEY=your_slack_signing_key
```

4. Start the database:
```bash
make docker-up
```

5. Run database migrations:
```bash
make migrate-up
```

6. Build and run the application:
```bash
make run
```

## Available Commands

### Slack Commands

- `/chatops manage {repositoryUrl}` - Add a repository to ChatOps
- `/chatops verify {repositoryName}` - Run default pipeline or select from available pipelines

## Documentation

- [Architecture Guide](docs/architecture.md)
- [Development Guide](docs/development.md)
- [API Reference](docs/api.md)
- [Deployment Guide](docs/deployment.md)

## Contributing

Please read our [Contributing Guide](.github/CONTRIBUTING.md) and [Code of Conduct](.github/CODE_OF_CONDUCT.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details 