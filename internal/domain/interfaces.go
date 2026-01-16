// Package domain contains the core domain interfaces for the Platform Orchestrator.
// These interfaces define the contracts that adapters and repositories must implement.
package domain

import (
	"context"

	"github.com/google/uuid"
)

// ProjectRepository defines the interface for project persistence
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	GetBySlug(ctx context.Context, slug string) (*Project, error)
	List(ctx context.Context, filter ProjectFilter) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectFilter defines filtering options for listing projects
type ProjectFilter struct {
	OwnerID *uuid.UUID
	TeamID  *uuid.UUID
	Status  *ProjectStatus
	Labels  map[string]string
	Search  string
	Limit   int
	Offset  int
}

// ServiceRepository defines the interface for service persistence
type ServiceRepository interface {
	Create(ctx context.Context, service *Service) error
	GetByID(ctx context.Context, id uuid.UUID) (*Service, error)
	GetBySlug(ctx context.Context, projectID uuid.UUID, slug string) (*Service, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, filter ServiceFilter) ([]*Service, error)
	Update(ctx context.Context, service *Service) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status ServiceStatus) error
}

// ServiceFilter defines filtering options for listing services
type ServiceFilter struct {
	Type   *ServiceType
	Status *ServiceStatus
	Labels map[string]string
	Search string
	Limit  int
	Offset int
}

// BuildRepository defines the interface for build persistence
type BuildRepository interface {
	Create(ctx context.Context, build *Build) error
	GetByID(ctx context.Context, id uuid.UUID) (*Build, error)
	ListByService(ctx context.Context, serviceID uuid.UUID, limit int) ([]*Build, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, limit int) ([]*Build, error)
	Update(ctx context.Context, build *Build) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status BuildStatus, errorMsg string) error
}

// DeploymentRepository defines the interface for deployment persistence
type DeploymentRepository interface {
	Create(ctx context.Context, deployment *Deployment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Deployment, error)
	GetLatestByService(ctx context.Context, serviceID uuid.UUID) (*Deployment, error)
	ListByService(ctx context.Context, serviceID uuid.UUID, limit int) ([]*Deployment, error)
	ListByCluster(ctx context.Context, clusterID uuid.UUID, limit int) ([]*Deployment, error)
	Update(ctx context.Context, deployment *Deployment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status DeploymentStatus, errorMsg string) error
}

