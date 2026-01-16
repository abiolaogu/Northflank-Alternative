// Package handlers contains HTTP handlers for the REST API.
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/errors"
	"github.com/northstack/platform/pkg/logger"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	repo     domain.ProjectRepository
	eventBus domain.EventBus
	logger   *logger.Logger
}

// NewProjectHandler creates a new ProjectHandler
func NewProjectHandler(repo domain.ProjectRepository, eventBus domain.EventBus, log *logger.Logger) *ProjectHandler {
	return &ProjectHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   log,
	}
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name        string            `json:"name" binding:"required,min=1,max=255"`
	Slug        string            `json:"slug" binding:"required,min=1,max=255,alphanum"`
	Description string            `json:"description,omitempty"`
	TeamID      *uuid.UUID        `json:"team_id,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	TeamID      *uuid.UUID        `json:"team_id,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// ProjectResponse represents the response body for a project
type ProjectResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status"`
	OwnerID     uuid.UUID         `json:"owner_id"`
	TeamID      *uuid.UUID        `json:"team_id,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Create handles POST /projects
func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errors.BadRequest(err.Error()))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, errors.Unauthorized("user not authenticated"))
		return
	}

	project := &domain.Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Status:      domain.ProjectStatusActive,
		OwnerID:     userID.(uuid.UUID),
		TeamID:      req.TeamID,
		Labels:      req.Labels,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(c.Request.Context(), project); err != nil {
		h.logger.Error().Err(err).Str("slug", req.Slug).Msg("Failed to create project")
		respondError(c, err)
		return
	}

	// Publish event
	h.eventBus.Publish(c.Request.Context(), "project.created", &domain.Event{
		Type:   "project.created",
		Source: "api",
		Data: map[string]interface{}{
			"project_id": project.ID.String(),
			"name":       project.Name,
			"owner_id":   project.OwnerID.String(),
		},
	})

	h.logger.Info().
		Str("project_id", project.ID.String()).
		Str("slug", project.Slug).
		Msg("Project created")

	c.JSON(http.StatusCreated, projectToResponse(project))
}

// Get handles GET /projects/:id
func (h *ProjectHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid project ID"))
		return
	}

	project, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, projectToResponse(project))
}

// GetBySlug handles GET /projects/slug/:slug
func (h *ProjectHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	project, err := h.repo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, projectToResponse(project))
}

// List handles GET /projects
func (h *ProjectHandler) List(c *gin.Context) {
	var filter domain.ProjectFilter

	// Parse query parameters
	if ownerIDStr := c.Query("owner_id"); ownerIDStr != "" {
		if ownerID, err := uuid.Parse(ownerIDStr); err == nil {
			filter.OwnerID = &ownerID
		}
	}

	if teamIDStr := c.Query("team_id"); teamIDStr != "" {
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			filter.TeamID = &teamID
		}
	}

	if status := c.Query("status"); status != "" {
		s := domain.ProjectStatus(status)
		filter.Status = &s
	}

	filter.Search = c.Query("search")
	filter.Limit = parseIntQuery(c, "limit", 50)
	filter.Offset = parseIntQuery(c, "offset", 0)

	projects, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}

	responses := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = projectToResponse(p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   responses,
		"count":  len(responses),
		"offset": filter.Offset,
		"limit":  filter.Limit,
	})
}

// Update handles PATCH /projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid project ID"))
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errors.BadRequest(err.Error()))
		return
	}

	project, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	// Apply updates
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.TeamID != nil {
		project.TeamID = req.TeamID
	}
	if req.Labels != nil {
		project.Labels = req.Labels
	}

	if err := h.repo.Update(c.Request.Context(), project); err != nil {
		respondError(c, err)
		return
	}

	h.logger.Info().
		Str("project_id", project.ID.String()).
		Msg("Project updated")

	c.JSON(http.StatusOK, projectToResponse(project))
}

// Delete handles DELETE /projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid project ID"))
		return
	}

	// Get project for event
	project, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}

	// Publish event
	h.eventBus.Publish(c.Request.Context(), "project.deleted", &domain.Event{
		Type:   "project.deleted",
		Source: "api",
		Data: map[string]interface{}{
			"project_id": project.ID.String(),
			"name":       project.Name,
		},
	})

	h.logger.Info().
		Str("project_id", id.String()).
		Msg("Project deleted")

	c.Status(http.StatusNoContent)
}

func projectToResponse(p *domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Status:      string(p.Status),
		OwnerID:     p.OwnerID,
		TeamID:      p.TeamID,
		Labels:      p.Labels,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
