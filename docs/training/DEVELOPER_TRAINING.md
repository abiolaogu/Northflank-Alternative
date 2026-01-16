# NorthStack Platform - Developer Training Manual

**For Software Developers**

---

## Course Overview

| Module | Duration | Topics |
|--------|----------|--------|
| 1 | 20 min | Introduction & Setup |
| 2 | 30 min | Deploying Your First App |
| 3 | 30 min | Database Services |
| 4 | 20 min | Environment Management |
| 5 | 20 min | CI/CD Workflows |
| 6 | 15 min | Monitoring & Debugging |

**Total Duration:** ~2.5 hours

---

## Module 1: Introduction & Setup

### What is NorthStack?

NorthStack is your internal platform for deploying applications. Think:
- **Heroku simplicity** + **Kubernetes power**
- Deploy with Git push
- Auto-scaling and HA

### Getting Access

1. Request access from your platform admin
2. Receive welcome email with login link
3. Generate API token in settings

### API Token Setup

```bash
# Store token securely
export NORTHSTACK_TOKEN="your-token-here"

# Test connection
curl -H "Authorization: Bearer $NORTHSTACK_TOKEN" \
  https://api.northstack.io/health
```

---

## Module 2: Deploying Your First App

### Step 1: Create a Project

```bash
curl -X POST https://api.northstack.io/api/v1/projects \
  -H "Authorization: Bearer $NORTHSTACK_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "hello-world"}'
```

### Step 2: Create a Service

```bash
curl -X POST https://api.northstack.io/api/v1/projects/{PROJECT_ID}/services \
  -H "Authorization: Bearer $NORTHSTACK_TOKEN" \
  -d '{
    "name": "api",
    "type": "webapp",
    "build_source": {
      "type": "git",
      "repository": "https://github.com/yourorg/myapp",
      "branch": "main",
      "dockerfile": "Dockerfile"
    },
    "ports": [{"port": 8080, "name": "http", "public": true}],
    "resources": {
      "cpu_request": "100m",
      "memory_request": "128Mi"
    }
  }'
```

### Step 3: Check Deployment Status

```bash
curl https://api.northstack.io/api/v1/services/{SERVICE_ID}
```

### Step 4: Access Your App

Your app is available at:
`https://{service-slug}.{project-slug}.northstack.io`

---

## Module 3: Database Services

### Creating a Development Database

```bash
curl -X POST /api/v1/projects/{PROJECT_ID}/databases \
  -d '{
    "name": "dev-db",
    "size": "small",
    "storage_gb": 10
  }'
```

### Connecting from Your App

1. Get connection secret name from API response
2. Mount secret in your service:

```yaml
env:
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: dev-db-credentials
        key: ysql_uri
```

### PostgreSQL Compatibility

YugabyteDB is PostgreSQL-compatible. Use any PostgreSQL driver:

```python
# Python
import psycopg2
conn = psycopg2.connect(os.environ['DATABASE_URL'])
```

```go
// Go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

```javascript
// Node.js
const { Pool } = require('pg');
const pool = new Pool({ connectionString: process.env.DATABASE_URL });
```

---

## Module 4: Environment Management

### Setting Environment Variables

```bash
curl -X PATCH /api/v1/services/{SERVICE_ID} \
  -d '{
    "env_vars": {
      "NODE_ENV": "production",
      "LOG_LEVEL": "info",
      "API_KEY": "secret"
    }
  }'
```

### Using Secrets

For sensitive values, create a secret first:

```bash
# Create secret in K8s (via admin)
kubectl create secret generic my-api-keys \
  --from-literal=STRIPE_KEY=sk_live_xxx

# Reference in service
curl -X PATCH /api/v1/services/{SERVICE_ID} \
  -d '{
    "secret_refs": [
      {"secret_name": "my-api-keys", "key": "STRIPE_KEY"}
    ]
  }'
```

---

## Module 5: CI/CD Workflows

### GitHub Actions Integration

Add this workflow to your repo:

```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy to NorthStack
        run: |
          curl -X POST $NORTHSTACK_API_URL/services/$SERVICE_ID/builds \
            -H "Authorization: Bearer ${{ secrets.NORTHSTACK_TOKEN }}" \
            -d '{"branch": "${{ github.ref_name }}"}'
```

### Preview Environments

PRs automatically get preview environments:
- URL: `pr-{number}.preview.northstack.io`
- Auto-deleted when PR closes

---

## Module 6: Monitoring & Debugging

### Viewing Logs

```bash
# Get recent logs
curl /api/v1/services/{SERVICE_ID}/logs?tail=100

# Stream logs (coming soon)
northstack logs -f {SERVICE_ID}
```

### Checking Health

```bash
# Service health
curl /api/v1/services/{SERVICE_ID}/health

# Response
{
  "status": "healthy",
  "replicas": {
    "desired": 3,
    "ready": 3
  }
}
```

### Common Issues

| Symptom | Possible Cause | Solution |
|---------|----------------|----------|
| CrashLoopBackOff | App crashes on start | Check logs, fix code |
| OOMKilled | Out of memory | Increase memory limit |
| ImagePullBackOff | Can't pull image | Check registry credentials |
| Pending | No resources | Contact admin |

---

## Hands-On Lab

Complete these tasks:

1. ✅ Create a project called "lab-project"
2. ✅ Deploy a web service from GitHub
3. ✅ Create a database
4. ✅ Connect your app to the database
5. ✅ Trigger a new build

---

## Quick Reference

### API Endpoints

| Action | Method | Endpoint |
|--------|--------|----------|
| List projects | GET | /api/v1/projects |
| Create project | POST | /api/v1/projects |
| Create service | POST | /api/v1/projects/{id}/services |
| Trigger build | POST | /api/v1/services/{id}/builds |
| Scale service | POST | /api/v1/services/{id}/scale |
| Create database | POST | /api/v1/projects/{id}/databases |

### GraphQL Endpoint

```
https://graphql.northstack.io/v1/graphql
```

---

## Certification

Complete all lab tasks to receive Developer Certification.
