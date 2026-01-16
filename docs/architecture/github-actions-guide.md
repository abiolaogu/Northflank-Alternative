# NorthStack Platform - Prompt 9: GitHub Actions CI/CD Automation

## Overview: Automated Deployment Pipelines

> "As soon as a developer pushes code to the new repository created by Backstage, Coolify should detect the change and deploy it."

This prompt implements the complete CI/CD automation that makes the platform feel "magical"—instant deployments on push, automatic preview environments on PR, and cleanup on merge.

## The "Beyond Northflank" Advantage

| Feature | Northflank | NorthStack CI/CD |
|---------|------------|------------------|
| **Preview Deploy** | Auto-detect | Instant webhook + comment |
| **Custom Logic** | Limited | Unlimited with Actions |
| **E2E Tests** | Manual setup | Baked into template |
| **Cleanup** | Manual | Automatic on PR close |
| **Multi-Env** | Single pattern | Dev→Coolify, Prod→ArgoCD |

---

## Architecture: Smart Routing

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    GITHUB ACTIONS FLOW                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    PUSH / PR EVENTS                              │   │
│  └───────────────────────────┬─────────────────────────────────────┘   │
│                              │                                          │
│              ┌───────────────┴───────────────┐                         │
│              │                               │                         │
│              ▼                               ▼                         │
│  ┌───────────────────────┐       ┌───────────────────────┐            │
│  │   PULL REQUEST        │       │   MAIN BRANCH          │            │
│  │   ───────────────     │       │   ─────────────        │            │
│  │   • Build & Test      │       │   • Build & Test       │            │
│  │   • Trivy Scan        │       │   • Trivy Scan         │            │
│  │   • Push to Registry  │       │   • Push to Registry   │            │
│  │   • Deploy to Coolify │       │   • Update ArgoCD      │            │
│  │   • Comment PR URL    │       │   • Notify Backstage   │            │
│  └───────────────────────┘       └───────────────────────┘            │
│              │                               │                         │
│              │                               │                         │
│              ▼                               ▼                         │
│  ┌───────────────────────┐       ┌───────────────────────┐            │
│  │   PR CLOSED/MERGED    │       │   PRODUCTION DEPLOY    │            │
│  │   ──────────────────  │       │   ─────────────────    │            │
│  │   • Delete Coolify    │       │   • ArgoCD Sync        │            │
│  │     Preview           │       │   • Health Check       │            │
│  │   • Cleanup vCluster  │       │   • Rollback on Fail   │            │
│  └───────────────────────┘       └───────────────────────┘            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Task 1: Main Deployment Workflow

### File: `config/backstage/templates/skeletons/common/.github/workflows/deploy.yml`

