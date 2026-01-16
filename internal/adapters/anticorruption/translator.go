// Package anticorruption provides the Anti-Corruption Layer (ACL) for Legacy Modernization.
// The ACL translates between the new domain model and legacy external systems,
// preventing legacy concepts from leaking into the core domain.
package anticorruption

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/northstack/platform/internal/domain"
)

// LegacyProjectDTO represents a project in legacy format (e.g., from Northflank-style API)
type LegacyProjectDTO struct {
	ProjectID   string                 `json:"projectId"`
	ProjectName string                 `json:"name"`
	Desc        string                 `json:"description"`
	TeamSlug    string                 `json:"teamSlug"`
	CreatedTime int64                  `json:"createdAt"`
	UpdatedTime int64                  `json:"updatedAt"`
	Tags        []string               `json:"tags,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// LegacyServiceDTO represents a service in legacy format
type LegacyServiceDTO struct {
	ServiceID     string            `json:"serviceId"`
	Name          string            `json:"name"`
	Type          string            `json:"kind"` // Different field name in legacy
	ProjectRef    string            `json:"projectId"`
	RepoURL       string            `json:"gitUrl"`
	BranchName    string            `json:"branch"`
	ContainerPort int               `json:"port"`
	EnvVars       map[string]string `json:"env"`
	Instances     int               `json:"replicas"`
	CPUShares     int               `json:"cpu"`
	MemoryMB      int               `json:"memory"`
	StatusStr     string            `json:"status"`
}

// ProjectTranslator translates between legacy and domain project representations
type ProjectTranslator struct{}

// NewProjectTranslator creates a new translator
func NewProjectTranslator() *ProjectTranslator {
	return &ProjectTranslator{}
}

// FromLegacy converts a legacy project DTO to domain Project
func (t *ProjectTranslator) FromLegacy(legacy *LegacyProjectDTO) (*domain.Project, error) {
	id, err := uuid.Parse(legacy.ProjectID)
	if err != nil {
		// Generate new ID if legacy ID is not UUID format
		id = uuid.New()
	}

	// Convert tags to labels
	labels := make(map[string]string)
	for i, tag := range legacy.Tags {
		labels[fmt.Sprintf("tag.%d", i)] = tag
	}

	return &domain.Project{
		ID:          id,
		Name:        legacy.ProjectName,
		Slug:        generateSlug(legacy.ProjectName),
		Description: legacy.Desc,
		Status:      domain.ProjectStatusActive,
		Labels:      labels,
		Metadata:    legacy.Settings,
		CreatedAt:   time.Unix(legacy.CreatedTime/1000, 0),
		UpdatedAt:   time.Unix(legacy.UpdatedTime/1000, 0),
	}, nil
}

// ToLegacy converts a domain Project to legacy DTO format
func (t *ProjectTranslator) ToLegacy(project *domain.Project) *LegacyProjectDTO {
	// Convert labels back to tags
	var tags []string
	for key, value := range project.Labels {
		if strings.HasPrefix(key, "tag.") {
			tags = append(tags, value)
		}
	}

	return &LegacyProjectDTO{
		ProjectID:   project.ID.String(),
		ProjectName: project.Name,
		Desc:        project.Description,
		CreatedTime: project.CreatedAt.UnixMilli(),
		UpdatedTime: project.UpdatedAt.UnixMilli(),
		Tags:        tags,
		Settings:    project.Metadata,
	}
}

// ServiceTranslator translates between legacy and domain service representations
type ServiceTranslator struct{}

// NewServiceTranslator creates a new translator
func NewServiceTranslator() *ServiceTranslator {
	return &ServiceTranslator{}
}

// FromLegacy converts a legacy service DTO to domain Service
func (t *ServiceTranslator) FromLegacy(legacy *LegacyServiceDTO) (*domain.Service, error) {
	id, _ := uuid.Parse(legacy.ServiceID)
	if id == uuid.Nil {
		id = uuid.New()
	}

	projectID, _ := uuid.Parse(legacy.ProjectRef)

	// Map legacy type to domain ServiceType
	serviceType := mapLegacyServiceType(legacy.Type)

	// Map legacy status to domain ServiceStatus
	status := mapLegacyServiceStatus(legacy.StatusStr)

	return &domain.Service{
		ID:        id,
		ProjectID: projectID,
		Name:      legacy.Name,
		Slug:      generateSlug(legacy.Name),
		Type:      serviceType,
		Status:    status,
		BuildSource: domain.BuildSource{
			Type:       "git",
			Repository: legacy.RepoURL,
			Branch:     legacy.BranchName,
		},
		Resources: domain.ResourceLimits{
			CPURequest:    fmt.Sprintf("%dm", legacy.CPUShares),
			CPULimit:      fmt.Sprintf("%dm", legacy.CPUShares*2),
			MemoryRequest: fmt.Sprintf("%dMi", legacy.MemoryMB),
			MemoryLimit:   fmt.Sprintf("%dMi", legacy.MemoryMB*2),
		},
		Scaling: domain.ScalingConfig{
			MinReplicas: int32(legacy.Instances),
			MaxReplicas: int32(legacy.Instances * 2),
		},
		Ports: []domain.ServicePort{
			{
				Name:       "http",
				Port:       int32(legacy.ContainerPort),
				TargetPort: int32(legacy.ContainerPort),
				Protocol:   "TCP",
			},
		},
		EnvVars:   legacy.EnvVars,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// ToLegacy converts a domain Service to legacy DTO format
func (t *ServiceTranslator) ToLegacy(service *domain.Service) *LegacyServiceDTO {
	port := 8080
	if len(service.Ports) > 0 {
		port = int(service.Ports[0].Port)
	}

	return &LegacyServiceDTO{
		ServiceID:     service.ID.String(),
		Name:          service.Name,
		Type:          mapDomainServiceType(service.Type),
		ProjectRef:    service.ProjectID.String(),
		RepoURL:       service.BuildSource.Repository,
		BranchName:    service.BuildSource.Branch,
		ContainerPort: port,
		EnvVars:       service.EnvVars,
		Instances:     int(service.Scaling.MinReplicas),
		StatusStr:     mapDomainServiceStatus(service.Status),
	}
}

// LegacySystemFacade provides a unified interface to legacy systems
type LegacySystemFacade struct {
	projectTranslator *ProjectTranslator
	serviceTranslator *ServiceTranslator
	legacyAPIEndpoint string
}

// NewLegacySystemFacade creates a new facade
func NewLegacySystemFacade(legacyEndpoint string) *LegacySystemFacade {
	return &LegacySystemFacade{
		projectTranslator: NewProjectTranslator(),
		serviceTranslator: NewServiceTranslator(),
		legacyAPIEndpoint: legacyEndpoint,
	}
}

// ImportProject imports a project from the legacy system
func (f *LegacySystemFacade) ImportProject(ctx context.Context, legacyJSON []byte) (*domain.Project, error) {
	var legacyProject LegacyProjectDTO
	if err := json.Unmarshal(legacyJSON, &legacyProject); err != nil {
		return nil, fmt.Errorf("failed to parse legacy project: %w", err)
	}

	return f.projectTranslator.FromLegacy(&legacyProject)
}

// ExportProject exports a project to legacy format
func (f *LegacySystemFacade) ExportProject(project *domain.Project) ([]byte, error) {
	legacyDTO := f.projectTranslator.ToLegacy(project)
	return json.Marshal(legacyDTO)
}

// ImportService imports a service from the legacy system
func (f *LegacySystemFacade) ImportService(ctx context.Context, legacyJSON []byte) (*domain.Service, error) {
	var legacyService LegacyServiceDTO
	if err := json.Unmarshal(legacyJSON, &legacyService); err != nil {
		return nil, fmt.Errorf("failed to parse legacy service: %w", err)
	}

	return f.serviceTranslator.FromLegacy(&legacyService)
}

// ExportService exports a service to legacy format
func (f *LegacySystemFacade) ExportService(service *domain.Service) ([]byte, error) {
	legacyDTO := f.serviceTranslator.ToLegacy(service)
	return json.Marshal(legacyDTO)
}

// Helper functions

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}

func mapLegacyServiceType(legacyType string) domain.ServiceType {
	switch strings.ToLower(legacyType) {
	case "combined", "web", "webapp":
		return domain.ServiceTypeWebApp
	case "worker", "process":
		return domain.ServiceTypeWorker
	case "cron", "scheduled":
		return domain.ServiceTypeCronJob
	case "database", "db":
		return domain.ServiceTypeStatefulDB
	default:
		return domain.ServiceTypeStateless
	}
}

func mapLegacyServiceStatus(legacyStatus string) domain.ServiceStatus {
	switch strings.ToLower(legacyStatus) {
	case "running", "active", "healthy":
		return domain.ServiceStatusRunning
	case "stopped", "paused":
		return domain.ServiceStatusStopped
	case "building", "queued":
		return domain.ServiceStatusBuilding
	case "deploying", "rolling":
		return domain.ServiceStatusDeploying
	case "failed", "error", "unhealthy":
		return domain.ServiceStatusFailed
	default:
		return domain.ServiceStatusPending
	}
}

func mapDomainServiceType(serviceType domain.ServiceType) string {
	switch serviceType {
	case domain.ServiceTypeWebApp:
		return "combined"
	case domain.ServiceTypeWorker:
		return "worker"
	case domain.ServiceTypeCronJob:
		return "cron"
	case domain.ServiceTypeStatefulDB:
		return "database"
	default:
		return "container"
	}
}

func mapDomainServiceStatus(status domain.ServiceStatus) string {
	switch status {
	case domain.ServiceStatusRunning:
		return "running"
	case domain.ServiceStatusStopped:
		return "stopped"
	case domain.ServiceStatusBuilding:
		return "building"
	case domain.ServiceStatusDeploying:
		return "deploying"
	case domain.ServiceStatusFailed:
		return "failed"
	default:
		return "pending"
	}
}
