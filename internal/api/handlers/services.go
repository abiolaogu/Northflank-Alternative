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

// ServiceHandler handles service-related HTTP requests
type ServiceHandler struct {
	serviceRepo domain.ServiceRepository
	projectRepo domain.ProjectRepository
	ciAdapter   domain.CIAdapter
	eventBus    domain.EventBus
	logger      *logger.Logger
}

// NewServiceHandler creates a new ServiceHandler
func NewServiceHandler(
	serviceRepo domain.ServiceRepository,
	projectRepo domain.ProjectRepository,
	ciAdapter domain.CIAdapter,
	eventBus domain.EventBus,
	log *logger.Logger,
) *ServiceHandler {
	return &ServiceHandler{
		serviceRepo: serviceRepo,
		projectRepo: projectRepo,
		ciAdapter:   ciAdapter,
		eventBus:    eventBus,
		logger:      log,
	}
}

// CreateServiceRequest represents the request body for creating a service
type CreateServiceRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	Slug        string                 `json:"slug" binding:"required,min=1,max=255"`
	Type        string                 `json:"type" binding:"required,oneof=webapp worker cronjob stateful_db stateless"`
	BuildSource BuildSourceRequest     `json:"build_source" binding:"required"`
	Resources   *ResourceLimitsRequest `json:"resources,omitempty"`
	Scaling     *ScalingConfigRequest  `json:"scaling,omitempty"`
	HealthCheck *HealthCheckRequest    `json:"health_check,omitempty"`
	EnvVars     map[string]string      `json:"env_vars,omitempty"`
	SecretRefs  []string               `json:"secret_refs,omitempty"`
	Ports       []PortRequest          `json:"ports,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
}

// BuildSourceRequest represents build source configuration
type BuildSourceRequest struct {
	Type       string `json:"type" binding:"required,oneof=git docker buildpack"`
	Repository string `json:"repository,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Dockerfile string `json:"dockerfile,omitempty"`
	Image      string `json:"image,omitempty"`
	Registry   string `json:"registry,omitempty"`
}

// ResourceLimitsRequest represents resource limits configuration
type ResourceLimitsRequest struct {
	CPURequest    string `json:"cpu_request,omitempty"`
	CPULimit      string `json:"cpu_limit,omitempty"`
	MemoryRequest string `json:"memory_request,omitempty"`
	MemoryLimit   string `json:"memory_limit,omitempty"`
	StorageSize   string `json:"storage_size,omitempty"`
}

// ScalingConfigRequest represents scaling configuration
type ScalingConfigRequest struct {
	MinReplicas  int32 `json:"min_replicas"`
	MaxReplicas  int32 `json:"max_replicas"`
	TargetCPU    int32 `json:"target_cpu,omitempty"`
	TargetMemory int32 `json:"target_memory,omitempty"`
}

// HealthCheckRequest represents health check configuration
type HealthCheckRequest struct {
	Type                string `json:"type" binding:"required,oneof=http tcp exec"`
	Path                string `json:"path,omitempty"`
	Port                int32  `json:"port,omitempty"`
	Command             string `json:"command,omitempty"`
	InitialDelaySeconds int32  `json:"initial_delay_seconds"`
	PeriodSeconds       int32  `json:"period_seconds"`
	TimeoutSeconds      int32  `json:"timeout_seconds"`
	FailureThreshold    int32  `json:"failure_threshold"`
}

