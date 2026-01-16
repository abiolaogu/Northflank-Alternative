package kine

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// KineConfig holds configuration for Kine with YugabyteDB backend
type KineConfig struct {
	YugabyteHost     string        `mapstructure:"yugabyte_host"`
	YugabytePort     int           `mapstructure:"yugabyte_port"`
	YugabyteUser     string        `mapstructure:"yugabyte_user"`
	YugabytePassword string        `mapstructure:"yugabyte_password"`
	YugabyteDatabase string        `mapstructure:"yugabyte_database"`
	YugabyteSSLMode  string        `mapstructure:"yugabyte_ssl_mode"`
	KineEndpoint     string        `mapstructure:"kine_endpoint"`
	KineListenPort   int           `mapstructure:"kine_listen_port"`
	CompactionPeriod time.Duration `mapstructure:"compaction_period"`
	MaxConnections   int           `mapstructure:"max_connections"`
	MinConnections   int           `mapstructure:"min_connections"`
	MaxConnLifetime  time.Duration `mapstructure:"max_conn_lifetime"`
}

// DefaultKineConfig returns sensible defaults
func DefaultKineConfig() *KineConfig {
	return &KineConfig{
		YugabyteHost:     "yugabyte-tserver",
		YugabytePort:     5433,
		YugabyteUser:     "yugabyte",
		YugabyteDatabase: "kine",
		YugabyteSSLMode:  "require",
		KineListenPort:   2379,
		CompactionPeriod: 5 * time.Minute,
		MaxConnections:   100,
		MinConnections:   10,
		MaxConnLifetime:  30 * time.Minute,
	}
}

// Client manages Kine operations
type Client struct {
	pool   *pgxpool.Pool
	config *KineConfig
	log    *zap.SugaredLogger
}

// NewClient creates a new Kine client
func NewClient(cfg *KineConfig, log *zap.SugaredLogger) (*Client, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d&pool_min_conns=%d",
		cfg.YugabyteUser,
		cfg.YugabytePassword,
		cfg.YugabyteHost,
		cfg.YugabytePort,
		cfg.YugabyteDatabase,
		cfg.YugabyteSSLMode,
		cfg.MaxConnections,
		cfg.MinConnections,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	client := &Client{
		pool:   pool,
		config: cfg,
		log:    log,
	}

	if err := client.initializeSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return client, nil
}

// initializeSchema creates the Kine tables in YugabyteDB
func (c *Client) initializeSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS kine (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		created BIGINT NOT NULL,
		deleted BIGINT,
		create_revision BIGINT NOT NULL,
		prev_revision BIGINT,
		lease BIGINT,
		value BYTEA,
		old_value BYTEA
	);

	CREATE INDEX IF NOT EXISTS kine_name_idx ON kine (name);
	CREATE INDEX IF NOT EXISTS kine_name_prev_idx ON kine (name, prev_revision);
	CREATE INDEX IF NOT EXISTS kine_id_deleted_idx ON kine (id, deleted);
	CREATE INDEX IF NOT EXISTS kine_prev_revision_idx ON kine (prev_revision);
	CREATE UNIQUE INDEX IF NOT EXISTS kine_name_create_revision_idx ON kine (name, create_revision);

	CREATE TABLE IF NOT EXISTS kine_lease (
		id BIGINT PRIMARY KEY,
		ttl BIGINT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS kine_compaction (
		id SERIAL PRIMARY KEY,
		revision BIGINT NOT NULL,
		compacted_at TIMESTAMPTZ DEFAULT NOW()
	);
	`

	_, err := c.pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	c.log.Info("Kine schema initialized successfully")
	return nil
}

// GetConnectionString returns the Kine connection string for K8s API server
func (c *Client) GetConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.config.YugabyteUser,
		c.config.YugabytePassword,
		c.config.YugabyteHost,
		c.config.YugabytePort,
		c.config.YugabyteDatabase,
		c.config.YugabyteSSLMode,
	)
}

// HealthCheck verifies the Kine backend is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	var result int
	err := c.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}

// KineStats holds Kine storage statistics
type KineStats struct {
	TotalKeys             int64
	UniqueKeys            int64
	LatestRevision        int64
	StorageSizeBytes      int64
	LastCompactedRevision int64
	LastCompactionTime    time.Time
}

// GetStats returns Kine storage statistics
func (c *Client) GetStats(ctx context.Context) (*KineStats, error) {
	stats := &KineStats{}

	err := c.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT name), COALESCE(MAX(id), 0)
		FROM kine WHERE deleted IS NULL
	`).Scan(&stats.TotalKeys, &stats.UniqueKeys, &stats.LatestRevision)
	if err != nil {
		return nil, err
	}

	err = c.pool.QueryRow(ctx, `
		SELECT pg_total_relation_size('kine')
	`).Scan(&stats.StorageSizeBytes)
	if err != nil {
		c.log.Warnw("Failed to get storage size", "error", err)
	}

	err = c.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(revision), 0), COALESCE(MAX(compacted_at), NOW())
		FROM kine_compaction
	`).Scan(&stats.LastCompactedRevision, &stats.LastCompactionTime)
	if err != nil {
		c.log.Warnw("Failed to get compaction info", "error", err)
	}

	return stats, nil
}

// Compact removes old revisions to reclaim space
func (c *Client) Compact(ctx context.Context, revision int64) error {
	c.log.Infow("Starting compaction", "target_revision", revision)

	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		DELETE FROM kine 
		WHERE id < $1 
		AND name IN (
			SELECT name FROM kine 
			WHERE id >= $1 
			AND deleted IS NULL
		)
	`, revision)
	if err != nil {
		return fmt.Errorf("failed to delete old revisions: %w", err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO kine_compaction (revision) VALUES ($1)`, revision)
	if err != nil {
		return fmt.Errorf("failed to record compaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit compaction: %w", err)
	}

	c.log.Infow("Compaction completed", "revision", revision, "deleted_rows", result.RowsAffected())
	return nil
}

// Close closes the client
func (c *Client) Close() {
	c.pool.Close()
}