// ClusterRepository defines the interface for cluster persistence
type ClusterRepository interface {
	Create(ctx context.Context, cluster *Cluster) error
	GetByID(ctx context.Context, id uuid.UUID) (*Cluster, error)
	GetBySlug(ctx context.Context, slug string) (*Cluster, error)
	List(ctx context.Context, filter ClusterFilter) ([]*Cluster, error)
	Update(ctx context.Context, cluster *Cluster) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ClusterFilter defines filtering options for listing clusters
type ClusterFilter struct {
	Provider *ClusterProvider
	Status   *ClusterStatus
	Region   string
	Labels   map[string]string
	Limit    int
	Offset   int
}

// EnvironmentRepository defines the interface for environment persistence
type EnvironmentRepository interface {
	Create(ctx context.Context, environment *Environment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Environment, error)
	GetBySlug(ctx context.Context, projectID uuid.UUID, slug string) (*Environment, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*Environment, error)
	ListByCluster(ctx context.Context, clusterID uuid.UUID) ([]*Environment, error)
	Update(ctx context.Context, environment *Environment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SecretRepository defines the interface for secret metadata persistence
type SecretRepository interface {
	Create(ctx context.Context, secret *Secret) error
	GetByID(ctx context.Context, id uuid.UUID) (*Secret, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*Secret, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*Secret, error)
	Update(ctx context.Context, secret *Secret) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// IngressRepository defines the interface for ingress persistence
type IngressRepository interface {
	Create(ctx context.Context, ingress *Ingress) error
	GetByID(ctx context.Context, id uuid.UUID) (*Ingress, error)
	GetByDomain(ctx context.Context, domain string) (*Ingress, error)
	ListByService(ctx context.Context, serviceID uuid.UUID) ([]*Ingress, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*Ingress, error)
	Update(ctx context.Context, ingress *Ingress) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PipelineRepository defines the interface for pipeline persistence
type PipelineRepository interface {
	Create(ctx context.Context, pipeline *Pipeline) error
	GetByID(ctx context.Context, id uuid.UUID) (*Pipeline, error)
	ListByService(ctx context.Context, serviceID uuid.UUID, limit int) ([]*Pipeline, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, limit int) ([]*Pipeline, error)
	Update(ctx context.Context, pipeline *Pipeline) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status PipelineStatus) error
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TeamRepository defines the interface for team persistence
type TeamRepository interface {
	Create(ctx context.Context, team *Team) error
	GetByID(ctx context.Context, id uuid.UUID) (*Team, error)
	GetBySlug(ctx context.Context, slug string) (*Team, error)
	List(ctx context.Context, limit, offset int) ([]*Team, error)
	Update(ctx context.Context, team *Team) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, membership *TeamMembership) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	GetMembers(ctx context.Context, teamID uuid.UUID) ([]*TeamMembership, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)
}

// AuditLogRepository defines the interface for audit log persistence
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, error)
}

// AuditLogFilter defines filtering options for listing audit logs
type AuditLogFilter struct {
	UserID       *uuid.UUID
	ProjectID    *uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	Action       *AuditAction
	StartTime    *int64
	EndTime      *int64
	Limit        int
	Offset       int
}

// CIAdapter defines the interface for CI/Build systems (e.g., Coolify)
type CIAdapter interface {
	// TriggerBuild triggers a new build for a service
	TriggerBuild(ctx context.Context, service *Service, source BuildSource) (*Build, error)
	// GetBuildStatus gets the current status of a build
	GetBuildStatus(ctx context.Context, buildID string) (*Build, error)
	// CancelBuild cancels a running build
	CancelBuild(ctx context.Context, buildID string) error
	// GetBuildLogs retrieves logs for a build
	GetBuildLogs(ctx context.Context, buildID string) (string, error)
	// CreateProject creates a project in the CI system
	CreateProject(ctx context.Context, project *Project) (string, error)
	// DeleteProject deletes a project from the CI system
	DeleteProject(ctx context.Context, externalID string) error
}

// ClusterManagerAdapter defines the interface for Kubernetes cluster management (e.g., Rancher)
type ClusterManagerAdapter interface {
	// CreateCluster provisions a new Kubernetes cluster
	CreateCluster(ctx context.Context, cluster *Cluster) (string, error)
	// GetCluster retrieves cluster information
	GetCluster(ctx context.Context, externalID string) (*Cluster, error)
	// UpdateCluster updates cluster configuration
	UpdateCluster(ctx context.Context, cluster *Cluster) error
	// DeleteCluster deprovisions a cluster
	DeleteCluster(ctx context.Context, externalID string) error
	// GetKubeConfig retrieves the kubeconfig for a cluster
	GetKubeConfig(ctx context.Context, externalID string) ([]byte, error)
	// ListClusters lists all managed clusters
	ListClusters(ctx context.Context) ([]*Cluster, error)
	// GetClusterHealth retrieves health status of a cluster
	GetClusterHealth(ctx context.Context, externalID string) (*ClusterHealth, error)
}

// ClusterHealth represents the health status of a cluster
type ClusterHealth struct {
	Status      ClusterStatus      `json:"status"`
	NodeCount   int32              `json:"node_count"`
	ReadyNodes  int32              `json:"ready_nodes"`
	CPUUsage    float64            `json:"cpu_usage"`
	MemoryUsage float64            `json:"memory_usage"`
	Conditions  []ClusterCondition `json:"conditions"`
}

// ClusterCondition represents a condition of a cluster
type ClusterCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GitOpsAdapter defines the interface for GitOps/CD systems (e.g., ArgoCD)
type GitOpsAdapter interface {
	// CreateApplication creates a new application in the GitOps system
	CreateApplication(ctx context.Context, service *Service, environment *Environment) (string, error)
	// UpdateApplication updates an existing application
	UpdateApplication(ctx context.Context, service *Service, environment *Environment) error
	// DeleteApplication removes an application from the GitOps system
	DeleteApplication(ctx context.Context, externalID string) error
	// SyncApplication triggers a sync for an application
	SyncApplication(ctx context.Context, externalID string) error
	// GetApplicationStatus retrieves the status of an application
	GetApplicationStatus(ctx context.Context, externalID string) (*ApplicationStatus, error)
	// GetApplicationHistory retrieves deployment history
	GetApplicationHistory(ctx context.Context, externalID string) ([]*Deployment, error)
	// RollbackApplication rolls back to a previous version
	RollbackApplication(ctx context.Context, externalID string, revision int64) error
}

// ApplicationStatus represents the status of a GitOps application
type ApplicationStatus struct {
	Health        string           `json:"health"`
	SyncStatus    string           `json:"sync_status"`
	CurrentImage  string           `json:"current_image"`
	DesiredImage  string           `json:"desired_image"`
	Replicas      int32            `json:"replicas"`
	ReadyReplicas int32            `json:"ready_replicas"`
	Resources     []ResourceStatus `json:"resources"`
}

// ResourceStatus represents the status of a Kubernetes resource
type ResourceStatus struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	Message   string `json:"message"`
}

// SecretsAdapter defines the interface for secrets management (e.g., Vault)
type SecretsAdapter interface {
	// CreateSecret creates a new secret
	CreateSecret(ctx context.Context, secret *Secret, data map[string][]byte) error
	// GetSecret retrieves a secret's data
	GetSecret(ctx context.Context, path string) (map[string][]byte, error)
	// UpdateSecret updates an existing secret
	UpdateSecret(ctx context.Context, secret *Secret, data map[string][]byte) error
	// DeleteSecret deletes a secret
	DeleteSecret(ctx context.Context, path string) error
	// ListSecrets lists secrets under a path
	ListSecrets(ctx context.Context, path string) ([]string, error)
	// CreateDynamicSecret creates a dynamic secret (e.g., database credentials)
	CreateDynamicSecret(ctx context.Context, name string, config map[string]interface{}) error
}

// EventBus defines the interface for event publishing and subscribing
type EventBus interface {
	// Publish publishes an event to a subject
	Publish(ctx context.Context, subject string, event *Event) error
	// Subscribe subscribes to events on a subject
	Subscribe(ctx context.Context, subject string, handler EventHandler) (Subscription, error)
	// QueueSubscribe subscribes to events with a queue group for load balancing
	QueueSubscribe(ctx context.Context, subject string, queue string, handler EventHandler) (Subscription, error)
	// Request sends a request and waits for a response
	Request(ctx context.Context, subject string, event *Event) (*Event, error)
	// Close closes the event bus connection
	Close() error
}

// Event represents an event in the system
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// EventHandler is a function that handles events
type EventHandler func(event *Event) error

// Subscription represents an event subscription
type Subscription interface {
	Unsubscribe() error
}

// KubernetesClient defines the interface for Kubernetes operations
type KubernetesClient interface {
	// ApplyManifest applies a Kubernetes manifest
	ApplyManifest(ctx context.Context, clusterID uuid.UUID, manifest []byte) error
	// DeleteResource deletes a Kubernetes resource
	DeleteResource(ctx context.Context, clusterID uuid.UUID, kind, namespace, name string) error
	// GetResource retrieves a Kubernetes resource
	GetResource(ctx context.Context, clusterID uuid.UUID, kind, namespace, name string) (map[string]interface{}, error)
	// ListResources lists Kubernetes resources
	ListResources(ctx context.Context, clusterID uuid.UUID, kind, namespace string, labels map[string]string) ([]map[string]interface{}, error)
	// GetPodLogs retrieves logs from a pod
	GetPodLogs(ctx context.Context, clusterID uuid.UUID, namespace, podName, container string, tailLines int64) (string, error)
	// ExecInPod executes a command in a pod
	ExecInPod(ctx context.Context, clusterID uuid.UUID, namespace, podName, container string, command []string) (string, error)
	// WatchResource watches for changes to a resource
	WatchResource(ctx context.Context, clusterID uuid.UUID, kind, namespace string, handler func(eventType string, obj map[string]interface{})) error
}

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	// GetServiceMetrics retrieves metrics for a service
	GetServiceMetrics(ctx context.Context, serviceID uuid.UUID, timeRange TimeRange) (*ServiceMetrics, error)
	// GetClusterMetrics retrieves metrics for a cluster
	GetClusterMetrics(ctx context.Context, clusterID uuid.UUID, timeRange TimeRange) (*ClusterMetrics, error)
	// GetProjectMetrics retrieves aggregated metrics for a project
	GetProjectMetrics(ctx context.Context, projectID uuid.UUID, timeRange TimeRange) (*ProjectMetrics, error)
}

