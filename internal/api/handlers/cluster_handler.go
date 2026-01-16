package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/logger"
)

// ClusterHandler handles Kubernetes cluster management endpoints
type ClusterHandler struct {
	clusterRepo domain.ClusterRepository
	eventBus    domain.EventBus
	logger      *logger.Logger
}

// NewClusterHandler creates a new ClusterHandler
func NewClusterHandler(clusterRepo domain.ClusterRepository, eventBus domain.EventBus, log *logger.Logger) *ClusterHandler {
	return &ClusterHandler{
		clusterRepo: clusterRepo,
		eventBus:    eventBus,
		logger:      log,
	}
}

// CreateClusterRequest represents a cluster creation request
type CreateClusterRequest struct {
	Name        string            `json:"name" binding:"required"`
	Provider    string            `json:"provider" binding:"required,oneof=rancher rke2 k3s eks gke aks"`
	Region      string            `json:"region" binding:"required"`
	KubeVersion string            `json:"kube_version"`
	NodeCount   int32             `json:"node_count" binding:"required,min=1"`
	Labels      map[string]string `json:"labels"`
}

// ClusterResponse represents a cluster in API responses
type ClusterResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Provider    string            `json:"provider"`
	Region      string            `json:"region"`
	KubeVersion string            `json:"kube_version"`
	Status      string            `json:"status"`
	Endpoint    string            `json:"endpoint,omitempty"`
	NodeCount   int32             `json:"node_count"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateCluster creates a new Kubernetes cluster
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster := &domain.Cluster{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Name, // Simplify for now
		Provider:    domain.ClusterProvider(req.Provider),
		Region:      req.Region,
		KubeVersion: req.KubeVersion,
		Status:      domain.ClusterStatusProvisioning,
		NodeCount:   req.NodeCount,
		Labels:      req.Labels,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.clusterRepo.Create(c.Request.Context(), cluster); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create cluster")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cluster"})
		return
	}

	// Publish event
	h.publishEvent(c.Request.Context(), "cluster.created", map[string]interface{}{
		"cluster_id": cluster.ID.String(),
		"name":       cluster.Name,
		"provider":   string(cluster.Provider),
	})

	c.JSON(http.StatusCreated, h.toResponse(cluster))
}

// ListClusters lists all clusters
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	filter := domain.ClusterFilter{
		Limit: 100,
	}

	clusters, err := h.clusterRepo.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list clusters")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list clusters"})
		return
	}

	responses := make([]ClusterResponse, len(clusters))
	for i, cluster := range clusters {
		responses[i] = h.toResponse(cluster)
	}

	c.JSON(http.StatusOK, gin.H{
		"clusters": responses,
		"total":    len(responses),
	})
}

// GetCluster retrieves a specific cluster
func (h *ClusterHandler) GetCluster(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	cluster, err := h.clusterRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
		return
	}

	c.JSON(http.StatusOK, h.toResponse(cluster))
}

// DeleteCluster deletes a cluster
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	cluster, err := h.clusterRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
		return
	}

	// Mark as deleting
	cluster.Status = domain.ClusterStatusDeleting
	if err := h.clusterRepo.Update(c.Request.Context(), cluster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cluster"})
		return
	}

	// Publish event
	h.publishEvent(c.Request.Context(), "cluster.deleted", map[string]interface{}{
		"cluster_id": cluster.ID.String(),
		"name":       cluster.Name,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Cluster deletion initiated"})
}

// GetClusterKubeconfig retrieves the kubeconfig for a cluster
func (h *ClusterHandler) GetClusterKubeconfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	cluster, err := h.clusterRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
		return
	}

	if cluster.Status != domain.ClusterStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster is not ready"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"kubeconfig": "# Kubeconfig retrieved from cluster provider",
		"cluster_id": cluster.ID,
		"endpoint":   cluster.APIEndpoint,
	})
}

func (h *ClusterHandler) toResponse(cluster *domain.Cluster) ClusterResponse {
	return ClusterResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Slug:        cluster.Slug,
		Provider:    string(cluster.Provider),
		Region:      cluster.Region,
		KubeVersion: cluster.KubeVersion,
		Status:      string(cluster.Status),
		Endpoint:    cluster.APIEndpoint,
		NodeCount:   cluster.NodeCount,
		Labels:      cluster.Labels,
		CreatedAt:   cluster.CreatedAt,
		UpdatedAt:   cluster.UpdatedAt,
	}
}

func (h *ClusterHandler) publishEvent(ctx context.Context, eventType string, data map[string]interface{}) {
	event := &domain.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	if err := h.eventBus.Publish(ctx, eventType, event); err != nil {
		h.logger.Warn().Err(err).Str("event", eventType).Msg("Failed to publish event")
	}
}
