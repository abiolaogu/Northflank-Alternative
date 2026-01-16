// Package audit provides audit logging functionality for the Platform Orchestrator.
// It tracks all user actions and system events for compliance and debugging.
package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/openpaas/platform-orchestrator/internal/domain"
	"github.com/openpaas/platform-orchestrator/pkg/logger"
)

// Logger handles audit logging
type Logger struct {
	repo     domain.AuditLogRepository
	eventBus domain.EventBus
	logger   *logger.Logger
}

// NewLogger creates a new audit Logger
func NewLogger(repo domain.AuditLogRepository, eventBus domain.EventBus, log *logger.Logger) *Logger {
	return &Logger{
		repo:     repo,
		eventBus: eventBus,
		logger:   log,
	}
}

// LogOptions contains options for creating an audit log entry
type LogOptions struct {
	UserID       uuid.UUID
	Action       domain.AuditAction
	ResourceType string
	ResourceID   uuid.UUID
	ResourceName string
	ProjectID    *uuid.UUID
	IPAddress    string
	UserAgent    string
	OldValue     map[string]interface{}
	NewValue     map[string]interface{}
	Metadata     map[string]interface{}
}

// Log creates a new audit log entry
func (l *Logger) Log(ctx context.Context, opts LogOptions) error {
	entry := &domain.AuditLog{
		ID:           uuid.New(),
		UserID:       opts.UserID,
		Action:       opts.Action,
		ResourceType: opts.ResourceType,
		ResourceID:   opts.ResourceID,
		ResourceName: opts.ResourceName,
		ProjectID:    opts.ProjectID,
		IPAddress:    opts.IPAddress,
		UserAgent:    opts.UserAgent,
		OldValue:     opts.OldValue,
		NewValue:     opts.NewValue,
		Metadata:     opts.Metadata,
		CreatedAt:    time.Now(),
	}

	// Persist to database
	if err := l.repo.Create(ctx, entry); err != nil {
		l.logger.Error().Err(err).Str("action", string(opts.Action)).Msg("Failed to create audit log")
		return err
	}

	// Publish event for real-time monitoring
	if l.eventBus != nil {
		event := &domain.Event{
			Type:   "audit." + string(opts.Action),
			Source: "audit-logger",
			Data: map[string]interface{}{
				"audit_id":      entry.ID.String(),
				"user_id":       entry.UserID.String(),
				"action":        string(entry.Action),
				"resource_type": entry.ResourceType,
				"resource_id":   entry.ResourceID.String(),
			},
		}
		l.eventBus.Publish(ctx, "audit.log", event)
	}

	l.logger.Debug().
		Str("audit_id", entry.ID.String()).
		Str("action", string(entry.Action)).
		Str("resource_type", entry.ResourceType).
		Msg("Audit log created")

	return nil
}

// LogCreate logs a create action
func (l *Logger) LogCreate(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, resourceName string, projectID *uuid.UUID, value map[string]interface{}) error {
	return l.Log(ctx, LogOptions{
		UserID:       userID,
		Action:       domain.AuditActionCreate,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		ProjectID:    projectID,
		NewValue:     value,
	})
}

// LogUpdate logs an update action
func (l *Logger) LogUpdate(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, resourceName string, projectID *uuid.UUID, oldValue, newValue map[string]interface{}) error {
	return l.Log(ctx, LogOptions{
		UserID:       userID,
		Action:       domain.AuditActionUpdate,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		ProjectID:    projectID,
		OldValue:     oldValue,
		NewValue:     newValue,
	})
}

// LogDelete logs a delete action
func (l *Logger) LogDelete(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, resourceName string, projectID *uuid.UUID, value map[string]interface{}) error {
	return l.Log(ctx, LogOptions{
		UserID:       userID,
		Action:       domain.AuditActionDelete,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		ProjectID:    projectID,
		OldValue:     value,
	})
}

// LogDeploy logs a deployment action
func (l *Logger) LogDeploy(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID, serviceName string, projectID *uuid.UUID, metadata map[string]interface{}) error {
	return l.Log(ctx, LogOptions{
		UserID:       userID,
		Action:       domain.AuditActionDeploy,
		ResourceType: "service",
		ResourceID:   serviceID,
		ResourceName: serviceName,
		ProjectID:    projectID,
		Metadata:     metadata,
	})
}

// LogBuild logs a build action
func (l *Logger) LogBuild(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID, serviceName string, projectID *uuid.UUID, metadata map[string]interface{}) error {
	return l.Log(ctx, LogOptions{
		UserID:       userID,
		Action:       domain.AuditActionBuild,
		ResourceType: "service",
		ResourceID:   serviceID,
		ResourceName: serviceName,
		ProjectID:    projectID,
		Metadata:     metadata,
	})
}

// LogLogin logs a login action
func (l *Logger) LogLogin(ctx context.Context, userID uuid.UUID, email string, ipAddress, userAgent string, success bool) error {
	metadata := map[string]interface{}{
		"success": success,
		"email":   email,
	}
	return l.Log(ctx, LogOptions{
		UserID:       userID,
		Action:       domain.AuditActionLogin,
		ResourceType: "user",
		ResourceID:   userID,
		ResourceName: email,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Metadata:     metadata,
	})
}
