# NorthStack Platform - Complete UI Screens

**Replicates all Northflank dashboard features + enhancements**

---

## Dashboard Screens Overview

| Category | Screens | Northflank Equivalent |
|----------|---------|----------------------|
| **Project Management** | 8 | ✅ Full parity |
| **Services** | 12 | ✅ Full parity |
| **Databases** | 6 | ✅ Enhanced (YugabyteDB) |
| **Build & Deploy** | 8 | ✅ Full parity |
| **Observability** | 6 | ✅ Enhanced (Grafana) |
| **Settings** | 6 | ✅ Full parity |
| **Total** | **46** | |

---

## 1. Project Management

| Screen | Route | Description |
|--------|-------|-------------|
| Projects List | `/projects` | All projects with stats |
| Project Overview | `/projects/:id` | Dashboard, resources, activity |
| Project Settings | `/projects/:id/settings` | Name, team, danger zone |
| Project Members | `/projects/:id/members` | Team access control |
| Project Billing | `/projects/:id/billing` | Usage, costs, quotas |
| Project Secrets | `/projects/:id/secrets` | Encrypted variables |
| Project Audit Log | `/projects/:id/audit` | Activity history |
| Create Project | `/projects/new` | Project wizard |

---

## 2. Services (Workloads)

| Screen | Route | Description |
|--------|-------|-------------|
| Services List | `/projects/:id/services` | All deployments |
| Service Overview | `/services/:id` | Health, replicas, resources |
| Service Logs | `/services/:id/logs` | Real-time log stream |
| Service Metrics | `/services/:id/metrics` | CPU, memory, network |
| Service Scaling | `/services/:id/scaling` | Auto-scale config |
| Service Environment | `/services/:id/env` | Environment variables |
| Service Ports | `/services/:id/ports` | Network config, domains |
| Service Health Checks | `/services/:id/health` | Probe configuration |
| Service Volumes | `/services/:id/volumes` | Persistent storage |
| Service Shell | `/services/:id/shell` | Container terminal |
| Service Events | `/services/:id/events` | K8s events timeline |
| Create Service | `/services/new` | Deployment wizard |

---

## 3. Databases (Addons)

| Screen | Route | Description |
|--------|-------|-------------|
| Databases List | `/projects/:id/databases` | All database clusters |
| Database Overview | `/databases/:id` | Status, nodes, storage |
| Database Connection | `/databases/:id/connect` | Connection strings, credentials |
| Database Backups | `/databases/:id/backups` | Backup history, restore |
| Database Scaling | `/databases/:id/scaling` | Replicas, storage |
| Database Metrics | `/databases/:id/metrics` | Query stats, connections |

---

## 4. Build & Deploy (CI/CD)

| Screen | Route | Description |
|--------|-------|-------------|
| Builds List | `/builds` | All build jobs |
| Build Detail | `/builds/:id` | Logs, artifacts, duration |
| Build Settings | `/services/:id/build` | Dockerfile, branch, triggers |
| Deployments List | `/deployments` | Deploy history |
| Deployment Detail | `/deployments/:id` | Rollout status, replicas |
| Pipeline View | `/projects/:id/pipelines` | CI/CD workflow |
| Git Connections | `/settings/git` | GitHub, GitLab, Bitbucket |
| Webhooks | `/settings/webhooks` | Webhook endpoints |

---

## 5. Observability (Monitoring)

| Screen | Route | Description |
|--------|-------|-------------|
| Dashboard | `/monitoring` | Platform overview |
| Metrics Explorer | `/monitoring/metrics` | Prometheus queries |
| Logs Explorer | `/monitoring/logs` | Loki log search |
| Alerts | `/monitoring/alerts` | Alert rules, history |
| Traces | `/monitoring/traces` | Distributed tracing |
| Status Page | `/status` | System health |

---

## 6. Settings & Admin

| Screen | Route | Description |
|--------|-------|-------------|
| Account Settings | `/settings/account` | Profile, password, 2FA |
| Team Management | `/settings/team` | Members, invites, roles |
| API Tokens | `/settings/tokens` | Personal access tokens |
| Billing | `/settings/billing` | Plans, invoices, usage |
| Integrations | `/settings/integrations` | Slack, PagerDuty, etc. |
| Clusters | `/admin/clusters` | K8s cluster management |

---

## UI Components in Backstage

All screens are implemented as Backstage plugins with:

- **React TypeScript** components
- **Material-UI** styling
- **Responsive** design (mobile-ready)
- **Dark mode** support
- **Real-time updates** via WebSocket/GraphQL subscriptions

---

## Screen Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    BACKSTAGE PORTAL                         │
├─────────────┬───────────────────────────────────────────────┤
│             │                                               │
│   Sidebar   │              Main Content Area                │
│             │                                               │
│  ┌────────┐ │  ┌─────────────────────────────────────────┐ │
│  │ Home   │ │  │  Header: Title + Actions               │ │
│  ├────────┤ │  ├─────────────────────────────────────────┤ │
│  │Projects│ │  │  Tabs: Overview | Logs | Metrics | ... │ │
│  ├────────┤ │  ├─────────────────────────────────────────┤ │
│  │Services│ │  │                                         │ │
│  ├────────┤ │  │  Content: Tables, Charts, Forms        │ │
│  │Databases│ │ │                                         │ │
│  ├────────┤ │  │                                         │ │
│  │Builds  │ │  │                                         │ │
│  ├────────┤ │  └─────────────────────────────────────────┘ │
│  │Monitor │ │                                               │
│  ├────────┤ │                                               │
│  │Settings│ │                                               │
│  └────────┘ │                                               │
│             │                                               │
└─────────────┴───────────────────────────────────────────────┘
```

---

## Comparison with Northflank

| Feature | Northflank | NorthStack |
|---------|------------|------------|
| Projects | ✅ | ✅ |
| Services (Combined) | ✅ | ✅ |
| Jobs (Cron) | ✅ | ✅ |
| Databases (Addons) | ✅ | ✅ YugabyteDB |
| Build Pipelines | ✅ | ✅ Coolify |
| GitOps | ❌ | ✅ ArgoCD |
| Secrets | ✅ | ✅ |
| Domains | ✅ | ✅ |
| Logs | ✅ | ✅ Loki |
| Metrics | ✅ | ✅ Prometheus |
| Alerts | ✅ | ✅ Grafana |
| Team Management | ✅ | ✅ |
| RBAC | ✅ | ✅ |
| API | ✅ | ✅ + GraphQL |
| CLI | ✅ | 📋 Planned |
| Mobile App | ❌ | ✅ Flutter |
| Self-Hosted | ❌ | ✅ |
