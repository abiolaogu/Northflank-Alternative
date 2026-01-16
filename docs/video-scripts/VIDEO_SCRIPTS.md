# NorthStack Platform - Video Training Scripts

## Video Series Overview

| Video | Duration | Audience |
|-------|----------|----------|
| 1. Platform Introduction | 5 min | All users |
| 2. Getting Started | 8 min | Developers |
| 3. Deploying Applications | 12 min | Developers |
| 4. Database Management | 10 min | All users |
| 5. CI/CD Pipelines | 10 min | DevOps |
| 6. Cluster Administration | 15 min | Admins |
| 7. Monitoring & Troubleshooting | 10 min | All users |

---

## Video 1: Platform Introduction

**Duration:** 5 minutes

### Script

```
[OPENING - NorthStack logo animation]

NARRATOR:
"Welcome to NorthStack, your internal platform for deploying and managing 
applications at scale.

[SCREEN: Architecture diagram]

NorthStack combines the simplicity of platforms like Heroku with the power 
of Kubernetes. Deploy your code with a simple API call, and let the platform 
handle scaling, load balancing, and high availability.

[SCREEN: Key features]

Key features include:
- One-click deployments from Git
- YugabyteDB distributed databases
- Automatic CI/CD pipelines
- Real-time GraphQL APIs
- Multi-cluster management

[SCREEN: User personas]

Whether you're a developer deploying your first app, a team lead monitoring 
production, or a platform admin managing clusters - NorthStack has you covered.

Let's get started!"

[END CARD: "Next: Getting Started"]
```

---

## Video 2: Getting Started

**Duration:** 8 minutes

### Script

```
[OPENING]

NARRATOR:
"In this video, you'll learn how to set up your NorthStack account and 
deploy your first application."

[SCREEN: Login page]

"First, navigate to your NorthStack instance and log in with your 
organization credentials."

[SCREEN: API token generation]

"For API access, generate a token from your profile settings. 
Keep this token secure - it grants full access to your account."

[SCREEN: Terminal]

"Let's test our connection:

curl -H 'Authorization: Bearer YOUR_TOKEN' \
  https://api.northstack.io/health

You should see a healthy response."

[SCREEN: Create project]

"Now let's create our first project:

curl -X POST https://api.northstack.io/api/v1/projects \
  -d '{\"name\": \"my-first-app\"}'

A project is a container for related services - like a frontend, 
backend, and database."

[SCREEN: Success response]

"Congratulations! You've created your first project. 
In the next video, we'll deploy an application to it."

[END CARD]
```

---

## Video 3: Deploying Applications

**Duration:** 12 minutes

### Script

```
[OPENING]

NARRATOR:
"Now let's deploy a real application to NorthStack."

[SCREEN: GitHub repo]

"We'll deploy this Node.js application from GitHub. 
The repo has a Dockerfile that NorthStack will use to build the image."

[SCREEN: API call]

"Create a service using the API:

curl -X POST /api/v1/projects/{id}/services \
  -d '{
    \"name\": \"api-server\",
    \"type\": \"webapp\",
    \"build_source\": {
      \"type\": \"git\",
      \"repository\": \"github.com/org/app\",
      \"branch\": \"main\"
    }
  }'"

[SCREEN: Build progress]

"NorthStack triggers a build automatically. You can watch the progress 
in real-time. The build pulls your code, runs the Dockerfile, and 
pushes the image to our registry."

[SCREEN: Deployment]

"Once built, ArgoCD deploys your application to Kubernetes. 
Health checks ensure the app is running before routing traffic."

[SCREEN: Running app]

"Your app is now live! Access it at the generated URL.
Let's add some environment variables..."

[SCREEN: Env vars]

"curl -X PATCH /api/v1/services/{id} \
  -d '{\"env_vars\": {\"LOG_LEVEL\": \"info\"}}'

The service automatically restarts with the new configuration."

[SCREEN: Scaling]

"Need more capacity? Scale with a single call:

curl -X POST /api/v1/services/{id}/scale \
  -d '{\"replicas\": 5}'

Your app now runs on 5 instances with automatic load balancing."

[END CARD]
```

---

## Video 4: Database Management

**Duration:** 10 minutes

### Script

