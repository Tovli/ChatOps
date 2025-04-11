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