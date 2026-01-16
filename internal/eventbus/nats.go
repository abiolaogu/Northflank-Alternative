// Package eventbus provides event publishing and subscribing functionality using NATS.
// It implements the EventBus interface from the domain package.
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/openpaas/platform-orchestrator/internal/config"
	"github.com/openpaas/platform-orchestrator/internal/domain"
	"github.com/openpaas/platform-orchestrator/pkg/logger"
)

// Event subjects for the platform
const (
	SubjectBuildStarted    = "build.started"
	SubjectBuildCompleted  = "build.completed"
	SubjectBuildFailed     = "build.failed"
	SubjectDeployStarted   = "deploy.started"
	SubjectDeployCompleted = "deploy.completed"
	SubjectDeployFailed    = "deploy.failed"
	SubjectServiceCreated  = "service.created"
	SubjectServiceUpdated  = "service.updated"
	SubjectServiceDeleted  = "service.deleted"
	SubjectServiceScaled   = "service.scaled"
	SubjectProjectCreated  = "project.created"
	SubjectProjectDeleted  = "project.deleted"
	SubjectClusterCreated  = "cluster.created"
	SubjectClusterUpdated  = "cluster.updated"
	SubjectClusterDeleted  = "cluster.deleted"
	SubjectSecretCreated   = "secret.created"
	SubjectSecretUpdated   = "secret.updated"
	SubjectSecretDeleted   = "secret.deleted"
	SubjectAlertFired      = "alert.fired"
	SubjectAlertResolved   = "alert.resolved"
	SubjectWebhookReceived = "webhook.received"
	SubjectAuditLog        = "audit.log"
)

// NATSEventBus implements the EventBus interface using NATS
type NATSEventBus struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config *config.NATSConfig
	logger *logger.Logger
	subs   []*nats.Subscription
	mu     sync.RWMutex
	closed bool
}

// natsSubscription wraps a NATS subscription
type natsSubscription struct {
	sub *nats.Subscription
}

func (s *natsSubscription) Unsubscribe() error {
	return s.sub.Unsubscribe()
}

// NewNATSEventBus creates a new NATS event bus
func NewNATSEventBus(cfg *config.NATSConfig, log *logger.Logger) (*NATSEventBus, error) {
	opts := []nats.Option{
		nats.Name(cfg.ClientID),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Warn().Err(err).Msg("NATS disconnected")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info().Str("url", nc.ConnectedUrl()).Msg("NATS reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info().Msg("NATS connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Error().Err(err).Str("subject", sub.Subject).Msg("NATS error")
		}),
	}

	// Add authentication if configured
	if cfg.Token != "" {
		opts = append(opts, nats.Token(cfg.Token))
	} else if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, nats.UserInfo(cfg.Username, cfg.Password))
	}

	// Add TLS if configured
	if cfg.TLSEnabled {
		opts = append(opts, nats.Secure())
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			opts = append(opts, nats.ClientCert(cfg.TLSCertFile, cfg.TLSKeyFile))
		}
		if cfg.TLSCAFile != "" {
			opts = append(opts, nats.RootCAs(cfg.TLSCAFile))
		}
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	bus := &NATSEventBus{
		conn:   conn,
		config: cfg,
		logger: log,
		subs:   make([]*nats.Subscription, 0),
	}

	// Initialize JetStream if enabled
	if cfg.JetStreamEnabled {
		js, err := conn.JetStream()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create JetStream context: %w", err)
		}
		bus.js = js

		// Create streams for durable event storage
		if err := bus.createStreams(); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create streams: %w", err)
		}
	}

	log.Info().Str("url", cfg.URL).Bool("jetstream", cfg.JetStreamEnabled).Msg("Connected to NATS")

	return bus, nil
}

