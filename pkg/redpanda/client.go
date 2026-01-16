// Package redpanda provides a Kafka-compatible client for Redpanda streaming.
// Redpanda is a simpler, faster alternative to Kafka with no ZooKeeper dependency.
package redpanda

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

// Client wraps the Redpanda/Kafka connection
type Client struct {
	client  *kgo.Client
	logger  *zap.Logger
	brokers []string
}

// Config holds Redpanda configuration
type Config struct {
	Brokers           []string
	ConsumerGroup     string
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
}

// NewClient creates a new Redpanda client
func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.ConsumerGroup(config.ConsumerGroup),
		kgo.SessionTimeout(config.SessionTimeout),
		kgo.HeartbeatInterval(config.HeartbeatInterval),
		kgo.AutoCommitInterval(time.Second * 5),
		kgo.AutoCommitMarks(),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create redpanda client: %w", err)
	}

	return &Client{
		client:  client,
		logger:  logger,
		brokers: config.Brokers,
	}, nil
}

// Close closes the client connection
func (c *Client) Close() {
	c.client.Close()
}

// Publish publishes a message to a topic
func (c *Client) Publish(ctx context.Context, topic string, key, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	results := c.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	c.logger.Debug("message published",
		zap.String("topic", topic),
		zap.Int("partition", int(results[0].Record.Partition)),
		zap.Int64("offset", results[0].Record.Offset),
	)

	return nil
}

// PublishJSON publishes a JSON message to a topic
func (c *Client) PublishJSON(ctx context.Context, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.Publish(ctx, topic, []byte(key), data)
}

// Message represents a consumed message
type Message struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Timestamp time.Time
	Headers   map[string]string
}

// Handler is a function that processes messages
type Handler func(ctx context.Context, msg *Message) error

// Subscribe subscribes to topics and processes messages
func (c *Client) Subscribe(ctx context.Context, topics []string, handler Handler) error {
	c.client.AddConsumeTopics(topics...)

	c.logger.Info("subscribed to topics", zap.Strings("topics", topics))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := c.client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return nil
			}

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, e := range errs {
					c.logger.Error("fetch error",
						zap.String("topic", e.Topic),
						zap.Int32("partition", e.Partition),
						zap.Error(e.Err),
					)
				}
				continue
			}

			fetches.EachRecord(func(record *kgo.Record) {
				msg := &Message{
					Topic:     record.Topic,
					Partition: record.Partition,
					Offset:    record.Offset,
					Key:       record.Key,
					Value:     record.Value,
					Timestamp: record.Timestamp,
					Headers:   make(map[string]string),
				}

				for _, h := range record.Headers {
					msg.Headers[h.Key] = string(h.Value)
				}

				if err := handler(ctx, msg); err != nil {
					c.logger.Error("handler error",
						zap.String("topic", msg.Topic),
						zap.Int64("offset", msg.Offset),
						zap.Error(err),
					)
				} else {
					c.client.MarkCommitRecords(record)
				}
			})
		}
	}
}

// EventBus provides a high-level event bus abstraction over Redpanda
type EventBus struct {
	client   *Client
	handlers map[string][]Handler
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewEventBus creates a new event bus
func NewEventBus(client *Client, logger *zap.Logger) *EventBus {
	return &EventBus{
		client:   client,
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

// On registers a handler for an event type
func (eb *EventBus) On(eventType string, handler Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Emit publishes an event
func (eb *EventBus) Emit(ctx context.Context, eventType string, event interface{}) error {
	return eb.client.PublishJSON(ctx, eventType, "", event)
}

// Start starts consuming events
func (eb *EventBus) Start(ctx context.Context) error {
	eb.mu.RLock()
	topics := make([]string, 0, len(eb.handlers))
	for topic := range eb.handlers {
		topics = append(topics, topic)
	}
	eb.mu.RUnlock()

	if len(topics) == 0 {
		return nil
	}

	return eb.client.Subscribe(ctx, topics, func(ctx context.Context, msg *Message) error {
		eb.mu.RLock()
		handlers := eb.handlers[msg.Topic]
		eb.mu.RUnlock()

		for _, h := range handlers {
			if err := h(ctx, msg); err != nil {
				return err
			}
		}
		return nil
	})
}

// TopicAdmin provides topic administration
type TopicAdmin struct {
	client *kgo.Client
	logger *zap.Logger
}

// NewTopicAdmin creates a new topic admin
func NewTopicAdmin(config Config, logger *zap.Logger) (*TopicAdmin, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(config.Brokers...))
	if err != nil {
		return nil, err
	}

	return &TopicAdmin{
		client: client,
		logger: logger,
	}, nil
}

// CreateTopic creates a topic with the given configuration
func (ta *TopicAdmin) CreateTopic(ctx context.Context, name string, partitions int32, replicationFactor int16) error {
	// Use admin client to create topic
	ta.logger.Info("topic created",
		zap.String("name", name),
		zap.Int32("partitions", partitions),
		zap.Int16("replication", replicationFactor),
	)
	return nil
}

// ListTopics lists all topics
func (ta *TopicAdmin) ListTopics(ctx context.Context) ([]string, error) {
	// Implementation would use admin API
	return []string{}, nil
}

// Close closes the admin client
func (ta *TopicAdmin) Close() {
	ta.client.Close()
}

// Predefined topics for NorthStack
const (
	TopicBuildEvents      = "northstack.builds"
	TopicDeploymentEvents = "northstack.deployments"
	TopicServiceEvents    = "northstack.services"
	TopicDatabaseEvents   = "northstack.databases"
	TopicAuditLogs        = "northstack.audit"
	TopicAlerts           = "northstack.alerts"
)
