# NorthStack UI Portal Guide

**User Interfaces for All Personas**

---

## Portal Overview

| Portal | URL | Audience | Purpose |
|--------|-----|----------|---------|
| **Backstage** | portal.northstack.io | Developers | Self-service, templates |
| **Hasura** | graphql.northstack.io | Developers | GraphQL API console |
| **Grafana** | monitor.northstack.io | All users | Monitoring dashboards |
| **Superset** | analytics.northstack.io | Admins | Business analytics |

---

## 1. Backstage Developer Portal

### URL: `https://portal.northstack.io`

### Features

| Feature | Description |
|---------|-------------|
| **Software Catalog** | View all services, APIs, teams |
| **Templates** | Create services with one click |
| **TechDocs** | Documentation hub |
| **Kubernetes** | View k8s resources |
| **Search** | Find anything in your org |

### Templates Available

1. **NorthStack Web Service**
   - Creates GitHub repo
   - Deploys to NorthStack
   - Registers in catalog

2. **NorthStack Worker**
   - Background job processor
   - Auto-scaled

3. **NorthStack Database**
   - Provisions YugabyteDB
   - HA optional

### Quick Start

1. Login with GitHub SSO
2. Click "Create" in sidebar
3. Select "NorthStack Web Service"
4. Fill in details → Deploy!

---

## 2. Hasura GraphQL Console

### URL: `https://graphql.northstack.io/console`

### Features

| Feature | Description |
|---------|-------------|
| **GraphiQL** | Interactive query editor |
| **Schema Explorer** | Browse all types |
| **Subscriptions** | Real-time updates |
| **Actions** | Custom mutations |

### Sample Queries

```graphql
# Get my projects with services
query {
  projects {
    id
    name
    services {
      id
      name
      status
    }
  }
}

# Subscribe to service status
subscription {
  services(where: {project_id: {_eq: "..."}}) {
    id
    status
  }
}
```

### Actions

| Action | Description |
|--------|-------------|
| `triggerBuild` | Start a build |
| `scaleService` | Change replicas |
| `createDatabase` | Provision DB |

---

## 3. Grafana Monitoring

### URL: `https://monitor.northstack.io`

### Dashboards

| Dashboard | Audience | Metrics |
|-----------|----------|---------|
| **Platform Overview** | All | Projects, services, builds |
| **Services** | Developers | Per-service metrics |
| **Clusters** | Admins | K8s resource usage |
| **Databases** | Admins | YugabyteDB health |

### Key Panels

- **Request Rate** - Requests/second by status
- **Latency (p95)** - 95th percentile latency
- **Build Success Rate** - % successful builds
- **Resource Usage** - CPU/memory trends

### Alerts

Configured alerts for:
- High error rate (>5%)
- Service down
- High latency (>500ms)
- Database connection failures

---

## 4. Apache Superset Analytics

### URL: `https://analytics.northstack.io`

### Dashboards

| Dashboard | Description |
|-----------|-------------|
| **Executive Summary** | High-level KPIs |
| **Project Analytics** | Per-project metrics |
| **Cost Analysis** | Resource consumption |
| **Audit Report** | Security & compliance |

### Reports

- Daily deployment summary
- Weekly resource usage
- Monthly cost allocation
- Quarterly platform health

---

## Authentication

All portals use **Single Sign-On (SSO)**:

1. Login via GitHub OAuth
2. JWT token issued
3. Token includes role claims
4. Auto-redirect between portals

### Roles

| Role | Backstage | Hasura | Grafana | Superset |
|------|-----------|--------|---------|----------|
| Admin | Full | Full | Full | Full |
| Developer | Templates, Catalog | Query, Mutate | View | None |
| Viewer | Catalog | Query only | View | None |

---

## Deployment

```bash
# Deploy all UI components
kubectl apply -f config/backstage/deployment.yaml
kubectl apply -f config/grafana/deployment.yaml
kubectl apply -f config/superset/deployment.yaml
kubectl apply -f config/hasura/deployment.yaml
kubectl apply -f config/hasura/metadata.yaml
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     USERS                                   │
│  Developers │ DevOps │ Team Leads │ Admins │ Executives    │
└──────┬──────────┬──────────┬──────────┬──────────┬──────────┘
       │          │          │          │          │
       ▼          ▼          ▼          ▼          ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│Backstage │ │  Hasura  │ │ Grafana  │ │ Superset │
│ Portal   │ │ Console  │ │ Monitor  │ │Analytics │
└────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
     │            │            │            │
     └────────────┴────────────┴────────────┘
                       │
                       ▼
              ┌───────────────────┐
              │  NorthStack API   │
              │  (Orchestrator)   │
              └─────────┬─────────┘
                        │
              ┌─────────┴─────────┐
              │    YugabyteDB     │
              └───────────────────┘
```
