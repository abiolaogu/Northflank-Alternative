# NorthStack Platform - Product Requirements Document (PRD)

**Version:** 1.0  
**Date:** January 16, 2026  
**Status:** Production Ready

---

## 1. Executive Summary

NorthStack is a next-generation Internal Platform-as-a-Service (iPaaS) designed to be **10x better than Northflank**. It provides enterprises with a unified platform to deploy, scale, and manage applications across Kubernetes clusters with:

- **YugabyteDB** for distributed SQL (vs single-node PostgreSQL)
- **Hasura GraphQL** for instant APIs
- **ArgoCD + Coolify** for GitOps deployments
- **vCluster** tenant isolation

---

## 2. Product Vision

> *"Empower development teams to deploy and scale applications without infrastructure expertise, while giving platform teams enterprise-grade control."*

### Target Users
| Persona | Description |
|---------|-------------|
| **Developer** | Deploys apps, views logs, manages services |
| **DevOps Engineer** | Configures pipelines, manages clusters |
| **Platform Admin** | Manages users, quotas, policies |
| **Team Lead** | Views metrics, approves deployments |

---

## 3. Features & Capabilities

### 3.1 Project Management
| Feature | Description | Status |
|---------|-------------|--------|
| Create Projects | Organize services into projects | ✅ |
| Project Labels | Metadata tagging for organization | ✅ |
| Team Assignment | Assign projects to teams | ✅ |
| Project Deletion | Cascade delete with cleanup | ✅ |

### 3.2 Service Deployment
| Feature | Description | Status |
|---------|-------------|--------|
| Web Applications | Deploy frontend/backend apps | ✅ |
| Workers | Background job processors | ✅ |
| Cron Jobs | Scheduled tasks | ✅ |
| Stateful Databases | YugabyteDB, Redis | ✅ |
| Container Images | Deploy from any registry | ✅ |
| Git-based Builds | Build from GitHub/GitLab | ✅ |

### 3.3 CI/CD Pipeline
| Feature | Description | Status |
|---------|-------------|--------|
| GitHub Actions | Automated build/deploy | ✅ |
| Preview Environments | PR-based ephemeral envs | ✅ |
| ArgoCD Integration | GitOps deployments | ✅ |
| Rollback Support | One-click rollback | ✅ |
| Trivy Scanning | Security vulnerability scan | ✅ |

### 3.4 Database Management
| Feature | Description | Status |
|---------|-------------|--------|
| YugabyteDB Provisioning | Create distributed DB clusters | ✅ |
| Connection Pooling | PgBouncer integration | ✅ |
| Automated Backups | Scheduled backup to S3 | ✅ |
| Point-in-Time Recovery | Restore to any point | ✅ |
| Horizontal Scaling | Add/remove nodes | ✅ |

### 3.5 Cluster Management
| Feature | Description | Status |
|---------|-------------|--------|
| Multi-Cluster | Manage multiple K8s clusters | ✅ |
| Rancher Integration | Cluster provisioning | ✅ |
| Node Pool Management | Scale node groups | ✅ |
| Kubeconfig Download | Secure cluster access | ✅ |

### 3.6 API Access
| Feature | Description | Status |
|---------|-------------|--------|
| REST API | Full CRUD operations | ✅ |
| GraphQL (Hasura) | Real-time subscriptions | ✅ |
| Webhooks | Event notifications | ✅ |
| JWT Authentication | Secure API access | ✅ |

### 3.7 Observability
| Feature | Description | Status |
|---------|-------------|--------|
| Prometheus Metrics | Custom metrics | ✅ |
| Log Aggregation | Centralized logging | 🔄 |
| Alerting | Slack/Email notifications | ✅ |
| Tracing | Distributed tracing | 🔄 |

### 3.8 Dashboard UI
| Feature | Description | Status |
|---------|-------------|--------|
| Web Dashboard | Browser-based UI | 🔄 Planned |
| Mobile App | iOS/Android | 📋 Roadmap |

---

## 4. Technical Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENTS                                 │
│   Web Dashboard │ CLI │ Mobile App │ CI/CD Webhooks            │
└─────────────────┬───────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────────┐
│                      API GATEWAY                                │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│   │  REST API   │  │   Hasura    │  │  Webhooks   │            │
│   │  (Gin)      │  │  (GraphQL)  │  │  Handler    │            │
│   └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────┬───────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────────┐
│                    PLATFORM CORE                                │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐           │
│   │Projects │  │Services │  │ Builds  │  │Clusters │           │
│   └─────────┘  └─────────┘  └─────────┘  └─────────┘           │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐           │
│   │Database │  │ Secrets │  │ Events  │  │  Users  │           │
│   └─────────┘  └─────────┘  └─────────┘  └─────────┘           │
└─────────────────┬───────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────────┐
│                       ADAPTERS                                  │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐           │
│   │ Coolify │  │ ArgoCD  │  │ Rancher │  │YugabyteDB│          │
│   └─────────┘  └─────────┘  └─────────┘  └─────────┘           │
└─────────────────┬───────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────────┐
│                    INFRASTRUCTURE                               │
│   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐│
│   │   Kubernetes    │  │   YugabyteDB    │  │      NATS       ││
│   │   (via Rancher) │  │   (Distributed) │  │   (Messaging)   ││
│   └─────────────────┘  └─────────────────┘  └─────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Availability | 99.9% uptime |
| Latency | < 200ms API response |
| Scalability | 1000+ services |
| Security | SOC2 compliant |
| Recovery | < 15min RTO |

---

## 6. Release Timeline

| Phase | Features | Target |
|-------|----------|--------|
| **v1.0** | Core API, YugabyteDB, CI/CD | ✅ Complete |
| **v1.1** | Dashboard UI | Q1 2026 |
| **v1.2** | Multi-region | Q2 2026 |
| **v2.0** | Marketplace | Q3 2026 |

---

## 7. Success Metrics

- **Deployment time** < 5 minutes
- **Build success rate** > 95%
- **User satisfaction** NPS > 50
- **Platform adoption** 100+ projects
