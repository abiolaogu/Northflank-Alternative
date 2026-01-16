package events

import "time"

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`   // e.g. deployment.started
	Source    string                 `json:"source"` // e.g. orchestrator
	Subject   string                 `json:"subject"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  EventMetadata          `json:"metadata"`
}

// EventMetadata contains tracing and correlation info
type EventMetadata struct {
	CorrelationID string `json:"correlationId"`
	CausationID   string `json:"causationId,omitempty"`
	UserID        string `json:"userId,omitempty"`
	TraceID       string `json:"traceId,omitempty"`
}

// Common Event Types
const (
	TypeDeploymentStarted   = "deployment.started"
	TypeDeploymentFinished  = "deployment.finished"
	TypeDeploymentFailed    = "deployment.failed"
	TypeApplicationCreated  = "application.created"
	TypeApplicationUpdated  = "application.updated"
	TypeClusterProvisioning = "cluster.provisioning"
	TypeClusterReady        = "cluster.ready"
	TypeAuditLog            = "audit.log"
)