// createStreams creates JetStream streams for event persistence
func (b *NATSEventBus) createStreams() error {
	streams := []struct {
		name     string
		subjects []string
	}{
		{
			name:     "BUILDS",
			subjects: []string{"build.>"},
		},
		{
			name:     "DEPLOYMENTS",
			subjects: []string{"deploy.>"},
		},
		{
			name:     "SERVICES",
			subjects: []string{"service.>"},
		},
		{
			name:     "PROJECTS",
			subjects: []string{"project.>"},
		},
		{
			name:     "CLUSTERS",
			subjects: []string{"cluster.>"},
		},
		{
			name:     "SECRETS",
			subjects: []string{"secret.>"},
		},
		{
			name:     "ALERTS",
			subjects: []string{"alert.>"},
		},
		{
			name:     "WEBHOOKS",
			subjects: []string{"webhook.>"},
		},
		{
			name:     "AUDIT",
			subjects: []string{"audit.>"},
		},
	}

	for _, stream := range streams {
		_, err := b.js.AddStream(&nats.StreamConfig{
			Name:      stream.name,
			Subjects:  stream.subjects,
			Retention: nats.LimitsPolicy,
			MaxAge:    7 * 24 * time.Hour, // Keep events for 7 days
			MaxBytes:  1024 * 1024 * 1024, // 1GB max
			Discard:   nats.DiscardOld,
			Storage:   nats.FileStorage,
			Replicas:  1,
		})
		if err != nil && err != nats.ErrStreamNameAlreadyInUse {
			// Stream might already exist, try to update
			_, err = b.js.UpdateStream(&nats.StreamConfig{
				Name:      stream.name,
				Subjects:  stream.subjects,
				Retention: nats.LimitsPolicy,
				MaxAge:    7 * 24 * time.Hour,
				MaxBytes:  1024 * 1024 * 1024,
				Discard:   nats.DiscardOld,
				Storage:   nats.FileStorage,
				Replicas:  1,
			})
			if err != nil {
				b.logger.Warn().Err(err).Str("stream", stream.name).Msg("Failed to create/update stream")
			}
		}
	}

	return nil
}

// Publish publishes an event to a subject
func (b *NATSEventBus) Publish(ctx context.Context, subject string, event *domain.Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return fmt.Errorf("event bus is closed")
	}
	b.mu.RUnlock()

	// Set event ID and timestamp if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixNano()
	}
	event.Subject = subject

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Use JetStream if available for durability
	if b.js != nil {
		_, err = b.js.Publish(subject, data)
	} else {
		err = b.conn.Publish(subject, data)
	}

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	b.logger.Debug().
		Str("subject", subject).
		Str("event_id", event.ID).
		Str("event_type", event.Type).
		Msg("Event published")

	return nil
}

// Subscribe subscribes to events on a subject
func (b *NATSEventBus) Subscribe(ctx context.Context, subject string, handler domain.EventHandler) (domain.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("event bus is closed")
	}

	sub, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {
		var event domain.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			b.logger.Error().Err(err).Str("subject", subject).Msg("Failed to unmarshal event")
			return
		}

		if err := handler(&event); err != nil {
			b.logger.Error().Err(err).Str("subject", subject).Str("event_id", event.ID).Msg("Event handler error")
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	b.subs = append(b.subs, sub)
	b.logger.Debug().Str("subject", subject).Msg("Subscribed to subject")

	return &natsSubscription{sub: sub}, nil
}

// QueueSubscribe subscribes with a queue group for load balancing
func (b *NATSEventBus) QueueSubscribe(ctx context.Context, subject string, queue string, handler domain.EventHandler) (domain.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("event bus is closed")
	}

	sub, err := b.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		var event domain.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			b.logger.Error().Err(err).Str("subject", subject).Msg("Failed to unmarshal event")
			return
		}

		if err := handler(&event); err != nil {
			b.logger.Error().Err(err).Str("subject", subject).Str("event_id", event.ID).Msg("Event handler error")
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe: %w", err)
	}

	b.subs = append(b.subs, sub)
	b.logger.Debug().Str("subject", subject).Str("queue", queue).Msg("Queue subscribed to subject")

	return &natsSubscription{sub: sub}, nil
}

