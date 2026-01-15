# OpenPaaS - Open Source Northflank Alternative

A comprehensive, self-hosted Platform-as-a-Service (PaaS) that unifies best-in-class open-source tools into a cohesive developer platform.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         UNIFIED PLATFORM LAYER                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Platform Orchestrator (Go)                        │   │
│  │   • Unified API Gateway    • Event Aggregation    • State Machine   │   │
│  │   • Webhook Router         • RBAC Proxy           • Audit Logger    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────────┤
│                              TOOL LAYER                                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Coolify   │    │   Rancher   │    │   ArgoCD    │    │   Vault     │  │
│  │  (CI/Build) │    │  (K8s Mgmt) │    │  (GitOps)   │    │  (Secrets)  │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
├─────────────────────────────────────────────────────────────────────────────┤
│                           INFRASTRUCTURE LAYER                               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │ Kubernetes  │    │ PostgreSQL  │    │    NATS     │    │ Prometheus  │  │
│  │  (K3s/RKE2) │    │  (Metadata) │    │ (Event Bus) │    │  + Grafana  │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Features

- **Unified API**: Single REST API to manage all platform resources
- **GitOps Native**: Automatic deployments from Git repositories via ArgoCD
- **Multi-Cluster**: Manage multiple Kubernetes clusters across providers via Rancher
- **CI/CD Pipeline**: Build and deploy containerized applications via Coolify
- **Event-Driven**: Real-time events via NATS JetStream
- **Audit Logging**: Complete audit trail of all platform actions
- **RBAC**: Role-based access control for teams and projects
- **Metrics**: Built-in Prometheus metrics and Grafana dashboards

## Component Stack

| Component | Tool | Purpose |
|-----------|------|---------|
| CI/Build | Coolify | Docker builds, buildpacks, preview environments |
| Kubernetes Management | Rancher | Multi-cluster management, provisioning |
| GitOps/CD | ArgoCD | Declarative deployments, auto-sync |
| Event Bus | NATS | Event streaming, persistence (JetStream) |
| Secrets | HashiCorp Vault | Secret management, dynamic credentials |
| Monitoring | Prometheus + Grafana | Metrics, alerting, dashboards |
| Logging | Loki + Promtail | Log aggregation |
| Ingress | Traefik | Automatic TLS, routing |
| Database | PostgreSQL | Platform metadata |
| Cache | Redis | Session management, rate limiting |

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for development)
- Kubernetes cluster (for production)

### Local Development

```bash
# Clone the repository
git clone https://github.com/openpaas/platform-orchestrator.git
cd platform-orchestrator

# Start infrastructure services
cd docker
docker-compose up -d postgres nats redis

# Run the orchestrator
cd ..
go run ./cmd/orchestrator --migrate
go run ./cmd/orchestrator
```

### Docker Compose (Full Stack)

```bash
cd docker
docker-compose up -d
```

Access the API at `http://localhost:8080`

### Kubernetes Deployment

Using Helm:

```bash
# Add the OpenPaaS Helm repository
helm repo add openpaas https://charts.openpaas.io
helm repo update

# Install with custom values
helm install openpaas openpaas/openpaas \
  --namespace openpaas \
  --create-namespace \
  -f values.yaml
```

Or using kubectl:

```bash
kubectl apply -f deployments/kubernetes/
```

## Configuration

Configuration is loaded from environment variables with the `OPENPAAS_` prefix:

```bash
# Server
OPENPAAS_SERVER_HOST=0.0.0.0
OPENPAAS_SERVER_PORT=8080

# Database
OPENPAAS_DATABASE_HOST=localhost
OPENPAAS_DATABASE_PORT=5432
OPENPAAS_DATABASE_NAME=openpaas
OPENPAAS_DATABASE_USER=openpaas
OPENPAAS_DATABASE_PASSWORD=secret

# NATS
OPENPAAS_NATS_URL=nats://localhost:4222

# Authentication
OPENPAAS_AUTH_JWT_SECRET=your-secret-key

# External Services
OPENPAAS_COOLIFY_URL=http://localhost:3000
OPENPAAS_COOLIFY_API_KEY=your-api-key
OPENPAAS_RANCHER_URL=https://rancher.local
OPENPAAS_RANCHER_TOKEN=your-token
OPENPAAS_ARGOCD_URL=https://argocd.local
OPENPAAS_ARGOCD_TOKEN=your-token
```

See `config/config.example.yaml` for a complete configuration reference.

## API Overview

### Projects

```bash
# Create a project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "My App", "slug": "my-app"}'

# List projects
curl http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN"
```

### Services

```bash
# Create a service
curl -X POST http://localhost:8080/api/v1/projects/{project_id}/services \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Server",
    "slug": "api-server",
    "type": "webapp",
    "build_source": {
      "type": "git",
      "repository": "https://github.com/user/repo",
      "branch": "main"
    },
    "ports": [{"name": "http", "port": 8080, "public": true}]
  }'

# Trigger a build
curl -X POST http://localhost:8080/api/v1/services/{service_id}/builds \
  -H "Authorization: Bearer $TOKEN"

# Scale a service
curl -X POST http://localhost:8080/api/v1/services/{service_id}/scale \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"replicas": 3}'
```

### Health Checks

```bash
# Liveness probe
curl http://localhost:8080/health/live

# Readiness probe
curl http://localhost:8080/health/ready
```

## Project Structure

```
.
├── cmd/
│   └── orchestrator/       # Main application entry point
├── internal/
│   ├── adapters/           # External service adapters
│   │   ├── argocd/         # ArgoCD GitOps adapter
│   │   ├── coolify/        # Coolify CI/Build adapter
│   │   ├── rancher/        # Rancher cluster management adapter
│   │   └── vault/          # HashiCorp Vault adapter
│   ├── api/                # HTTP API
│   │   ├── handlers/       # Request handlers
│   │   └── middleware/     # HTTP middleware
│   ├── audit/              # Audit logging
│   ├── config/             # Configuration management
│   ├── domain/             # Core domain models and interfaces
│   ├── eventbus/           # NATS event bus
│   ├── repository/         # Database repositories
│   └── workflow/           # Deployment state machine
├── pkg/
│   ├── errors/             # Error handling utilities
│   └── logger/             # Structured logging
├── deployments/
│   ├── kubernetes/         # Kubernetes manifests
│   └── helm/               # Helm charts
└── docker/                 # Docker configuration
```

## Development

### Building

```bash
# Build binary
go build -o bin/orchestrator ./cmd/orchestrator

# Build Docker image
docker build -t openpaas/platform-orchestrator:dev -f docker/Dockerfile .
```

### Testing

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Database Migrations

```bash
# Run migrations
./bin/orchestrator --migrate
```

## Roadmap

- [ ] Web UI Dashboard
- [ ] Database-as-a-Service (PostgreSQL, MySQL, MongoDB)
- [ ] Redis/Message Queue provisioning
- [ ] Custom domain management
- [ ] Preview environments for PRs
- [ ] Cost tracking and resource quotas
- [ ] Terraform integration for infrastructure
- [ ] Plugin system for custom integrations

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting a pull request.

## License

Apache License 2.0 - see LICENSE file for details.
