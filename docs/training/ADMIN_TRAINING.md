# NorthStack Platform - Administrator Training Manual

**For Platform Administrators and DevOps Engineers**

---

## Module 1: Platform Overview (30 minutes)

### Learning Objectives
- Understand NorthStack architecture
- Identify key components
- Know when to use each feature

### Key Concepts

| Component | Role |
|-----------|------|
| Orchestrator API | Central control plane |
| Hasura | GraphQL API layer |
| YugabyteDB | Distributed database |
| Coolify | CI/CD builds |
| ArgoCD | GitOps deployments |
| Rancher | Cluster management |

### Architecture Review
*Refer to docs/architecture/ARCHITECTURE.md*

---

## Module 2: User Management (45 minutes)

### User Roles

| Role | Permissions |
|------|-------------|
| Admin | Full access, cluster management |
| Owner | Team + project management |
| Member | Deploy services, view logs |
| Viewer | Read-only access |

### Creating Users

```bash
# Via API
curl -X POST https://api.northstack.io/api/v1/users \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "role": "member"
  }'
```

### Managing Teams

```bash
# Create team
curl -X POST /api/v1/teams \
  -d '{"name": "Backend Team"}'

# Add member
curl -X POST /api/v1/teams/{id}/members \
  -d '{"user_id": "...", "role": "member"}'
```

---

## Module 3: Cluster Management (60 minutes)

### Adding a Cluster

```bash
curl -X POST /api/v1/clusters \
  -d '{
    "name": "production-us-east",
    "provider": "rancher",
    "region": "us-east-1",
    "kube_version": "1.28",
    "node_count": 5
  }'
```

### Cluster Providers

| Provider | Use Case |
|----------|----------|
| rancher | On-premise, Harvester |
| rke2 | Production grade |
| k3s | Development, edge |
| eks | AWS managed |
| gke | GCP managed |
| aks | Azure managed |

### Scaling Nodes

```bash
# Scale cluster
curl -X PATCH /api/v1/clusters/{id} \
  -d '{"node_count": 10}'
```

### Getting Kubeconfig

```bash
curl /api/v1/clusters/{id}/kubeconfig > ~/.kube/production
```

---

## Module 4: Database Administration (60 minutes)

### YugabyteDB Overview

- **YSQL** - PostgreSQL-compatible (port 5433)
- **YCQL** - Cassandra-compatible (port 9042)
- **Distributed** - Data replicated across nodes
- **HA** - Automatic failover

### Creating Production Database

```bash
curl -X POST /api/v1/projects/{project}/databases \
  -d '{
    "name": "production",
    "size": "large",
    "storage_gb": 500,
    "high_availability": true,
    "backup_enabled": true,
    "tls_enabled": true
  }'
```

### Scaling Database

```bash
curl -X POST /api/v1/databases/{id}/scale \
  -d '{"replicas": 5}'
```

### Backup Management

Backups are automatic when `backup_enabled: true`:
- Daily: Full backup at 2 AM
- Weekly: Full backup Sunday
- Retention: 30 days

### Restore from Backup

```bash
curl -X POST /api/v1/databases/{id}/restore \
  -d '{"backup_id": "...", "point_in_time": "2026-01-15T10:00:00Z"}'
```

---

## Module 5: CI/CD Configuration (45 minutes)

### GitHub Actions Secrets

Add these secrets to your GitHub repository:

| Secret | Description |
|--------|-------------|
| `NORTHSTACK_API_TOKEN` | API authentication |
| `COOLIFY_API_TOKEN` | Build trigger |
| `ARGOCD_TOKEN` | Deployment sync |

### Workflow Configuration

See `.github/workflows/deploy.yml` for complete example.

### Preview Environments

Every PR automatically gets:
- Ephemeral namespace
- Database clone
- Unique URL: `pr-123.preview.northstack.io`

---

## Module 6: Monitoring & Alerting (45 minutes)

### Prometheus Metrics

Access at: `https://api.northstack.io/metrics`

Key metrics to monitor:
- `http_requests_total{status="5xx"}` - Errors
- `http_request_duration_seconds_bucket` - Latency
- `go_goroutines` - Resource usage

### Setting Up Alerts

Configure in Prometheus/Alertmanager:

```yaml
groups:
  - name: northstack
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status="5xx"}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
```

### Slack Notifications

Configure webhook in cluster settings:
```bash
curl -X PATCH /api/v1/settings \
  -d '{"slack_webhook": "https://hooks.slack.com/..."}'
```

---

## Module 7: Security Best Practices (30 minutes)

### Checklist

- [ ] Enable TLS for all databases
- [ ] Rotate API tokens quarterly
- [ ] Review audit logs weekly
- [ ] Enable Trivy scanning in CI
- [ ] Use least-privilege RBAC
- [ ] Configure network policies

### Viewing Audit Logs

```bash
curl /api/v1/audit-logs?user_id=...&action=delete
```

---

## Module 8: Disaster Recovery (30 minutes)

### Backup Strategy

| Component | Strategy | RTO | RPO |
|-----------|----------|-----|-----|
| API State | YugabyteDB | 15min | 1min |
| Secrets | External Secrets | 5min | 0 |
| Config | Git (ArgoCD) | 5min | 0 |
| Databases | S3 Backups | 30min | 1hr |

### Recovery Procedure

1. Verify cluster health
2. Restore YugabyteDB from backup
3. Trigger ArgoCD sync
4. Verify services

---

## Assessment

Complete the following to earn certification:

1. Create a multi-node cluster
2. Deploy a service with HA
3. Provision a YugabyteDB database
4. Configure monitoring alerts
5. Perform a database restore
