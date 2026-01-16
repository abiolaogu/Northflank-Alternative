// Package domain contains the core domain models and interfaces for the Platform Orchestrator.
// This is the heart of the system, defining the fundamental entities that the platform manages.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProjectStatus represents the current state of a project
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusInactive ProjectStatus = "inactive"
	ProjectStatusPending  ProjectStatus = "pending"
	ProjectStatusDeleting ProjectStatus = "deleting"
)

// Project represents a collection of services and resources
type Project struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description,omitempty"`
	Status      ProjectStatus          `json:"status"`
	OwnerID     uuid.UUID              `json:"owner_id"`
	TeamID      *uuid.UUID             `json:"team_id,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ServiceType represents the type of service being deployed
type ServiceType string

const (
	ServiceTypeWebApp     ServiceType = "webapp"
	ServiceTypeWorker     ServiceType = "worker"
	ServiceTypeCronJob    ServiceType = "cronjob"
	ServiceTypeStatefulDB ServiceType = "stateful_db"
	ServiceTypeStateless  ServiceType = "stateless"
)

// ServiceStatus represents the current state of a service
type ServiceStatus string

const (
	ServiceStatusPending     ServiceStatus = "pending"
	ServiceStatusBuilding    ServiceStatus = "building"
	ServiceStatusDeploying   ServiceStatus = "deploying"
	ServiceStatusRunning     ServiceStatus = "running"
	ServiceStatusStopped     ServiceStatus = "stopped"
	ServiceStatusFailed      ServiceStatus = "failed"
	ServiceStatusTerminating ServiceStatus = "terminating"
)

// BuildSource defines where the code comes from
type BuildSource struct {
	Type       string `json:"type"` // "git", "docker", "buildpack"
	Repository string `json:"repository,omitempty"`
	Branch     string `json:"branch,omitempty"`
	CommitSHA  string `json:"commit_sha,omitempty"`
	Dockerfile string `json:"dockerfile,omitempty"`
	Image      string `json:"image,omitempty"`
	Registry   string `json:"registry,omitempty"`
}

// ResourceLimits defines the compute resources for a service
type ResourceLimits struct {
	CPURequest    string `json:"cpu_request,omitempty"`
	CPULimit      string `json:"cpu_limit,omitempty"`
	MemoryRequest string `json:"memory_request,omitempty"`
	MemoryLimit   string `json:"memory_limit,omitempty"`
	StorageSize   string `json:"storage_size,omitempty"`
}

// ScalingConfig defines how a service should scale
type ScalingConfig struct {
	MinReplicas          int32 `json:"min_replicas"`
	MaxReplicas          int32 `json:"max_replicas"`
	TargetCPU            int32 `json:"target_cpu,omitempty"`
	TargetMemory         int32 `json:"target_memory,omitempty"`
	ScaleDownDelay       int32 `json:"scale_down_delay,omitempty"`
	ScaleUpStabilization int32 `json:"scale_up_stabilization,omitempty"`
}

// HealthCheck defines the health check configuration
type HealthCheck struct {
	Type                string `json:"type"` // "http", "tcp", "exec"
	Path                string `json:"path,omitempty"`
	Port                int32  `json:"port,omitempty"`
	Command             string `json:"command,omitempty"`
	InitialDelaySeconds int32  `json:"initial_delay_seconds"`
	PeriodSeconds       int32  `json:"period_seconds"`
	TimeoutSeconds      int32  `json:"timeout_seconds"`
	FailureThreshold    int32  `json:"failure_threshold"`
	SuccessThreshold    int32  `json:"success_threshold"`
}

// Service represents a deployable unit within a project
type Service struct {
	ID              uuid.UUID              `json:"id"`
	ProjectID       uuid.UUID              `json:"project_id"`
	Name            string                 `json:"name"`
	Slug            string                 `json:"slug"`
	Type            ServiceType            `json:"type"`
	Status          ServiceStatus          `json:"status"`
	BuildSource     BuildSource            `json:"build_source"`
	Resources       ResourceLimits         `json:"resources"`
	Scaling         ScalingConfig          `json:"scaling"`
	HealthCheck     *HealthCheck           `json:"health_check,omitempty"`
	EnvVars         map[string]string      `json:"env_vars,omitempty"`
	SecretRefs      []string               `json:"secret_refs,omitempty"`
	Ports           []ServicePort          `json:"ports,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	Annotations     map[string]string      `json:"annotations,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CurrentBuildID  *uuid.UUID             `json:"current_build_id,omitempty"`
	CurrentVersion  string                 `json:"current_version,omitempty"`
	TargetClusterID *uuid.UUID             `json:"target_cluster_id,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ServicePort defines a port exposed by a service
type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"target_port"`
	Protocol   string `json:"protocol"` // TCP, UDP
	Public     bool   `json:"public"`
}

