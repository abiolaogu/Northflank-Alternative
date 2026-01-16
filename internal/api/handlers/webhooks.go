package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/errors"
	"github.com/northstack/platform/pkg/logger"
)

// WebhookHandler handles incoming webhooks from external services
type WebhookHandler struct {
	eventBus      domain.EventBus
	serviceRepo   domain.ServiceRepository
	webhookSecret string
	logger        *logger.Logger
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(
	eventBus domain.EventBus,
	serviceRepo domain.ServiceRepository,
	webhookSecret string,
	log *logger.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		eventBus:      eventBus,
		serviceRepo:   serviceRepo,
		webhookSecret: webhookSecret,
		logger:        log,
	}
}

// HandleGitHub handles GitHub webhooks
func (h *WebhookHandler) HandleGitHub(c *gin.Context) {
	// Verify signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if h.webhookSecret != "" && !h.verifyGitHubSignature(c, signature) {
		respondError(c, errors.Unauthorized("invalid webhook signature"))
		return
	}

	eventType := c.GetHeader("X-GitHub-Event")
	deliveryID := c.GetHeader("X-GitHub-Delivery")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, errors.BadRequest("failed to read request body"))
		return
	}

	// Publish webhook event for processing
	event := &domain.Event{
		Type:   "webhook.github." + eventType,
		Source: "github",
		Data: map[string]interface{}{
			"delivery_id": deliveryID,
			"event_type":  eventType,
			"payload":     string(body),
		},
		Metadata: map[string]string{
			"source": "github",
		},
	}

	if err := h.eventBus.Publish(c.Request.Context(), "webhook.received", event); err != nil {
		h.logger.Error().Err(err).Str("event_type", eventType).Msg("Failed to publish webhook event")
	}

	h.logger.Info().
		Str("event_type", eventType).
		Str("delivery_id", deliveryID).
		Msg("GitHub webhook received")

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// HandleGitLab handles GitLab webhooks
func (h *WebhookHandler) HandleGitLab(c *gin.Context) {
	token := c.GetHeader("X-Gitlab-Token")
	if h.webhookSecret != "" && token != h.webhookSecret {
		respondError(c, errors.Unauthorized("invalid webhook token"))
		return
	}

	eventType := c.GetHeader("X-Gitlab-Event")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, errors.BadRequest("failed to read request body"))
		return
	}

	event := &domain.Event{
		Type:   "webhook.gitlab." + eventType,
		Source: "gitlab",
		Data: map[string]interface{}{
			"event_type": eventType,
			"payload":    string(body),
		},
		Metadata: map[string]string{
			"source": "gitlab",
		},
	}

	if err := h.eventBus.Publish(c.Request.Context(), "webhook.received", event); err != nil {
		h.logger.Error().Err(err).Str("event_type", eventType).Msg("Failed to publish webhook event")
	}

	h.logger.Info().
		Str("event_type", eventType).
		Msg("GitLab webhook received")

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// HandleCoolify handles Coolify webhooks
func (h *WebhookHandler) HandleCoolify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, errors.BadRequest("failed to read request body"))
		return
	}

	event := &domain.Event{
		Type:   "webhook.coolify.build",
		Source: "coolify",
		Data: map[string]interface{}{
			"payload": string(body),
		},
		Metadata: map[string]string{
			"source": "coolify",
		},
	}

	if err := h.eventBus.Publish(c.Request.Context(), "webhook.received", event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to publish Coolify webhook event")
	}

	h.logger.Info().Msg("Coolify webhook received")
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// HandleArgoCD handles ArgoCD webhooks
func (h *WebhookHandler) HandleArgoCD(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, errors.BadRequest("failed to read request body"))
		return
	}

	event := &domain.Event{
		Type:   "webhook.argocd.sync",
		Source: "argocd",
		Data: map[string]interface{}{
			"payload": string(body),
		},
		Metadata: map[string]string{
			"source": "argocd",
		},
	}

	if err := h.eventBus.Publish(c.Request.Context(), "webhook.received", event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to publish ArgoCD webhook event")
	}

	h.logger.Info().Msg("ArgoCD webhook received")
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// HandleGeneric handles generic webhooks
func (h *WebhookHandler) HandleGeneric(c *gin.Context) {
	source := c.Param("source")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, errors.BadRequest("failed to read request body"))
		return
	}

	event := &domain.Event{
		Type:   "webhook." + source,
		Source: source,
		Data: map[string]interface{}{
			"payload": string(body),
			"headers": c.Request.Header,
		},
		Metadata: map[string]string{
			"source": source,
		},
	}

	if err := h.eventBus.Publish(c.Request.Context(), "webhook.received", event); err != nil {
		h.logger.Error().Err(err).Str("source", source).Msg("Failed to publish webhook event")
	}

	h.logger.Info().Str("source", source).Msg("Generic webhook received")
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// verifyGitHubSignature verifies the GitHub webhook signature
func (h *WebhookHandler) verifyGitHubSignature(c *gin.Context, signature string) bool {
	if signature == "" {
		return false
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}

	// Reset body for subsequent reads
	c.Request.Body = io.NopCloser(c.Request.Body)

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
