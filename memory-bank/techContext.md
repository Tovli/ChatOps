# Technology Stack & Development Context

## Core Technologies

### Backend
- Language: Go 1.22+
  - Rationale: High performance, excellent concurrency, strong typing, great for cloud-native apps
- Framework: 
  - Fiber (Web Framework)
  - Wire (Dependency Injection)
  - Zap (Logging)
  - Viper (Configuration)

### Infrastructure
- Container Orchestration: Kubernetes
- Cloud Provider: Azure
  - Azure Kubernetes Service (AKS)
  - Azure Container Registry
  - Azure Key Vault
- Infrastructure as Code: Terraform
- CI/CD: GitHub Actions

### Data Storage
- Primary Database: PostgreSQL
  - Audit logs, configuration, user data
  - Using sqlc for type-safe SQL
- Cache Layer: Redis
  - Session management, rate limiting

### Integration Technologies
- API Gateway: Azure API Management
- Message Queue: Azure Service Bus
- Monitoring: Azure Monitor + Grafana
- Logging: Azure Log Analytics

### Security
- OAuth2 for authentication
- JWT for service-to-service communication
- Azure Key Vault for secrets management

## Development Tools
- IDE: VS Code/Cursor with Go extensions
- API Testing: Playwright
- Load Testing: k6
- Code Quality: 
  - golangci-lint
  - gofmt
  - staticcheck
- Testing: 
  - testing package (standard library)
  - testify
  - gomock

## Development Standards
- Code Style: gofmt + golangci-lint
- API Documentation: OpenAPI/Swagger
- Commit Convention: Conventional Commits
- Branch Strategy: GitFlow
- Go specific:
  - Uber Go Style Guide
  - Project Layout Standard

## Open Source Setup
- GitHub Repository Structure
  - .github/
    - ISSUE_TEMPLATE/
    - PULL_REQUEST_TEMPLATE.md
    - workflows/
    - CONTRIBUTING.md
    - CODE_OF_CONDUCT.md
  - docs/
  - examples/
  - internal/
  - pkg/

## Environment Variable Management

### godotenv Implementation
- Package: github.com/joho/godotenv v1.5.1
- Purpose: Loads environment variables from .env files with proper precedence
- File Loading Order (highest to lowest priority):
  1. OS environment variables (always take precedence)
  2. `.env.{environment}.local` (e.g., .env.development.local)
  3. `.env.local` (skipped in test environment)
  4. `.env.{environment}` (e.g., .env.development)
  5. `.env` (base configuration)

### Environment Types
- development: Local development environment
- test: Testing environment (CI/CD, local tests)
- staging: Pre-production environment
- production: Live environment

### Best Practices
- Never commit .env files to version control
- Use .env.example as a template
- Store sensitive values in Azure Key Vault for production
- Use environment-specific files for different configurations
- Validate required environment variables on startup

### Environment Variable Conventions
- Prefix: CHATOPS_
- Format: UPPERCASE_WITH_UNDERSCORES
- Examples:
  - CHATOPS_GITHUB_TOKEN
  - CHATOPS_SLACK_BOT_TOKEN
  - CHATOPS_SLACK_SIGNING_KEY
  - CHATOPS_DB_HOST

### Implementation Details
- Custom wrapper in internal/infrastructure/env
- Automatic environment detection
- Graceful handling of missing files
- Comprehensive logging of loaded files
- Support for default values
- Required variable validation 