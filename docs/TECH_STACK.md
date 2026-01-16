# SkyForge Technology Stack

**Version:** 1.0.0  
**Last Updated:** January 2026

---

## Overview

SkyForge is built using modern, cloud-native technologies with a focus on:
- **Domain-Driven Design (DDD)** for maintainable code
- **Extreme Programming (XP)** practices
- **Multi-cloud portability**
- **Production-ready security**

---

## Core Technologies

### Backend

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.21+ | High-performance backend |
| **Framework** | Gin | HTTP routing & middleware |
| **Architecture** | DDD | Clean, maintainable code |
| **API** | REST + GraphQL | Flexible API access |

### Data Layer

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Primary Database** | YugabyteDB | Distributed SQL, HA |
| **Cache** | DragonflyDB | Redis-compatible, high perf |
| **Messaging** | NATS JetStream | Async events |
| **Streaming** | Redpanda | Kafka-compatible |

### Deployment & CI/CD

| Component | Technology | Purpose |
|-----------|------------|---------|
| **GitOps** | ArgoCD | Declarative deployments |
| **Build** | Coolify | Container builds |
| **CI** | GitHub Actions | Automated testing |
| **Container Registry** | GHCR / ECR / GCR | Image storage |

### Observability

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Metrics** | Prometheus | Time-series metrics |
| **Visualization** | Grafana | Dashboards |
| **Logs** | Loki | Log aggregation |
| **Traces** | Jaeger | Distributed tracing |

### Frontend

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Developer Portal** | Backstage | Unified UI (46 screens) |
| **Mobile App** | Flutter | iOS + Android |
| **Analytics** | Apache Superset | Business dashboards |

### Infrastructure

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Kubernetes** | RKE2 / EKS / GKE / AKS | Container orchestration |
| **Ingress** | Nginx / ALB / Gateway API | Traffic routing |
| **TLS** | cert-manager | Certificate automation |
| **Secrets** | External Secrets Operator | Multi-cloud secrets |

---

## Multi-Cloud Support

| Platform | Kubernetes | Storage | LoadBalancer | Secrets |
|----------|------------|---------|--------------|---------|
| **Harvester** | RKE2 | Longhorn | Nginx/MetalLB | K8s Secrets |
| **AWS** | EKS | EBS CSI | ALB | Secrets Manager |
| **GCP** | GKE | PD CSI | Gateway API | Secret Manager |
| **Azure** | AKS | Disk CSI | AGIC | Key Vault |
| **OpenStack** | Magnum | Cinder | Octavia | Barbican |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENTS                                  │
│         Web Portal (Backstage)  │  Mobile App (Flutter)        │
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────────┐
│                    LOAD BALANCER                                │
│            (Nginx / ALB / Gateway API / AGIC)                   │
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────────┐
│                    API GATEWAY                                  │
│        Rate Limiting │ Auth │ Tracing │ Metrics                │
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────────┐
│                  SKYFORGE API (Go/Gin)                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Projects    │  │ Services    │  │ Databases   │             │
│  ├─────────────┤  ├─────────────┤  ├─────────────┤             │
│  │ Domains     │  │ Builds      │  │ Secrets     │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└────────┬───────────────┬───────────────┬────────────────────────┘
         │               │               │
┌────────▼───────┐ ┌─────▼─────┐ ┌───────▼───────┐
│  YugabyteDB    │ │   NATS    │ │   Redpanda    │
│  (SQL)         │ │ JetStream │ │   (Kafka)     │
└────────────────┘ └───────────┘ └───────────────┘
         │
┌────────▼──────────────────────────────────────────────────────┐
│                    INTEGRATIONS                                │
│  ArgoCD  │  Coolify  │  Rancher  │  Hasura  │  Grafana        │
└───────────────────────────────────────────────────────────────┘
```

---

## Package Dependencies

### Go Modules

```
github.com/gin-gonic/gin          # HTTP framework
github.com/jackc/pgx/v5           # PostgreSQL driver
github.com/nats-io/nats.go        # NATS client
github.com/twmb/franz-go          # Kafka/Redpanda client
github.com/redis/go-redis/v9      # Redis client
github.com/golang-jwt/jwt/v5      # JWT auth
go.uber.org/zap                   # Structured logging
golang.org/x/time/rate            # Rate limiting
```

### Frontend

```
@backstage/core-components        # UI components
@backstage/core-plugin-api        # Plugin framework
@material-ui/core                 # Material Design
react-router-dom                  # Routing
```

---

## Security Features

| Feature | Implementation |
|---------|----------------|
| **Authentication** | JWT + OAuth 2.0 |
| **Authorization** | RBAC per project |
| **Secrets** | External Secrets Operator |
| **Network** | Zero-trust NetworkPolicies |
| **TLS** | cert-manager, Let's Encrypt |
| **Scanning** | gosec, govulncheck, trivy |

---

## Performance Targets

| Metric | Target |
|--------|--------|
| API Latency (p95) | < 100ms |
| Availability | 99.9% |
| Max Concurrent Users | 10,000 |
| Build Time | < 5 min |
| Deploy Time | < 60 sec |

---

## Development Tools

| Tool | Purpose |
|------|---------|
| `go test` | Unit & E2E testing |
| `gosec` | Security scanning |
| `govulncheck` | Vulnerability check |
| `trivy` | Container scanning |
| `helm lint` | Chart validation |