```
[OPENING]

NARRATOR:
"NorthStack uses YugabyteDB for databases - a distributed SQL database 
that's PostgreSQL compatible."

[SCREEN: YugabyteDB architecture]

"Unlike traditional PostgreSQL, YugabyteDB:
- Distributes data across multiple nodes
- Provides automatic failover
- Scales horizontally
- Maintains PostgreSQL compatibility"

[SCREEN: Create database]

"Let's create a production database:

curl -X POST /api/v1/projects/{id}/databases \
  -d '{
    \"name\": \"production\",
    \"size\": \"large\",
    \"storage_gb\": 100,
    \"high_availability\": true
  }'"

[SCREEN: Connection info]

"Get your connection details:

curl /api/v1/databases/{id}/connection

The response includes the YSQL endpoint (PostgreSQL compatible) 
and credentials secret name."

[SCREEN: Code example]

"Connect from your application using any PostgreSQL driver:

const pool = new Pool({
  connectionString: process.env.DATABASE_URL
});

await pool.query('SELECT * FROM users');"

[SCREEN: Scaling]

"Scale your database as traffic grows:

curl -X POST /api/v1/databases/{id}/scale \
  -d '{\"replicas\": 5}'

Data is automatically rebalanced across nodes."

[END CARD]
```

---

## Video 5: CI/CD Pipelines

**Duration:** 10 minutes

### Script

```
[OPENING]

NARRATOR:
"Automate your deployments with NorthStack's CI/CD integration."

[SCREEN: GitHub Actions workflow]

"Add this workflow to your repository:

name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy
        run: |
          curl -X POST /api/v1/services/{id}/builds"

[SCREEN: Webhook flow diagram]

"When you push to main:
1. GitHub triggers the workflow
2. NorthStack builds your container
3. Trivy scans for vulnerabilities
4. ArgoCD deploys to production
5. Slack sends notification"

[SCREEN: Preview environments]

"Pull requests get automatic preview environments:
- Unique URL for each PR
- Isolated database
- Auto-cleanup on merge

Test changes before they hit production."

[END CARD]
```

---

## Video 6: Cluster Administration

**Duration:** 15 minutes

### Script

```
[OPENING]

NARRATOR:
"This video covers cluster management for platform administrators."

[SCREEN: Cluster providers]

"NorthStack supports multiple Kubernetes providers:
- Rancher for on-premise
- RKE2 for production
- K3s for development
- EKS, GKE, AKS for cloud"

[SCREEN: Create cluster]

"Create a production cluster:

curl -X POST /api/v1/clusters \
  -d '{
    \"name\": \"production\",
    \"provider\": \"rancher\",
    \"region\": \"us-east-1\",
    \"node_count\": 5
  }'"

[SCREEN: Node management]

"Scale nodes as needed:

curl -X PATCH /api/v1/clusters/{id} \
  -d '{\"node_count\": 10}'"

[SCREEN: Kubeconfig]

"Download kubeconfig for direct cluster access:

curl /api/v1/clusters/{id}/kubeconfig > ~/.kube/prod"

[SCREEN: Multi-cluster]

"Route services to specific clusters for:
- Geographic distribution
- Environment isolation
- Compliance requirements"

[END CARD]
```

---

## Video 7: Monitoring & Troubleshooting

**Duration:** 10 minutes

### Script

```
[OPENING]

NARRATOR:
"Learn how to monitor your applications and troubleshoot issues."

[SCREEN: Prometheus metrics]

"NorthStack exposes Prometheus metrics at /metrics:
- http_requests_total - Request count
- http_request_duration_seconds - Latency
- go_goroutines - Resource usage"

[SCREEN: Grafana dashboard]

"View pre-built dashboards in Grafana showing:
- Request rates
- Error rates
- Response times
- Resource utilization"

[SCREEN: Logs]

"Access service logs via API:

curl /api/v1/services/{id}/logs?tail=100"

[SCREEN: Troubleshooting table]

"Common issues and solutions:
- CrashLoopBackOff → Check application logs
- OOMKilled → Increase memory limits
- ImagePullBackOff → Verify registry credentials"

[SCREEN: Alerts]

"Configure Slack alerts for critical issues like:
- High error rate
- Service down
- Database connection failures"

[END CARD: "Thank you for completing NorthStack training!"]
```

---

## Production Notes

### Equipment Needed
- Screen recording software (OBS, Camtasia)
- Microphone with pop filter
- NorthStack test environment
- Sample applications for demos

### Style Guide
- Use consistent color scheme
- Include captions/subtitles
- Add chapter markers
- End each video with call-to-action