// BuildStatus represents the current state of a build
type BuildStatus string

const (
	BuildStatusQueued    BuildStatus = "queued"
	BuildStatusRunning   BuildStatus = "running"
	BuildStatusSucceeded BuildStatus = "succeeded"
	BuildStatusFailed    BuildStatus = "failed"
	BuildStatusCanceled  BuildStatus = "canceled"
)

// Build represents a build job for a service
type Build struct {
	ID           uuid.UUID              `json:"id"`
	ServiceID    uuid.UUID              `json:"service_id"`
	ProjectID    uuid.UUID              `json:"project_id"`
	Status       BuildStatus            `json:"status"`
	Source       BuildSource            `json:"source"`
	ImageTag     string                 `json:"image_tag,omitempty"`
	ImageDigest  string                 `json:"image_digest,omitempty"`
	BuildLogs    string                 `json:"build_logs,omitempty"`
	Duration     int64                  `json:"duration,omitempty"`
	TriggeredBy  string                 `json:"triggered_by"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// DeploymentStatus represents the current state of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "pending"
	DeploymentStatusInProgress DeploymentStatus = "in_progress"
	DeploymentStatusSucceeded  DeploymentStatus = "succeeded"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back"
)

// DeploymentStrategy defines how a deployment should be rolled out
type DeploymentStrategy string

const (
	DeploymentStrategyRollingUpdate DeploymentStrategy = "rolling_update"
	DeploymentStrategyBlueGreen     DeploymentStrategy = "blue_green"
	DeploymentStrategyCanary        DeploymentStrategy = "canary"
	DeploymentStrategyRecreate      DeploymentStrategy = "recreate"
)

// Deployment represents a deployment of a service version
type Deployment struct {
	ID              uuid.UUID              `json:"id"`
	ServiceID       uuid.UUID              `json:"service_id"`
	ProjectID       uuid.UUID              `json:"project_id"`
	BuildID         uuid.UUID              `json:"build_id"`
	ClusterID       uuid.UUID              `json:"cluster_id"`
	Status          DeploymentStatus       `json:"status"`
	Strategy        DeploymentStrategy     `json:"strategy"`
	Version         string                 `json:"version"`
	PreviousVersion string                 `json:"previous_version,omitempty"`
	Replicas        int32                  `json:"replicas"`
	ReadyReplicas   int32                  `json:"ready_replicas"`
	TriggeredBy     string                 `json:"triggered_by"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// ClusterProvider represents the cloud provider for a cluster
type ClusterProvider string

const (
	ClusterProviderAWS          ClusterProvider = "aws"
	ClusterProviderGCP          ClusterProvider = "gcp"
	ClusterProviderAzure        ClusterProvider = "azure"
	ClusterProviderDigitalOcean ClusterProvider = "digitalocean"
	ClusterProviderLinode       ClusterProvider = "linode"
	ClusterProviderOnPrem       ClusterProvider = "on_prem"
	ClusterProviderK3s          ClusterProvider = "k3s"
)

// ClusterStatus represents the current state of a cluster
type ClusterStatus string

const (
	ClusterStatusProvisioning ClusterStatus = "provisioning"
	ClusterStatusActive       ClusterStatus = "active"
	ClusterStatusUpgrading    ClusterStatus = "upgrading"
	ClusterStatusUnhealthy    ClusterStatus = "unhealthy"
	ClusterStatusDeleting     ClusterStatus = "deleting"
)

// Cluster represents a Kubernetes cluster managed by the platform
type Cluster struct {
	ID               uuid.UUID              `json:"id"`
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Provider         ClusterProvider        `json:"provider"`
	Region           string                 `json:"region"`
	Status           ClusterStatus          `json:"status"`
	KubeVersion      string                 `json:"kube_version"`
	APIEndpoint      string                 `json:"api_endpoint,omitempty"`
	NodeCount        int32                  `json:"node_count"`
	Labels           map[string]string      `json:"labels,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	RancherClusterID string                 `json:"rancher_cluster_id,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// EnvironmentType represents the type of environment
type EnvironmentType string

const (
	EnvironmentTypeDevelopment EnvironmentType = "development"
	EnvironmentTypeStaging     EnvironmentType = "staging"
	EnvironmentTypeProduction  EnvironmentType = "production"
	EnvironmentTypePreview     EnvironmentType = "preview"
)

// Environment represents a deployment environment (dev, staging, prod, etc.)
type Environment struct {
	ID        uuid.UUID              `json:"id"`
	ProjectID uuid.UUID              `json:"project_id"`
	ClusterID uuid.UUID              `json:"cluster_id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Type      EnvironmentType        `json:"type"`
	Namespace string                 `json:"namespace"`
	IsDefault bool                   `json:"is_default"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SecretType represents the type of secret
type SecretType string

const (
	SecretTypeOpaque       SecretType = "opaque"
	SecretTypeTLS          SecretType = "tls"
	SecretTypeDockerConfig SecretType = "docker_config"
	SecretTypeSSHAuth      SecretType = "ssh_auth"
	SecretTypeBasicAuth    SecretType = "basic_auth"
)

// Secret represents a secret managed by the platform (stored in Vault)
type Secret struct {
	ID        uuid.UUID         `json:"id"`
	ProjectID uuid.UUID         `json:"project_id"`
	Name      string            `json:"name"`
	Type      SecretType        `json:"type"`
	Keys      []string          `json:"keys"` // Only store key names, not values
	VaultPath string            `json:"vault_path"`
	Version   int               `json:"version"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// IngressType represents the type of ingress
type IngressType string

const (
	IngressTypeHTTP IngressType = "http"
	IngressTypeGRPC IngressType = "grpc"
	IngressTypeTCP  IngressType = "tcp"
)

// TLSConfig defines TLS configuration for an ingress
type TLSConfig struct {
	Enabled    bool   `json:"enabled"`
	SecretName string `json:"secret_name,omitempty"`
	AutoTLS    bool   `json:"auto_tls"` // Let's Encrypt via cert-manager
}

// Ingress represents an ingress route for a service
type Ingress struct {
	ID          uuid.UUID         `json:"id"`
	ServiceID   uuid.UUID         `json:"service_id"`
	ProjectID   uuid.UUID         `json:"project_id"`
	Domain      string            `json:"domain"`
	Path        string            `json:"path"`
	Type        IngressType       `json:"type"`
	TLS         TLSConfig         `json:"tls"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PipelineStatus represents the status of a pipeline
type PipelineStatus string

const (
	PipelineStatusPending   PipelineStatus = "pending"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusSucceeded PipelineStatus = "succeeded"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCanceled  PipelineStatus = "canceled"
)

// PipelineStage defines a stage in a pipeline
type PipelineStage struct {
	Name      string         `json:"name"`
	Status    PipelineStatus `json:"status"`
	StartedAt *time.Time     `json:"started_at,omitempty"`
	EndedAt   *time.Time     `json:"ended_at,omitempty"`
	Logs      string         `json:"logs,omitempty"`
}

// Pipeline represents a CI/CD pipeline run
type Pipeline struct {
	ID           uuid.UUID              `json:"id"`
	ServiceID    uuid.UUID              `json:"service_id"`
	ProjectID    uuid.UUID              `json:"project_id"`
	Status       PipelineStatus         `json:"status"`
	Trigger      string                 `json:"trigger"` // "push", "pr", "manual", "schedule"
	Branch       string                 `json:"branch"`
	CommitSHA    string                 `json:"commit_sha"`
	Stages       []PipelineStage        `json:"stages"`
	BuildID      *uuid.UUID             `json:"build_id,omitempty"`
	DeploymentID *uuid.UUID             `json:"deployment_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleOwner  UserRole = "owner"
	UserRoleMember UserRole = "member"
	UserRoleViewer UserRole = "viewer"
)

// User represents a platform user
type User struct {
	ID          uuid.UUID         `json:"id"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	AvatarURL   string            `json:"avatar_url,omitempty"`
	Role        UserRole          `json:"role"`
	IsActive    bool              `json:"is_active"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Team represents a team of users
type Team struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description string            `json:"description,omitempty"`
	OwnerID     uuid.UUID         `json:"owner_id"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// TeamMembership represents a user's membership in a team
type TeamMembership struct {
	ID        uuid.UUID `json:"id"`
	TeamID    uuid.UUID `json:"team_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditAction represents an auditable action
type AuditAction string

const (
	AuditActionCreate  AuditAction = "create"
	AuditActionUpdate  AuditAction = "update"
	AuditActionDelete  AuditAction = "delete"
	AuditActionDeploy  AuditAction = "deploy"
	AuditActionBuild   AuditAction = "build"
	AuditActionScale   AuditAction = "scale"
	AuditActionRestart AuditAction = "restart"
	AuditActionLogin   AuditAction = "login"
	AuditActionLogout  AuditAction = "logout"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"user_id"`
	Action       AuditAction            `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   uuid.UUID              `json:"resource_id"`
	ResourceName string                 `json:"resource_name"`
	ProjectID    *uuid.UUID             `json:"project_id,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	OldValue     map[string]interface{} `json:"old_value,omitempty"`
	NewValue     map[string]interface{} `json:"new_value,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}