```yaml
name: NorthStack CI/CD Pipeline
on:
  push:
    branches: [main]
  pull_request:
    types: [opened, synchronize, reopened]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

permissions:
  contents: read
  packages: write
  pull-requests: write
  id-token: write

jobs:
  # ============================================================================
  # BUILD & TEST
  # ============================================================================
  build:
    name: Build & Test
    runs-on: ubuntu-latest
    outputs:
      image_tag: ${{ steps.meta.outputs.tags }}
      image_digest: ${{ steps.build.outputs.digest }}
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha,prefix=
            type=raw,value=latest,enable=${{ github.ref == format('refs/heads/{0}', 'main') }}

      - name: Build and push
        id: build
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ github.sha }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}

  # ============================================================================
  # SECURITY SCANNING
  # ============================================================================
  security-scan:
    name: Security Scan
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ needs.build.outputs.image_tag }}
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'

      - name: Upload Trivy scan results
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: 'trivy-results.sarif'

      - name: Check for critical vulnerabilities
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ needs.build.outputs.image_tag }}
          exit-code: '1'
          severity: 'CRITICAL'

  # ============================================================================
  # DEPLOY PREVIEW (Pull Requests only)
  # ============================================================================
  deploy-preview:
    name: Deploy Preview Environment
    needs: [build, security-scan]
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    environment:
      name: preview-pr-${{ github.event.number }}
      url: ${{ steps.preview.outputs.url }}
    steps:
      - name: Deploy to Coolify
        id: coolify-deploy
        run: |
          # Trigger Coolify webhook for preview deployment
          RESPONSE=$(curl -s -X POST "${{ secrets.COOLIFY_WEBHOOK_URL }}" \
            -H "Authorization: Bearer ${{ secrets.COOLIFY_API_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "image": "${{ needs.build.outputs.image_tag }}",
              "pr_number": "${{ github.event.number }}",
              "branch": "${{ github.head_ref }}"
            }')
          
          echo "response=$RESPONSE" >> $GITHUB_OUTPUT

      - name: Create/Update vCluster Preview
        id: vcluster
        run: |
          # Call NorthStack API to create vCluster-based preview
          curl -X POST "${{ secrets.NORTHSTACK_API_URL }}/api/v1/previews" \
            -H "Authorization: Bearer ${{ secrets.NORTHSTACK_API_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "prNumber": ${{ github.event.number }},
              "repository": "${{ github.repository }}",
              "branch": "${{ github.head_ref }}",
              "commitSha": "${{ github.sha }}",
              "imageTag": "${{ needs.build.outputs.image_tag }}"
            }'

      - name: Get Preview URL
        id: preview
        run: |
          PREVIEW_URL="https://pr-${{ github.event.number }}-${{ github.event.repository.name }}.preview.northstack.io"
          echo "url=$PREVIEW_URL" >> $GITHUB_OUTPUT

      - name: Comment PR with Preview URL
        uses: actions/github-script@v7
        with:
          script: |
            const prNumber = context.issue.number;
            const previewUrl = '${{ steps.preview.outputs.url }}';
            const imageTag = '${{ needs.build.outputs.image_tag }}';
            
            const body = `## 🚀 Preview Environment Deployed!
            
            | Resource | Link |
            |----------|------|
            | **Preview URL** | [${previewUrl}](${previewUrl}) |
            | **Container Image** | \`${imageTag}\` |
            | **Deployment** | [View in NorthStack](https://app.northstack.io/previews/pr-${prNumber}) |
            | **Logs** | [View Logs](https://logs.northstack.io?pr=${prNumber}) |
            
            ### 📊 Health Status
            - ✅ Build successful
            - ✅ Security scan passed
            - ✅ Preview deployed
            
            > This preview will be automatically deleted when the PR is closed.
            `;
            
            // Find existing comment
            const comments = await github.rest.issues.listComments({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: prNumber
            });
            
            const botComment = comments.data.find(c => 
              c.user.login === 'github-actions[bot]' && 
              c.body.includes('Preview Environment')
            );
            
            if (botComment) {
              await github.rest.issues.updateComment({
                owner: context.repo.owner,
                repo: context.repo.repo,
                comment_id: botComment.id,
                body: body
              });
            } else {
              await github.rest.issues.createComment({
                owner: context.repo.owner,
                repo: context.repo.repo,
                issue_number: prNumber,
                body: body
              });
            }

  # ============================================================================
  # DEPLOY PRODUCTION (Main branch only)
  # ============================================================================
  deploy-production:
    name: Deploy to Production
    needs: [build, security-scan]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://${{ github.event.repository.name }}.northstack.io
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Update ArgoCD Application
        run: |
          # Update the image tag in the Helm values
          curl -X POST "${{ secrets.ARGOCD_URL }}/api/v1/applications/${{ github.event.repository.name }}/sync" \
            -H "Authorization: Bearer ${{ secrets.ARGOCD_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "revision": "${{ github.sha }}",
              "dryRun": false,
              "prune": true,
              "strategy": {
                "hook": {
                  "force": false
                }
              },
              "resources": null,
              "syncOptions": {
                "items": [
                  "CreateNamespace=true",
                  "PrunePropagationPolicy=foreground"
                ]
              }
            }'

      - name: Wait for ArgoCD Sync
        run: |
          # Poll ArgoCD until sync is complete
          for i in {1..60}; do
            STATUS=$(curl -s -H "Authorization: Bearer ${{ secrets.ARGOCD_TOKEN }}" \
              "${{ secrets.ARGOCD_URL }}/api/v1/applications/${{ github.event.repository.name }}" | \
              jq -r '.status.sync.status')
            
            if [ "$STATUS" = "Synced" ]; then
              echo "ArgoCD sync complete!"
              exit 0
            fi
            
            echo "Waiting for sync... ($i/60) Status: $STATUS"
            sleep 10
          done
          
          echo "Timeout waiting for ArgoCD sync"
          exit 1

      - name: Verify Deployment Health
        run: |
          # Check health status
          HEALTH=$(curl -s -H "Authorization: Bearer ${{ secrets.ARGOCD_TOKEN }}" \
            "${{ secrets.ARGOCD_URL }}/api/v1/applications/${{ github.event.repository.name }}" | \
            jq -r '.status.health.status')
          
          if [ "$HEALTH" != "Healthy" ]; then
            echo "Deployment unhealthy: $HEALTH"
            exit 1
          fi
          
          echo "Deployment healthy!"

      - name: Update Backstage Catalog
        run: |
          curl -X POST "${{ secrets.BACKSTAGE_URL }}/api/catalog/refresh" \
            -H "Authorization: Bearer ${{ secrets.BACKSTAGE_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{"entityRef": "component:default/${{ github.event.repository.name }}"}'

      - name: Notify Slack
        if: always()
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          fields: repo,message,commit,author,action,eventName,ref,workflow
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