// PortRequest represents a port configuration
type PortRequest struct {
	Name       string `json:"name" binding:"required"`
	Port       int32  `json:"port" binding:"required"`
	TargetPort int32  `json:"target_port,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Public     bool   `json:"public"`
}

// ServiceResponse represents the response body for a service
type ServiceResponse struct {
	ID             uuid.UUID             `json:"id"`
	ProjectID      uuid.UUID             `json:"project_id"`
	Name           string                `json:"name"`
	Slug           string                `json:"slug"`
	Type           string                `json:"type"`
	Status         string                `json:"status"`
	BuildSource    domain.BuildSource    `json:"build_source"`
	Resources      domain.ResourceLimits `json:"resources"`
	Scaling        domain.ScalingConfig  `json:"scaling"`
	HealthCheck    *domain.HealthCheck   `json:"health_check,omitempty"`
	EnvVars        map[string]string     `json:"env_vars,omitempty"`
	SecretRefs     []string              `json:"secret_refs,omitempty"`
	Ports          []domain.ServicePort  `json:"ports,omitempty"`
	Labels         map[string]string     `json:"labels,omitempty"`
	CurrentVersion string                `json:"current_version,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// Create handles POST /projects/:project_id/services
func (h *ServiceHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		respondError(c, errors.BadRequest("invalid project ID"))
		return
	}

	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errors.BadRequest(err.Error()))
		return
	}

	// Verify project exists
	_, err = h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		respondError(c, err)
		return
	}

	service := &domain.Service{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      req.Name,
		Slug:      req.Slug,
		Type:      domain.ServiceType(req.Type),
		Status:    domain.ServiceStatusPending,
		BuildSource: domain.BuildSource{
			Type:       req.BuildSource.Type,
			Repository: req.BuildSource.Repository,
			Branch:     req.BuildSource.Branch,
			Dockerfile: req.BuildSource.Dockerfile,
			Image:      req.BuildSource.Image,
			Registry:   req.BuildSource.Registry,
		},
		EnvVars:    req.EnvVars,
		SecretRefs: req.SecretRefs,
		Labels:     req.Labels,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Set defaults for scaling
	if req.Scaling != nil {
		service.Scaling = domain.ScalingConfig{
			MinReplicas:  req.Scaling.MinReplicas,
			MaxReplicas:  req.Scaling.MaxReplicas,
			TargetCPU:    req.Scaling.TargetCPU,
			TargetMemory: req.Scaling.TargetMemory,
		}
	} else {
		service.Scaling = domain.ScalingConfig{
			MinReplicas: 1,
			MaxReplicas: 3,
			TargetCPU:   80,
		}
	}

	// Set resources with defaults
	if req.Resources != nil {
		service.Resources = domain.ResourceLimits{
			CPURequest:    req.Resources.CPURequest,
			CPULimit:      req.Resources.CPULimit,
			MemoryRequest: req.Resources.MemoryRequest,
			MemoryLimit:   req.Resources.MemoryLimit,
			StorageSize:   req.Resources.StorageSize,
		}
	} else {
		service.Resources = domain.ResourceLimits{
			CPURequest:    "100m",
			CPULimit:      "500m",
			MemoryRequest: "128Mi",
			MemoryLimit:   "512Mi",
		}
	}

	// Set health check
	if req.HealthCheck != nil {
		service.HealthCheck = &domain.HealthCheck{
			Type:                req.HealthCheck.Type,
			Path:                req.HealthCheck.Path,
			Port:                req.HealthCheck.Port,
			Command:             req.HealthCheck.Command,
			InitialDelaySeconds: req.HealthCheck.InitialDelaySeconds,
			PeriodSeconds:       req.HealthCheck.PeriodSeconds,
			TimeoutSeconds:      req.HealthCheck.TimeoutSeconds,
			FailureThreshold:    req.HealthCheck.FailureThreshold,
			SuccessThreshold:    1,
		}
	}

	// Set ports
	if len(req.Ports) > 0 {
		service.Ports = make([]domain.ServicePort, len(req.Ports))
		for i, p := range req.Ports {
			targetPort := p.TargetPort
			if targetPort == 0 {
				targetPort = p.Port
			}
			protocol := p.Protocol
			if protocol == "" {
				protocol = "TCP"
			}
			service.Ports[i] = domain.ServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: targetPort,
				Protocol:   protocol,
				Public:     p.Public,
			}
		}
	}

	if err := h.serviceRepo.Create(c.Request.Context(), service); err != nil {
		respondError(c, err)
		return
	}

	// Publish event
	h.eventBus.Publish(c.Request.Context(), "service.created", &domain.Event{
		Type:   "service.created",
		Source: "api",
		Data: map[string]interface{}{
			"service_id": service.ID.String(),
			"project_id": projectID.String(),
			"name":       service.Name,
			"type":       string(service.Type),
		},
	})

	h.logger.Info().
		Str("service_id", service.ID.String()).
		Str("project_id", projectID.String()).
		Str("slug", service.Slug).
		Msg("Service created")

	c.JSON(http.StatusCreated, serviceToResponse(service))
}

// Get handles GET /services/:id
func (h *ServiceHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid service ID"))
		return
	}

	service, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, serviceToResponse(service))
}

// ListByProject handles GET /projects/:project_id/services
func (h *ServiceHandler) ListByProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		respondError(c, errors.BadRequest("invalid project ID"))
		return
	}

	var filter domain.ServiceFilter

	if serviceType := c.Query("type"); serviceType != "" {
		t := domain.ServiceType(serviceType)
		filter.Type = &t
	}

	if status := c.Query("status"); status != "" {
		s := domain.ServiceStatus(status)
		filter.Status = &s
	}

	filter.Search = c.Query("search")
	filter.Limit = parseIntQuery(c, "limit", 50)
	filter.Offset = parseIntQuery(c, "offset", 0)

	services, err := h.serviceRepo.ListByProject(c.Request.Context(), projectID, filter)
	if err != nil {
		respondError(c, err)
		return
	}

	responses := make([]ServiceResponse, len(services))
	for i, s := range services {
		responses[i] = serviceToResponse(s)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   responses,
		"count":  len(responses),
		"offset": filter.Offset,
		"limit":  filter.Limit,
	})
}

