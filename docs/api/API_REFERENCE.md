# NorthStack Platform - API Reference

**Version:** v1  
**Base URL:** `https://api.northstack.io/api/v1`

---

## Authentication

All requests require a Bearer token:

```
Authorization: Bearer <your-token>
```

---

## Projects

### List Projects

```http
GET /projects
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| limit | int | Max results (default: 50) |
| offset | int | Pagination offset |
| search | string | Search by name |

**Response:** `200 OK`
```json
{
  "projects": [...],
  "total": 42
}
```

### Create Project

```http
POST /projects
```

**Request Body:**
```json
{
  "name": "my-project",
  "description": "Optional description",
  "labels": {"env": "production"}
}
```

**Response:** `201 Created`

### Get Project

```http
GET /projects/{id}
```

### Update Project

```http
PATCH /projects/{id}
```

### Delete Project

```http
DELETE /projects/{id}
```

---

## Services

### Create Service

```http
POST /projects/{project_id}/services
```

**Request Body:**
```json
{
  "name": "api-server",
  "type": "webapp",
  "build_source": {
    "type": "git",
    "repository": "https://github.com/org/repo",
    "branch": "main",
    "dockerfile": "Dockerfile"
  },
  "resources": {
    "cpu_request": "250m",
    "cpu_limit": "1",
    "memory_request": "256Mi",
    "memory_limit": "512Mi"
  },
  "scaling": {
    "min_replicas": 2,
    "max_replicas": 10,
    "target_cpu": 70
  },
  "ports": [
    {"name": "http", "port": 8080, "public": true}
  ],
  "env_vars": {
    "NODE_ENV": "production"
  }
}
```

### Service Types

| Type | Description |
|------|-------------|
| webapp | Web application with HTTP |
| worker | Background processor |
| cronjob | Scheduled task |
| stateful_db | Database service |

### List Services

```http
GET /projects/{project_id}/services
```

### Get Service

```http
GET /services/{id}
```

### Update Service

```http
PATCH /services/{id}
```

### Delete Service

```http
DELETE /services/{id}
```

### Scale Service

```http
POST /services/{id}/scale
```

**Request Body:**
```json
{"replicas": 5}
```

### Trigger Build

```http
POST /services/{id}/builds
```

**Request Body:**
```json
{
  "branch": "main",
  "commit_sha": "abc123"
}
```

---

## Databases

### Create Database

```http
POST /projects/{project_id}/databases
```

**Request Body:**
```json
{
  "name": "production-db",
  "size": "large",
  "storage_gb": 100,
  "high_availability": true,
  "backup_enabled": true,
  "tls_enabled": true,
  "version": "2.20.1.0-b97"
}
```

### Database Sizes

| Size | CPU | Memory |
|------|-----|--------|
| small | 500m | 1Gi |
| medium | 1 | 2Gi |
| large | 2 | 4Gi |
| xlarge | 4 | 8Gi |

### List Databases

```http
GET /projects/{project_id}/databases
```

### Get Database

```http
GET /databases/{id}
```

### Delete Database

```http
DELETE /databases/{id}
```

### Scale Database

```http
POST /databases/{id}/scale
```

**Request Body:**
```json
{"replicas": 5}
```

### Get Connection Info

```http
GET /databases/{id}/connection
```

**Response:**
```json
{
  "ysql_endpoint": "db-yb-tserver.svc:5433",
  "ycql_endpoint": "db-yb-tserver.svc:9042",
  "secret_name": "db-credentials"
}
```

---

## Clusters

### Create Cluster

```http
POST /clusters
```

**Request Body:**
```json
{
  "name": "production",
  "provider": "rancher",
  "region": "us-east-1",
  "kube_version": "1.28",
  "node_count": 5,
  "labels": {"tier": "production"}
}
```

### Providers

| Provider | Description |
|----------|-------------|
| rancher | Rancher-managed |
| rke2 | RKE2 Kubernetes |
| k3s | Lightweight K3s |
| eks | AWS EKS |
| gke | Google GKE |
| aks | Azure AKS |

### List Clusters

```http
GET /clusters
```

### Get Cluster

```http
GET /clusters/{id}
```

### Delete Cluster

```http
DELETE /clusters/{id}
```

### Get Kubeconfig

```http
GET /clusters/{id}/kubeconfig
```

---

## Authentication

### Login

```http
POST /auth/login
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_at": "2026-01-17T00:00:00Z",
  "user": {...}
}
```

### Register

```http
POST /auth/register
```

### Refresh Token

```http
POST /auth/refresh
```

### Logout

```http
POST /auth/logout
```

### Get Current User

```http
GET /users/me
```

---

## Webhooks

### GitHub Webhook

```http
POST /webhooks/github
```

Automatically triggered by GitHub for:
- Push events
- Pull request events

---

## Health Checks

### Liveness

```http
GET /health/live
```

### Readiness

```http
GET /health/ready
```

---

## Error Responses

All errors return:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {...}
}
```

### Status Codes

| Code | Meaning |
|------|---------|
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 500 | Internal Error |

---

## Rate Limits

- 100 requests per minute per user
- 1000 requests per hour per user

Headers returned:
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`

---

## GraphQL API

Available at: `https://graphql.northstack.io/v1/graphql`

Uses Hasura with JWT authentication.

```graphql
query GetProjects {
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
```