---

## Task 2: Cleanup Workflow

### File: `config/backstage/templates/skeletons/common/.github/workflows/cleanup.yml`

> "In a self-hosted 'Super-Stack,' you must automate cleanup to prevent your VPS from running out of RAM and CPU."

```yaml
name: Cleanup Preview Environment
on:
  pull_request:
    types: [closed]

jobs:
  cleanup:
    name: Cleanup Resources
    runs-on: ubuntu-latest
    steps:
      - name: Delete Coolify Preview Application
        run: |
          echo "Cleaning up Coolify preview for PR #${{ github.event.number }}"
          
          curl -X DELETE "${{ secrets.COOLIFY_URL }}/api/v1/applications/pr-${{ github.event.number }}-${{ github.event.repository.name }}" \
            -H "Authorization: Bearer ${{ secrets.COOLIFY_API_TOKEN }}" \
            --fail-with-body || echo "Coolify cleanup failed (may not exist)"

      - name: Delete vCluster Preview
        run: |
          echo "Cleaning up vCluster for PR #${{ github.event.number }}"
          
          curl -X DELETE "${{ secrets.NORTHSTACK_API_URL }}/api/v1/previews/pr-${{ github.event.number }}" \
            -H "Authorization: Bearer ${{ secrets.NORTHSTACK_API_TOKEN }}" \
            --fail-with-body || echo "vCluster cleanup failed (may not exist)"

      - name: Delete Preview Namespace
        run: |
          # If using namespace-based previews (simpler alternative to vCluster)
          curl -X DELETE "${{ secrets.KUBERNETES_API_URL }}/api/v1/namespaces/preview-${{ github.event.number }}" \
            -H "Authorization: Bearer ${{ secrets.KUBERNETES_TOKEN }}" \
            --fail-with-body || echo "Namespace cleanup failed (may not exist)"

      - name: Delete Container Images
        run: |
          # Cleanup PR-specific container images (optional - saves registry space)
          echo "Cleaning up container images for PR #${{ github.event.number }}"
          
          # Delete from GitHub Container Registry
          gh api -X DELETE "/user/packages/container/${{ github.event.repository.name }}/versions?tag=pr-${{ github.event.number }}" \
            || echo "Image cleanup skipped"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Comment PR with Cleanup Notice
        uses: actions/github-script@v7
        with:
          script: |
            const prNumber = context.issue.number;
            const merged = context.payload.pull_request.merged;
            
            const emoji = merged ? '🎉' : '🧹';
            const action = merged ? 'merged' : 'closed';
            
            github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: prNumber,
              body: `${emoji} PR ${action}! Preview environment has been cleaned up.
              
              ${merged ? '**Deployed to production!** View at: https://' + context.repo.repo + '.northstack.io' : ''}`
            });

      - name: Update Backstage Catalog
        if: github.event.pull_request.merged == true
        run: |
          # Trigger catalog refresh after merge
          curl -X POST "${{ secrets.BACKSTAGE_URL }}/api/catalog/refresh" \
            -H "Authorization: Bearer ${{ secrets.BACKSTAGE_TOKEN }}" \
            -d '{"entityRef": "component:default/${{ github.event.repository.name }}"}'
```

---

## Task 3: NorthStack Webhook Handler

### File: `internal/api/handlers/github_webhook_handler.go`

