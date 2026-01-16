// Package events contains domain event definitions following DDD patterns.
// Domain events represent something significant that happened in the domain.
package events

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the base interface for all domain events
type DomainEvent interface {
	EventID() uuid.UUID
	EventType() string
	AggregateID() uuid.UUID
	AggregateType() string
	OccurredAt() time.Time
	Version() int
}

// BaseEvent provides common fields for all domain events
type BaseEvent struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	AggregateId   uuid.UUID `json:"aggregate_id"`
	AggregateKind string    `json:"aggregate_type"`
	Occurred      time.Time `json:"occurred_at"`
	EventVersion  int       `json:"version"`
}

func (e BaseEvent) EventID() uuid.UUID     { return e.ID }
func (e BaseEvent) EventType() string      { return e.Type }
func (e BaseEvent) AggregateID() uuid.UUID { return e.AggregateId }
func (e BaseEvent) AggregateType() string  { return e.AggregateKind }
func (e BaseEvent) OccurredAt() time.Time  { return e.Occurred }
func (e BaseEvent) Version() int           { return e.EventVersion }

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType string, aggregateID uuid.UUID, aggregateType string, version int) BaseEvent {
	return BaseEvent{
		ID:            uuid.New(),
		Type:          eventType,
		AggregateId:   aggregateID,
		AggregateKind: aggregateType,
		Occurred:      time.Now(),
		EventVersion:  version,
	}
}

// ===== Project Events =====

// ProjectCreated is raised when a new project is created
type ProjectCreated struct {
	BaseEvent
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	OwnerID     string `json:"owner_id"`
	Description string `json:"description"`
}

// ProjectUpdated is raised when a project is updated
type ProjectUpdated struct {
	BaseEvent
	Changes map[string]interface{} `json:"changes"`
}

// ProjectDeleted is raised when a project is deleted
type ProjectDeleted struct {
	BaseEvent
	DeletedBy string `json:"deleted_by"`
}

// ===== Service Events =====

// ServiceCreated is raised when a new service is created
type ServiceCreated struct {
	BaseEvent
	ProjectID   uuid.UUID `json:"project_id"`
	Name        string    `json:"name"`
	ServiceType string    `json:"service_type"`
}

// ServiceDeployed is raised when a service is deployed
type ServiceDeployed struct {
	BaseEvent
	DeploymentID uuid.UUID `json:"deployment_id"`
	Environment  string    `json:"environment"`
	Version      string    `json:"version"`
	Image        string    `json:"image"`
}

// ServiceScaled is raised when a service is scaled
type ServiceScaled struct {
	BaseEvent
	PreviousReplicas int    `json:"previous_replicas"`
	NewReplicas      int    `json:"new_replicas"`
	Reason           string `json:"reason"`
}

// ServiceHealthChanged is raised when service health status changes
type ServiceHealthChanged struct {
	BaseEvent
	PreviousHealth string `json:"previous_health"`
	NewHealth      string `json:"new_health"`
	Reason         string `json:"reason"`
}

// ===== Cluster Events =====

// ClusterProvisioned is raised when a cluster is provisioned
type ClusterProvisioned struct {
	BaseEvent
	Provider    string `json:"provider"`
	Region      string `json:"region"`
	KubeVersion string `json:"kube_version"`
	NodeCount   int    `json:"node_count"`
}

// ClusterDeleted is raised when a cluster is deleted
type ClusterDeleted struct {
	BaseEvent
	DeletedBy string `json:"deleted_by"`
}

// ===== Database Events =====

// DatabaseCreated is raised when a database is created
type DatabaseCreated struct {
	BaseEvent
	ProjectID        uuid.UUID `json:"project_id"`
	DatabaseType     string    `json:"database_type"` // yugabytedb, postgresql, etc.
	Size             string    `json:"size"`
	HighAvailability bool      `json:"high_availability"`
}

// DatabaseScaled is raised when a database is scaled
type DatabaseScaled struct {
	BaseEvent
	PreviousReplicas int `json:"previous_replicas"`
	NewReplicas      int `json:"new_replicas"`
}

// ===== Build Events =====

// BuildStarted is raised when a build starts
type BuildStarted struct {
	BaseEvent
	ServiceID uuid.UUID `json:"service_id"`
	CommitSHA string    `json:"commit_sha"`
	Branch    string    `json:"branch"`
	Trigger   string    `json:"trigger"` // push, pr, manual
}

// BuildCompleted is raised when a build completes
type BuildCompleted struct {
	BaseEvent
	ServiceID uuid.UUID     `json:"service_id"`
	Status    string        `json:"status"` // success, failed
	Duration  time.Duration `json:"duration"`
	ImageTag  string        `json:"image_tag"`
	ErrorMsg  string        `json:"error_msg,omitempty"`
}

// ===== User Events =====

// UserRegistered is raised when a user registers
type UserRegistered struct {
	BaseEvent
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UserLoggedIn is raised when a user logs in
type UserLoggedIn struct {
	BaseEvent
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// EventStore provides methods for persisting and retrieving events
type EventStore interface {
	// Save persists events to the store
	Save(events []DomainEvent) error
	// GetEvents retrieves events for an aggregate
	GetEvents(aggregateID uuid.UUID, fromVersion int) ([]DomainEvent, error)
	// GetAllEvents retrieves all events (for rebuilding projections)
	GetAllEvents(fromEventID uuid.UUID, limit int) ([]DomainEvent, error)
}

// EventPublisher publishes events to external systems
type EventPublisher interface {
	// Publish publishes an event
	Publish(event DomainEvent) error
	// PublishBatch publishes multiple events
	PublishBatch(events []DomainEvent) error
}
