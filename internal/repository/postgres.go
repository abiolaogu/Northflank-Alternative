// Package repository provides database access for the Platform Orchestrator.
// It implements the repository interfaces from the domain package using PostgreSQL.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/pkg/logger"
)

// PostgresDB wraps a pgxpool for database operations
type PostgresDB struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection pool
func NewPostgresDB(ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger) (*PostgresDB, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	// Configure connection settings
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Msg("Connected to PostgreSQL")

	return &PostgresDB{
		pool:   pool,
		logger: log,
	}, nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	db.pool.Close()
	db.logger.Info().Msg("PostgreSQL connection closed")
}

// Pool returns the underlying connection pool
func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}

// Exec executes a query that doesn't return rows
func (db *PostgresDB) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := db.pool.Exec(ctx, sql, args...)
	return err
}

// QueryRow executes a query that returns at most one row
func (db *PostgresDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

// Query executes a query that returns rows
func (db *PostgresDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

// BeginTx starts a new transaction
func (db *PostgresDB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// WithTx executes a function within a transaction
func (db *PostgresDB) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			db.logger.Error().Err(rbErr).Msg("Failed to rollback transaction")
		}
		return err
	}

	return tx.Commit(ctx)
}

// Migrate runs database migrations
func (db *PostgresDB) Migrate(ctx context.Context) error {
	migrations := []string{
		migrationCreateProjects,
		migrationCreateServices,
		migrationCreateBuilds,
		migrationCreateDeployments,
		migrationCreateClusters,
		migrationCreateEnvironments,
		migrationCreateSecrets,
		migrationCreateIngresses,
		migrationCreatePipelines,
		migrationCreateUsers,
		migrationCreateTeams,
		migrationCreateAuditLogs,
		migrationCreateIndexes,
	}

	for i, migration := range migrations {
		if _, err := db.pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	db.logger.Info().Int("count", len(migrations)).Msg("Database migrations completed")
	return nil
}

// SQL migrations
const migrationCreateProjects = `
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    owner_id UUID NOT NULL,
    team_id UUID,
    labels JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateServices = `
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    build_source JSONB NOT NULL DEFAULT '{}',
    resources JSONB NOT NULL DEFAULT '{}',
    scaling JSONB NOT NULL DEFAULT '{}',
    health_check JSONB,
    env_vars JSONB DEFAULT '{}',
    secret_refs JSONB DEFAULT '[]',
    ports JSONB DEFAULT '[]',
    labels JSONB DEFAULT '{}',
    annotations JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    current_build_id UUID,
    current_version VARCHAR(255),
    target_cluster_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, slug)
);
`

const migrationCreateBuilds = `
CREATE TABLE IF NOT EXISTS builds (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'queued',
    source JSONB NOT NULL DEFAULT '{}',
    image_tag VARCHAR(512),
    image_digest VARCHAR(255),
    build_logs TEXT,
    duration BIGINT,
    triggered_by VARCHAR(255) NOT NULL,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateDeployments = `
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    build_id UUID NOT NULL REFERENCES builds(id),
    cluster_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    strategy VARCHAR(50) NOT NULL DEFAULT 'rolling_update',
    version VARCHAR(255) NOT NULL,
    previous_version VARCHAR(255),
    replicas INTEGER NOT NULL DEFAULT 1,
    ready_replicas INTEGER NOT NULL DEFAULT 0,
    triggered_by VARCHAR(255) NOT NULL,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateClusters = `
CREATE TABLE IF NOT EXISTS clusters (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL,
    region VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'provisioning',
    kube_version VARCHAR(50),
    api_endpoint VARCHAR(512),
    node_count INTEGER NOT NULL DEFAULT 0,
    labels JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    rancher_cluster_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateEnvironments = `
CREATE TABLE IF NOT EXISTS environments (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    cluster_id UUID NOT NULL REFERENCES clusters(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    labels JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, slug)
);
`

const migrationCreateSecrets = `
CREATE TABLE IF NOT EXISTS secrets (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'opaque',
    keys JSONB NOT NULL DEFAULT '[]',
    vault_path VARCHAR(512) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);
`

const migrationCreateIngresses = `
CREATE TABLE IF NOT EXISTS ingresses (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    domain VARCHAR(512) NOT NULL,
    path VARCHAR(255) NOT NULL DEFAULT '/',
    type VARCHAR(50) NOT NULL DEFAULT 'http',
    tls JSONB NOT NULL DEFAULT '{"enabled": false}',
    annotations JSONB DEFAULT '{}',
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(domain, path)
);
`

const migrationCreatePipelines = `
CREATE TABLE IF NOT EXISTS pipelines (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    trigger VARCHAR(50) NOT NULL,
    branch VARCHAR(255),
    commit_sha VARCHAR(255),
    stages JSONB NOT NULL DEFAULT '[]',
    build_id UUID REFERENCES builds(id),
    deployment_id UUID REFERENCES deployments(id),
    metadata JSONB DEFAULT '{}',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateUsers = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    avatar_url VARCHAR(512),
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateTeams = `
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id),
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS team_memberships (
    id UUID PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, user_id)
);
`

const migrationCreateAuditLogs = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID NOT NULL,
    resource_name VARCHAR(255),
    project_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    old_value JSONB,
    new_value JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateIndexes = `
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_team_id ON projects(team_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_services_project_id ON services(project_id);
CREATE INDEX IF NOT EXISTS idx_services_status ON services(status);
CREATE INDEX IF NOT EXISTS idx_builds_service_id ON builds(service_id);
CREATE INDEX IF NOT EXISTS idx_builds_project_id ON builds(project_id);
CREATE INDEX IF NOT EXISTS idx_builds_status ON builds(status);
CREATE INDEX IF NOT EXISTS idx_builds_created_at ON builds(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_deployments_service_id ON deployments(service_id);
CREATE INDEX IF NOT EXISTS idx_deployments_cluster_id ON deployments(cluster_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON deployments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_clusters_provider ON clusters(provider);
CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status);
CREATE INDEX IF NOT EXISTS idx_environments_project_id ON environments(project_id);
CREATE INDEX IF NOT EXISTS idx_environments_cluster_id ON environments(cluster_id);
CREATE INDEX IF NOT EXISTS idx_secrets_project_id ON secrets(project_id);
CREATE INDEX IF NOT EXISTS idx_ingresses_service_id ON ingresses(service_id);
CREATE INDEX IF NOT EXISTS idx_ingresses_domain ON ingresses(domain);
CREATE INDEX IF NOT EXISTS idx_pipelines_service_id ON pipelines(service_id);
CREATE INDEX IF NOT EXISTS idx_pipelines_status ON pipelines(status);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_team_memberships_team_id ON team_memberships(team_id);
CREATE INDEX IF NOT EXISTS idx_team_memberships_user_id ON team_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type_id ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_project_id ON audit_logs(project_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
`