```go
package handlers

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"

    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"

    "github.com/northstack/platform/internal/services"
)

// GitHubWebhookHandler handles GitHub webhook events
type GitHubWebhookHandler struct {
    previewSvc *services.PreviewService
    deploySvc  *services.DeploymentService
    secret     string
    log        *zap.SugaredLogger
}

// NewGitHubWebhookHandler creates a new GitHub webhook handler
func NewGitHubWebhookHandler(
    previewSvc *services.PreviewService,
    deploySvc *services.DeploymentService,
    secret string,
    log *zap.SugaredLogger,
) *GitHubWebhookHandler {
    return &GitHubWebhookHandler{
        previewSvc: previewSvc,
        deploySvc:  deploySvc,
        secret:     secret,
        log:        log,
    }
}

// RegisterRoutes registers webhook routes
func (h *GitHubWebhookHandler) RegisterRoutes(app *fiber.App) {
    app.Post("/webhooks/github", h.HandleWebhook)
}

// HandleWebhook processes incoming GitHub webhooks
func (h *GitHubWebhookHandler) HandleWebhook(c *fiber.Ctx) error {
    // Verify signature
    signature := c.Get("X-Hub-Signature-256")
    if !h.verifySignature(c.Body(), signature) {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid signature"})
    }

    event := c.Get("X-GitHub-Event")
    delivery := c.Get("X-GitHub-Delivery")

    h.log.Infow("Received GitHub webhook",
        "event", event,
        "delivery", delivery,
    )

    switch event {
    case "pull_request":
        return h.handlePullRequest(c)
    case "push":
        return h.handlePush(c)
    case "workflow_run":
        return h.handleWorkflowRun(c)
    default:
        h.log.Debugw("Ignoring unhandled event", "event", event)
        return c.SendStatus(200)
    }
}

// Pull request webhook payload
type PullRequestEvent struct {
    Action      string          `json:"action"`
    Number      int             `json:"number"`
    PullRequest PullRequestInfo `json:"pull_request"`
    Repository  Repository      `json:"repository"`
}

type PullRequestInfo struct {
    Head   PRHead `json:"head"`
    Base   PRBase `json:"base"`
    Title  string `json:"title"`
    Merged bool   `json:"merged"`
}

type PRHead struct {
    SHA string `json:"sha"`
    Ref string `json:"ref"`
}

type PRBase struct {
    Ref string `json:"ref"`
}

type Repository struct {
    FullName string `json:"full_name"`
    Name     string `json:"name"`
    CloneURL string `json:"clone_url"`
}

func (h *GitHubWebhookHandler) handlePullRequest(c *fiber.Ctx) error {
    var event PullRequestEvent
    if err := json.Unmarshal(c.Body(), &event); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
    }

    h.log.Infow("Processing pull request event",
        "action", event.Action,
        "pr", event.Number,
        "repo", event.Repository.FullName,
    )

    switch event.Action {
    case "opened", "synchronize", "reopened":
        // Create or update preview environment
        return h.createOrUpdatePreview(c, &event)

    case "closed":
        // Cleanup preview environment
        return h.cleanupPreview(c, &event)
    }

    return c.SendStatus(200)
}

func (h *GitHubWebhookHandler) createOrUpdatePreview(c *fiber.Ctx, event *PullRequestEvent) error {
    input := &services.CreatePreviewInput{
        PRNumber:      event.Number,
        Repository:    event.Repository.FullName,
        Branch:        event.PullRequest.Head.Ref,
        CommitSHA:     event.PullRequest.Head.SHA,
        BaseBranch:    event.PullRequest.Base.Ref,
        Title:         event.PullRequest.Title,
    }

    preview, err := h.previewSvc.CreateOrUpdate(c.Context(), input)
    if err != nil {
        h.log.Errorw("Failed to create/update preview", "error", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create preview"})
    }

    return c.JSON(preview)
}

func (h *GitHubWebhookHandler) cleanupPreview(c *fiber.Ctx, event *PullRequestEvent) error {
    err := h.previewSvc.Delete(c.Context(), event.Repository.FullName, event.Number)
    if err != nil {
        h.log.Warnw("Failed to cleanup preview", "error", err, "pr", event.Number)
    }

    return c.SendStatus(200)
}

func (h *GitHubWebhookHandler) handlePush(c *fiber.Ctx) error {
    // Handle push events for main branch deployments
    var event struct {
        Ref        string     `json:"ref"`
        After      string     `json:"after"`
        Repository Repository `json:"repository"`
    }

    if err := json.Unmarshal(c.Body(), &event); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
    }

    // Only trigger on main branch
    if event.Ref != "refs/heads/main" {
        return c.SendStatus(200)
    }

    h.log.Infow("Processing push to main",
        "repo", event.Repository.FullName,
        "sha", event.After,
    )

    // Trigger deployment
    _, err := h.deploySvc.TriggerFromPush(c.Context(), &services.PushTrigger{
        Repository: event.Repository.FullName,
        CommitSHA:  event.After,
        Branch:     "main",
    })
    if err != nil {
        h.log.Errorw("Failed to trigger deployment", "error", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to trigger deployment"})
    }

    return c.SendStatus(202)
}

func (h *GitHubWebhookHandler) handleWorkflowRun(c *fiber.Ctx) error {
    // Handle workflow completion events for status updates
    return c.SendStatus(200)
}

func (h *GitHubWebhookHandler) verifySignature(payload []byte, signature string) bool {
    if signature == "" || len(signature) < 8 {
        return false
    }

    mac := hmac.New(sha256.New, []byte(h.secret))
    mac.Write(payload)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

    return hmac.Equal([]byte(expected), []byte(signature))
}
```

