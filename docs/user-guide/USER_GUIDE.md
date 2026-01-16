# NorthStack Platform - User Guide

**For Developers and Team Leads**

---

## 1. Getting Started

### 1.1 Logging In

1. Navigate to `https://api.northstack.io`
2. Login with your organization email
3. You'll receive a JWT token for API access

### 1.2 First-Time Setup

```bash
# Install CLI (coming soon)
curl -sSL https://get.northstack.io | sh

# Login
northstack login

# Create your first project
northstack project create my-app
```

---

## 2. Projects

A **Project** is a container for related services (e.g., frontend, backend, database).

### Creating a Project

**API:**
```bash
curl -X POST https://api.northstack.io/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Application",
    "description": "Production app"
  }'
```

**GraphQL:**
```graphql
mutation {
  insert_projects_one(object: {
    name: "My Application"
    description: "Production app"
  }) {
    id
    slug
  }
}
```

### Listing Projects

```bash
curl https://api.northstack.io/api/v1/projects \
  -H "Authorization: Bearer $TOKEN"
```

---

## 3. Services

A **Service** is a deployable unit (web app, worker, database, etc.).

### Service Types

| Type | Use Case |
|------|----------|
| `webapp` | Frontend/backend apps with HTTP |
| `worker` | Background processors |
| `cronjob` | Scheduled tasks |
| `stateful_db` | Databases (YugabyteDB) |

### Deploying a Service

**From Git:**
```bash
curl -X POST https://api.northstack.io/api/v1/projects/{project_id}/services \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "api-server",
    "type": "webapp",
    "build_source": {
      "type": "git",
      "repository": "https://github.com/org/repo",
      "branch": "main"
    },
    "resources": {
      "cpu_request": "250m",
      "memory_request": "256Mi"
    },
    "scaling": {
      "min_replicas": 2,
      "max_replicas": 10
    }
  }'
```

**From Docker Image:**
```bash
curl -X POST https://api.northstack.io/api/v1/projects/{project_id}/services \
  -d '{
    "name": "api-server",
    "build_source": {
      "type": "image",
      "image": "ghcr.io/org/app:latest"
    }
  }'
```

### Scaling a Service

```bash
curl -X POST https://api.northstack.io/api/v1/services/{id}/scale \
  -d '{"replicas": 5}'
```

---

## 4. Databases

NorthStack uses **YugabyteDB** for distributed SQL databases.

### Creating a Database

```bash
curl -X POST https://api.northstack.io/api/v1/projects/{project_id}/databases \
  -d '{
    "name": "production-db",
    "size": "medium",
    "storage_gb": 100,
    "high_availability": true,
    "backup_enabled": true
  }'
```

### Database Sizes

| Size | CPU | Memory | Use Case |
|------|-----|--------|----------|
| small | 500m | 1Gi | Development |
| medium | 1 | 2Gi | Staging |
| large | 2 | 4Gi | Production |
| xlarge | 4 | 8Gi | High traffic |

### Connecting to Your Database

```bash
# Get connection info
curl https://api.northstack.io/api/v1/databases/{id}/connection

# Response:
{
  "ysql_endpoint": "mydb-yb-tserver.northstack.svc:5433",
  "ycql_endpoint": "mydb-yb-tserver.northstack.svc:9042",
  "secret_name": "mydb-credentials"
}
```

---

## 5. Environment Variables

Set environment variables for your services:

```bash
curl -X PATCH https://api.northstack.io/api/v1/services/{id} \
  -d '{
    "env_vars": {
      "DATABASE_URL": "postgresql://...",
      "API_KEY": "secret123"
    }
  }'
```

---

## 6. Builds

### Triggering a Build

```bash
curl -X POST https://api.northstack.io/api/v1/services/{id}/builds \
  -d '{
    "branch": "main",
    "commit_sha": "abc123"
  }'
```

### Viewing Build Logs

```bash
curl https://api.northstack.io/api/v1/builds/{build_id}/logs
```

---

## 7. Monitoring

### Health Endpoints

| Endpoint | Purpose |
|----------|---------|
| `/health` | Basic health |
| `/health/live` | Liveness probe |
| `/health/ready` | Readiness probe |
| `/metrics` | Prometheus metrics |

### Key Metrics

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `service_replicas` - Running replicas
- `build_duration_seconds` - Build time

---

## 8. Troubleshooting

### Service Not Starting

1. Check build logs
2. Verify resource limits
3. Check health checks
4. Review environment variables

### Database Connection Issues

1. Verify secret is mounted
2. Check service account permissions
3. Confirm TLS settings

### Build Failures

1. Check Dockerfile syntax
2. Verify dependencies
3. Review Trivy scan results