// TimeRange defines a time range for metrics queries
type TimeRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
	Step  int64 `json:"step"`
}

// ServiceMetrics represents metrics for a service
type ServiceMetrics struct {
	ServiceID    uuid.UUID      `json:"service_id"`
	CPUUsage     []MetricPoint  `json:"cpu_usage"`
	MemoryUsage  []MetricPoint  `json:"memory_usage"`
	RequestCount []MetricPoint  `json:"request_count"`
	ErrorRate    []MetricPoint  `json:"error_rate"`
	Latency      LatencyMetrics `json:"latency"`
	Replicas     []MetricPoint  `json:"replicas"`
}

// ClusterMetrics represents metrics for a cluster
type ClusterMetrics struct {
	ClusterID   uuid.UUID     `json:"cluster_id"`
	CPUUsage    []MetricPoint `json:"cpu_usage"`
	MemoryUsage []MetricPoint `json:"memory_usage"`
	PodCount    []MetricPoint `json:"pod_count"`
	NodeCount   int32         `json:"node_count"`
	DiskUsage   []MetricPoint `json:"disk_usage"`
}

// ProjectMetrics represents aggregated metrics for a project
type ProjectMetrics struct {
	ProjectID     uuid.UUID     `json:"project_id"`
	ServiceCount  int           `json:"service_count"`
	TotalCPU      []MetricPoint `json:"total_cpu"`
	TotalMemory   []MetricPoint `json:"total_memory"`
	TotalRequests []MetricPoint `json:"total_requests"`
	TotalErrors   []MetricPoint `json:"total_errors"`
}

// MetricPoint represents a single metric data point
type MetricPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// LatencyMetrics represents latency percentiles
type LatencyMetrics struct {
	P50 []MetricPoint `json:"p50"`
	P90 []MetricPoint `json:"p90"`
	P99 []MetricPoint `json:"p99"`
}

// Notifier defines the interface for sending notifications
type Notifier interface {
	// SendNotification sends a notification
	SendNotification(ctx context.Context, notification *Notification) error
	// SendBuildNotification sends a build status notification
	SendBuildNotification(ctx context.Context, build *Build) error
	// SendDeploymentNotification sends a deployment status notification
	SendDeploymentNotification(ctx context.Context, deployment *Deployment) error
	// SendAlertNotification sends an alert notification
	SendAlertNotification(ctx context.Context, alert *Alert) error
}

// Notification represents a notification to send
type Notification struct {
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"` // slack, email, webhook, etc.
	Recipient string                 `json:"recipient"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Severity  string                 `json:"severity"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Alert represents an alert
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	Source      string            `json:"source"`
	Message     string            `json:"message"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    int64             `json:"starts_at"`
	EndsAt      int64             `json:"ends_at,omitempty"`
}