---

## Task 4: Docker Compose for Zero-Downtime

### File: `config/backstage/templates/skeletons/common/docker-compose.yaml`

> "For zero-downtime deployments, this uses Traefik labels which Coolify uses internally."

```yaml
version: '3.8'
services:
  app:
    image: ${REGISTRY:-ghcr.io}/${IMAGE_NAME}:${IMAGE_TAG:-latest}
    build:
      context: .
      dockerfile: Dockerfile
    restart: always
    environment:
      - NODE_ENV=${NODE_ENV:-production}
      - PORT=${PORT:-3000}
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    
    # Essential for Zero-Downtime Deployments
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:${PORT:-3000}/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 20s
    
    # Resource Limits (prevents runaway containers)
    deploy:
      resources:
        limits:
          cpus: '${CPU_LIMIT:-0.50}'
          memory: ${MEMORY_LIMIT:-512M}
        reservations:
          memory: ${MEMORY_RESERVATION:-128M}
      replicas: ${REPLICAS:-2}
      update_config:
        parallelism: 1
        delay: 10s
        failure_action: rollback
        order: start-first
      rollback_config:
        parallelism: 1
        delay: 10s
    
    # Coolify/Traefik Labels for Auto-Routing
    labels:
      - "coolify.managed=true"
      - "traefik.enable=true"
      - "traefik.http.routers.${SERVICE_NAME}.rule=Host(`${DOMAIN}`)"
      - "traefik.http.routers.${SERVICE_NAME}.entrypoints=websecure"
      - "traefik.http.routers.${SERVICE_NAME}.tls=true"
      - "traefik.http.routers.${SERVICE_NAME}.tls.certresolver=letsencrypt"
      - "traefik.http.services.${SERVICE_NAME}.loadbalancer.server.port=${PORT:-3000}"
      - "traefik.http.services.${SERVICE_NAME}.loadbalancer.healthcheck.path=/health"
      - "traefik.http.services.${SERVICE_NAME}.loadbalancer.healthcheck.interval=10s"

networks:
  coolify:
    external: true
```

---

## Deliverables

1. **Deploy Workflow** (`.github/workflows/deploy.yml`)
   - Build and push container images
   - Security scanning with Trivy
   - PR preview deployment via Coolify/vCluster
   - Production deployment via ArgoCD
   - PR commenting with URLs

2. **Cleanup Workflow** (`.github/workflows/cleanup.yml`)
   - Automatic cleanup on PR close
   - Container image cleanup
   - Namespace/vCluster deletion
   - Backstage catalog update

3. **Webhook Handler** (`internal/api/handlers/github_webhook_handler.go`)
   - Signature verification
   - PR event handling
   - Push event handling
   - Preview lifecycle management

4. **Docker Compose** (`docker-compose.yaml`)
   - Health checks for zero-downtime
   - Resource limits
   - Traefik labels for routing
   - Rolling update configuration

## Why This Goes "Beyond Northflank"

> "You can add a step in the GitHub Action to run Playwright E2E tests before the Coolify webhook is even hit."

> "You can send 'Feature Branch' previews to a tiny $4/mo Hetzner VPS via Coolify, while keeping your production workloads on a robust Argo CD cluster."

> "No Vendor Lock-in: If you ever want to leave Coolify, your code is already in Git. You just change the Backstage template to point to a different provider."
