package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/northstack/platform/pkg/logger"
)

// GitHubWebhookHandler handles GitHub webhook events
type GitHubWebhookHandler struct {
	secret string
	logger *logger.Logger
}

// NewGitHubWebhookHandler creates a new GitHub webhook handler
func NewGitHubWebhookHandler(secret string, log *logger.Logger) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{
		secret: secret,
		logger: log,
	}
}

// HandleWebhook processes incoming GitHub webhooks
func (h *GitHubWebhookHandler) HandleWebhook(c *gin.Context) {
	// Read body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if !h.verifySignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")
	delivery := c.GetHeader("X-GitHub-Delivery")

	h.logger.Info().
		Str("event", event).
		Str("delivery", delivery).
		Msg("Received GitHub webhook")

	switch event {
	case "pull_request":
		h.handlePullRequest(c, body)
	case "push":
		h.handlePush(c, body)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
	}
}

// PullRequestEvent represents a GitHub PR event
type PullRequestEvent struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		Head struct {
			SHA string `json:"sha"`
			Ref string `json:"ref"`
		} `json:"head"`
		Title  string `json:"title"`
		Merged bool   `json:"merged"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
		Name     string `json:"name"`
	} `json:"repository"`
}

func (h *GitHubWebhookHandler) handlePullRequest(c *gin.Context, body []byte) {
	var event PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	h.logger.Info().
		Str("action", event.Action).
		Int("pr", event.Number).
		Str("repo", event.Repository.FullName).
		Msg("Processing pull request event")

	switch event.Action {
	case "opened", "synchronize", "reopened":
		// Preview deployment is handled by GitHub Actions workflow
		c.JSON(http.StatusOK, gin.H{
			"message": "PR event received",
			"pr":      event.Number,
			"action":  "deploy_preview",
		})
	case "closed":
		// Cleanup is handled by GitHub Actions workflow
		c.JSON(http.StatusOK, gin.H{
			"message": "PR closed event received",
			"pr":      event.Number,
			"action":  "cleanup",
		})
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Action ignored"})
	}
}

func (h *GitHubWebhookHandler) handlePush(c *gin.Context, body []byte) {
	var event struct {
		Ref        string `json:"ref"`
		After      string `json:"after"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Only process main branch
	if event.Ref != "refs/heads/main" {
		c.JSON(http.StatusOK, gin.H{"message": "Branch ignored"})
		return
	}

	h.logger.Info().
		Str("repo", event.Repository.FullName).
		Str("sha", event.After).
		Msg("Processing push to main")

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Push event received",
		"sha":     event.After,
		"action":  "deploy_production",
	})
}

func (h *GitHubWebhookHandler) verifySignature(payload []byte, signature string) bool {
	if h.secret == "" {
		return true // No secret configured, skip verification
	}

	if signature == "" || len(signature) < 8 {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}
