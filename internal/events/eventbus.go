package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/northstack/platform/internal/config"
)

type EventBus struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	config config.NATSConfig
}

func NewEventBus(cfg config.NATSConfig) (*EventBus, error) {
	// Connect to NATS
	nc, err := nats.Connect(cfg.URL,
		nats.Name(cfg.ClientID),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	bus := &EventBus{
		nc:     nc,
		js:     js,
		config: cfg,
	}

	// Initialize streams if enabled
	if cfg.JetStreamEnabled {
		if err := bus.ensureStreams(context.Background()); err != nil {
			return nil, err
		}
	}

	return bus, nil
}

func (b *EventBus) ensureStreams(ctx context.Context) error {
	// Define default streams if none provided in config, or use config
	// This is a simplified bootstrap implementation
	streams := []jetstream.StreamConfig{
		{
			Name:     "DEPLOYMENTS",
			Subjects: []string{"deployment.>"},
			MaxAge:   30 * 24 * time.Hour,
		},
		{
			Name:     "APPLICATIONS",
			Subjects: []string{"application.>"},
			MaxAge:   30 * 24 * time.Hour,
		},
		{
			Name:     "AUDIT",
			Subjects: []string{"audit.>"},
			MaxAge:   365 * 24 * time.Hour,
		},
	}

	for _, cfg := range streams {
		_, err := b.js.CreateOrUpdateStream(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to create stream %s: %w", cfg.Name, err)
		}
	}
	return nil
}

func (b *EventBus) Publish(ctx context.Context, subject string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to JetStream
	_, err = b.js.Publish(ctx, subject, data)
	return err
}

func (b *EventBus) PublishAsync(subject string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// PublishAsync isn't directly on new JetStream interface like legacy JS
	// We use the synchronous Publish with a context for now, or would use a wrapper goroutine
	// For true async high-throughput, we'd use the AsyncPublish on the nats.Conn or custom buffer
	// Fallback to core NATS for async fire-and-forget if JS not strictly required for this call
	return b.nc.Publish(subject, data)
}

func (b *EventBus) Subscribe(ctx context.Context, subject string, handler func(Event) error) (jetstream.Consumer, error) {
	// Create a consumer for the stream
	// This assumes the subject maps to a known stream
	// In a real impl, we'd look up the stream name
	streamName := "DEPLOYMENTS" // simplistic default

	consumer, err := b.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       "processor-" + subject, // Durable name
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: subject,
	})
	if err != nil {
		return nil, err
	}

	// Start consuming
	_, err = consumer.Consume(func(msg jetstream.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			msg.Nak()
			return
		}

		if err := handler(event); err != nil {
			msg.Nak() // Retry
		} else {
			msg.Ack()
		}
	})

	return consumer, err
}

func (b *EventBus) Health() error {
	if b.nc.Status() != nats.CONNECTED {
		return fmt.Errorf("nats not connected")
	}
	return nil
}

func (b *EventBus) Close() {
	b.nc.Close()
}
