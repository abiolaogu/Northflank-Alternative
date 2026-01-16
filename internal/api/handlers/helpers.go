package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openpaas/platform-orchestrator/pkg/errors"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// respondError sends an error response to the client
func respondError(c *gin.Context, err error) {
	pe := errors.GetPlatformError(err)

	c.JSON(pe.HTTPStatus, ErrorResponse{
		Code:    pe.Code,
		Message: pe.Message,
		Details: pe.Details,
		Meta:    pe.Metadata,
	})
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// parseBoolQuery parses a boolean query parameter with a default value
func parseBoolQuery(c *gin.Context, key string, defaultValue bool) bool {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Services    map[string]string `json:"services,omitempty"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	version string
	env     string
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(version, env string) *HealthHandler {
	return &HealthHandler{
		version: version,
		env:     env,
	}
}

// Live handles GET /health/live
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:      "ok",
		Version:     h.version,
		Environment: h.env,
	})
}

// Ready handles GET /health/ready
func (h *HealthHandler) Ready(c *gin.Context) {
	// In a full implementation, this would check database connectivity,
	// message queue connectivity, etc.
	c.JSON(http.StatusOK, HealthResponse{
		Status:      "ok",
		Version:     h.version,
		Environment: h.env,
		Services: map[string]string{
			"database": "ok",
			"nats":     "ok",
			"coolify":  "ok",
			"rancher":  "ok",
			"argocd":   "ok",
		},
	})
}