// Update handles PATCH /services/:id
func (h *ServiceHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid service ID"))
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errors.BadRequest(err.Error()))
		return
	}

	service, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	// Apply updates (simplified - in production would use proper struct mapping)
	if name, ok := req["name"].(string); ok {
		service.Name = name
	}
	if envVars, ok := req["env_vars"].(map[string]interface{}); ok {
		service.EnvVars = make(map[string]string)
		for k, v := range envVars {
			if str, ok := v.(string); ok {
				service.EnvVars[k] = str
			}
		}
	}
	if labels, ok := req["labels"].(map[string]interface{}); ok {
		service.Labels = make(map[string]string)
		for k, v := range labels {
			if str, ok := v.(string); ok {
				service.Labels[k] = str
			}
		}
	}

	if err := h.serviceRepo.Update(c.Request.Context(), service); err != nil {
		respondError(c, err)
		return
	}

	// Publish event
	h.eventBus.Publish(c.Request.Context(), "service.updated", &domain.Event{
		Type:   "service.updated",
		Source: "api",
		Data: map[string]interface{}{
			"service_id": service.ID.String(),
			"project_id": service.ProjectID.String(),
		},
	})

	h.logger.Info().
		Str("service_id", service.ID.String()).
		Msg("Service updated")

	c.JSON(http.StatusOK, serviceToResponse(service))
}

// Delete handles DELETE /services/:id
func (h *ServiceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid service ID"))
		return
	}

	service, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	if err := h.serviceRepo.Delete(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}

	// Publish event
	h.eventBus.Publish(c.Request.Context(), "service.deleted", &domain.Event{
		Type:   "service.deleted",
		Source: "api",
		Data: map[string]interface{}{
			"service_id": service.ID.String(),
			"project_id": service.ProjectID.String(),
			"name":       service.Name,
		},
	})

	h.logger.Info().
		Str("service_id", id.String()).
		Msg("Service deleted")

	c.Status(http.StatusNoContent)
}

// TriggerBuild handles POST /services/:id/builds
func (h *ServiceHandler) TriggerBuild(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid service ID"))
		return
	}

	service, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	var req struct {
		Branch    string `json:"branch,omitempty"`
		CommitSHA string `json:"commit_sha,omitempty"`
	}
	c.ShouldBindJSON(&req)

	source := service.BuildSource
	if req.Branch != "" {
		source.Branch = req.Branch
	}
	if req.CommitSHA != "" {
		source.CommitSHA = req.CommitSHA
	}

	build, err := h.ciAdapter.TriggerBuild(c.Request.Context(), service, source)
	if err != nil {
		respondError(c, err)
		return
	}

	// Update service status
	h.serviceRepo.UpdateStatus(c.Request.Context(), id, domain.ServiceStatusBuilding)

	h.logger.Info().
		Str("service_id", id.String()).
		Str("build_id", build.ID.String()).
		Msg("Build triggered")

	c.JSON(http.StatusAccepted, gin.H{
		"build_id": build.ID,
		"status":   string(build.Status),
		"message":  "Build started",
	})
}

// Scale handles POST /services/:id/scale
func (h *ServiceHandler) Scale(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, errors.BadRequest("invalid service ID"))
		return
	}

	var req struct {
		Replicas int32 `json:"replicas" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errors.BadRequest(err.Error()))
		return
	}

	service, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	// Update scaling config
	service.Scaling.MinReplicas = req.Replicas
	service.Scaling.MaxReplicas = req.Replicas

	if err := h.serviceRepo.Update(c.Request.Context(), service); err != nil {
		respondError(c, err)
		return
	}

	// Publish scale event
	h.eventBus.Publish(c.Request.Context(), "service.scaled", &domain.Event{
		Type:   "service.scaled",
		Source: "api",
		Data: map[string]interface{}{
			"service_id": service.ID.String(),
			"replicas":   req.Replicas,
		},
	})

	h.logger.Info().Str("service_id", id.String()).Int("replicas", int(req.Replicas)).Msg("Scaling service")

	c.JSON(http.StatusOK, gin.H{
		"message":  "Service scaled",
		"replicas": req.Replicas,
	})
}

func serviceToResponse(s *domain.Service) ServiceResponse {
	return ServiceResponse{
		ID:             s.ID,
		ProjectID:      s.ProjectID,
		Name:           s.Name,
		Slug:           s.Slug,
		Type:           string(s.Type),
		Status:         string(s.Status),
		BuildSource:    s.BuildSource,
		Resources:      s.Resources,
		Scaling:        s.Scaling,
		HealthCheck:    s.HealthCheck,
		EnvVars:        s.EnvVars,
		SecretRefs:     s.SecretRefs,
		Ports:          s.Ports,
		Labels:         s.Labels,
		CurrentVersion: s.CurrentVersion,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}