// Request sends a request and waits for a response
func (b *NATSEventBus) Request(ctx context.Context, subject string, event *domain.Event) (*domain.Event, error) {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return nil, fmt.Errorf("event bus is closed")
	}
	b.mu.RUnlock()

	// Set event ID and timestamp if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixNano()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Get timeout from context or use default
	timeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	msg, err := b.conn.Request(subject, data, timeout)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var response domain.Event
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// Close closes the event bus connection
func (b *NATSEventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Unsubscribe all subscriptions
	for _, sub := range b.subs {
		if err := sub.Unsubscribe(); err != nil {
			b.logger.Warn().Err(err).Msg("Failed to unsubscribe")
		}
	}

	// Drain and close connection
	if err := b.conn.Drain(); err != nil {
		b.logger.Warn().Err(err).Msg("Failed to drain NATS connection")
	}

	b.logger.Info().Msg("NATS event bus closed")
	return nil
}

// PublishBuildEvent publishes a build-related event
func (b *NATSEventBus) PublishBuildEvent(ctx context.Context, eventType string, build *domain.Build) error {
	event := &domain.Event{
		Type:   eventType,
		Source: "platform-orchestrator",
		Data: map[string]interface{}{
			"build_id":   build.ID.String(),
			"service_id": build.ServiceID.String(),
			"project_id": build.ProjectID.String(),
			"status":     string(build.Status),
			"image_tag":  build.ImageTag,
		},
	}

	var subject string
	switch eventType {
	case "build.started":
		subject = SubjectBuildStarted
	case "build.completed":
		subject = SubjectBuildCompleted
	case "build.failed":
		subject = SubjectBuildFailed
		event.Data["error"] = build.ErrorMessage
	default:
		subject = "build." + eventType
	}

	return b.Publish(ctx, subject, event)
}

// PublishDeploymentEvent publishes a deployment-related event
func (b *NATSEventBus) PublishDeploymentEvent(ctx context.Context, eventType string, deployment *domain.Deployment) error {
	event := &domain.Event{
		Type:   eventType,
		Source: "platform-orchestrator",
		Data: map[string]interface{}{
			"deployment_id": deployment.ID.String(),
			"service_id":    deployment.ServiceID.String(),
			"project_id":    deployment.ProjectID.String(),
			"cluster_id":    deployment.ClusterID.String(),
			"status":        string(deployment.Status),
			"version":       deployment.Version,
			"replicas":      deployment.Replicas,
		},
	}

	var subject string
	switch eventType {
	case "deploy.started":
		subject = SubjectDeployStarted
	case "deploy.completed":
		subject = SubjectDeployCompleted
	case "deploy.failed":
		subject = SubjectDeployFailed
		event.Data["error"] = deployment.ErrorMessage
	default:
		subject = "deploy." + eventType
	}

	return b.Publish(ctx, subject, event)
}

// PublishServiceEvent publishes a service-related event
func (b *NATSEventBus) PublishServiceEvent(ctx context.Context, eventType string, service *domain.Service) error {
	event := &domain.Event{
		Type:   eventType,
		Source: "platform-orchestrator",
		Data: map[string]interface{}{
			"service_id": service.ID.String(),
			"project_id": service.ProjectID.String(),
			"name":       service.Name,
			"type":       string(service.Type),
			"status":     string(service.Status),
		},
	}

	var subject string
	switch eventType {
	case "service.created":
		subject = SubjectServiceCreated
	case "service.updated":
		subject = SubjectServiceUpdated
	case "service.deleted":
		subject = SubjectServiceDeleted
	case "service.scaled":
		subject = SubjectServiceScaled
	default:
		subject = "service." + eventType
	}

	return b.Publish(ctx, subject, event)
}

// PublishAuditEvent publishes an audit log event
func (b *NATSEventBus) PublishAuditEvent(ctx context.Context, auditLog *domain.AuditLog) error {
	event := &domain.Event{
		Type:   string(auditLog.Action),
		Source: "platform-orchestrator",
		Data: map[string]interface{}{
			"audit_id":      auditLog.ID.String(),
			"user_id":       auditLog.UserID.String(),
			"action":        string(auditLog.Action),
			"resource_type": auditLog.ResourceType,
			"resource_id":   auditLog.ResourceID.String(),
			"resource_name": auditLog.ResourceName,
			"ip_address":    auditLog.IPAddress,
		},
	}

	if auditLog.ProjectID != nil {
		event.Data["project_id"] = auditLog.ProjectID.String()
	}

	return b.Publish(ctx, SubjectAuditLog, event)
}
