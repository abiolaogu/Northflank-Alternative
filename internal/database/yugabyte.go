package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/northstack/platform/internal/config"
)

type YugabyteDB struct {
	pool   *pgxpool.Pool
	config config.YugabyteDBConfig
}

type ClusterInfo struct {
	Nodes   int
	Zones   []string
	Regions []string
}

func NewYugabyteDB(cfg config.YugabyteDBConfig) (*YugabyteDB, error) {
	dsn := cfg.DSN()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply connection pool settings
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Multi-host / HA logic could be enhanced here by parsing cfg.Host if it contains multiple IPs
	// For standard pgx, it supports connection strings with multiple hosts

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &YugabyteDB{
		pool:   pool,
		config: cfg,
	}, nil
}

func (y *YugabyteDB) Pool() *pgxpool.Pool {
	return y.pool
}

func (y *YugabyteDB) Health(ctx context.Context) error {
	return y.pool.Ping(ctx)
}

func (y *YugabyteDB) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	// Query system catalogs to get node info
	// This is a placeholder implementation
	var count int
	err := y.pool.QueryRow(ctx, "SELECT count(*) FROM yb_servers()").Scan(&count)
	if err != nil {
		// Fallback for standard Postgres or if yb_servers doesn't exist yet
		return &ClusterInfo{Nodes: 1}, nil
	}

	return &ClusterInfo{
		Nodes: count,
	}, nil
}

func (y *YugabyteDB) ExecuteInTransaction(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	tx, err := y.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (y *YugabyteDB) Close() {
	y.pool.Close()
}
