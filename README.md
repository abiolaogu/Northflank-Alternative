<p align="center">
  <img src="docs/assets/logo.png" alt="SkyForge Logo" width="200"/>
</p>

<h1 align="center">SkyForge</h1>
<p align="center">
  <strong>The Open-Source Internal Platform as a Service</strong>
</p>

<p align="center">
  <a href="https://github.com/abiolaogu/Northflank-Alternative/actions"><img src="https://github.com/abiolaogu/Northflank-Alternative/workflows/CI/badge.svg" alt="CI Status"></a>
  <a href="https://goreportcard.com/report/github.com/northstack/platform"><img src="https://goreportcard.com/badge/github.com/northstack/platform" alt="Go Report Card"></a>
  <a href="https://github.com/abiolaogu/Northflank-Alternative/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://github.com/abiolaogu/Northflank-Alternative/releases"><img src="https://img.shields.io/github/v/release/abiolaogu/Northflank-Alternative" alt="Release"></a>
</p>

---

## 🚀 What is SkyForge?

**SkyForge** is a comprehensive, self-hosted Internal Platform as a Service (iPaaS) that gives engineering teams the power of Northflank, Heroku, and Vercel—on your own infrastructure.

Deploy to **Harvester**, **AWS**, **GCP**, **Azure**, or **OpenStack** with a single Helm command.

---

## ✨ Key Features

| Category | Features |
|----------|----------|
| **🏗️ Project Management** | Multi-tenant projects, team RBAC, secrets management |
| **🚢 Deployment** | GitOps (ArgoCD), Build pipelines (Coolify), Auto-scaling |
| **💾 Databases** | YugabyteDB (Distributed SQL), auto-provisioning, backups |
| **📊 Observability** | Grafana dashboards, Prometheus metrics, Loki logs |
| **🔄 Event Streaming** | Redpanda (Kafka-compatible), NATS JetStream |
| **🖥️ Developer Portal** | Backstage with 46+ screens, API explorer, mobile app |
| **🌍 Multi-Cloud** | Deploy anywhere: Harvester, AWS, GCP, Azure, OpenStack |

---

## 🎯 Why SkyForge?

| vs. Northflank | vs. Heroku | vs. DIY |
|----------------|------------|---------|
| ✅ Self-hosted | ✅ No vendor lock-in | ✅ Pre-built, production-ready |
| ✅ No per-seat pricing | ✅ Full Kubernetes power | ✅ Best practices built-in |
| ✅ GitOps native | ✅ Custom domains | ✅ Unified portal |
| ✅ Multi-cloud | ✅ Team collaboration | ✅ Observability included |

---

## 🏃 Quick Start

### One-Command Deploy

```bash
# Add Helm repo
helm repo add skyforge https://charts.skyforge.io

# Deploy to your cluster
helm install skyforge skyforge/skyforge \
  --namespace skyforge \
  --create-namespace \
  --set global.cloudProvider=harvester
```

### Cloud-Specific Deployments

```bash
# AWS EKS
helm install skyforge skyforge/skyforge --set global.cloudProvider=aws --set aws.enabled=true

# GCP GKE  
helm install skyforge skyforge/skyforge --set global.cloudProvider=gcp --set gcp.enabled=true

# Azure AKS
helm install skyforge skyforge/skyforge --set global.cloudProvider=azure --set azure.enabled=true
```

---

## 📁 Project Structure

```
├── cmd/orchestrator/          # API server entrypoint
├── internal/
│   ├── api/                   # HTTP handlers & middleware
│   ├── domain/                # DDD domain models
│   ├── adapters/              # External service integrations
│   └── repository/            # Data access layer
├── pkg/                       # Shared packages
├── config/                    # Kubernetes & Backstage configs
├── deploy/                    # Multi-cloud deployment manifests
│   ├── harvester/
│   ├── aws/
│   ├── gcp/
│   ├── azure/
│   └── helm/
├── mobile/                    # Flutter mobile app
└── docs/                      # Documentation
```

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go, Gin, DDD architecture |
| **Database** | YugabyteDB (Distributed SQL) |
| **Cache** | DragonflyDB (Redis-compatible) |
| **Messaging** | NATS JetStream, Redpanda |
| **GitOps** | ArgoCD |
| **CI/CD** | Coolify, GitHub Actions |
| **Portal** | Backstage |
| **Monitoring** | Prometheus, Grafana, Loki |
| **Mobile** | Flutter |

---

## 📖 Documentation

- [Architecture Guide](docs/architecture/ARCHITECTURE.md)
- [Tech Stack](docs/TECH_STACK.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [UI Screens](docs/UI_SCREENS.md)
- [API Reference](docs/api/)

---

## 🖥️ UI Screens (46 Total)

| Category | Screens |
|----------|---------|
| Project Management | 8 |
| Services | 12 |
| Databases | 6 |
| Build & Deploy | 8 |
| Observability | 6 |
| Settings | 6 |

See [UI_SCREENS.md](docs/UI_SCREENS.md) for full catalog.

---

## 🔒 Security

- JWT + OAuth 2.0 authentication
- RBAC authorization
- Network policies (zero trust)
- Secrets via External Secrets Operator
- Security scanner included (`scripts/security-scan.sh`)

---

## 🧪 Testing

```bash
# Unit tests
go test ./... -v -short

# E2E tests
go test ./tests/e2e/... -v

# Security scan
./scripts/security-scan.sh
```

---

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

---

## 📜 License

Apache 2.0 - See [LICENSE](LICENSE)

---

<p align="center">
  <strong>Built with ❤️ for Platform Engineers</strong>
</p>
