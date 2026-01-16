// Package config provides configuration management for the unified platform.
// It supports loading configuration from files, environment variables, and Vault.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the platform orchestrator
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`    // Legacy PostgreSQL (optional)
	YugabyteDB    YugabyteDBConfig    `mapstructure:"yugabytedb"`  // Primary distributed database
	DragonflyDB   DragonflyDBConfig   `mapstructure:"dragonflydb"` // High-performance cache (Redis replacement)
	Redis         RedisConfig         `mapstructure:"redis"`       // Legacy Redis (optional fallback)
	NATS          NATSConfig          `mapstructure:"nats"`
	Integrations  IntegrationsConfig  `mapstructure:"integrations"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

// YugabyteDBConfig holds YugabyteDB distributed SQL database configuration
type YugabyteDBConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`

	// Connection pool
	MaxConns          int32         `mapstructure:"max_conns"`
	MinConns          int32         `mapstructure:"min_conns"`
	MaxConnLifetime   time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`

	// Cluster management (Master HTTP API)
	MasterHTTPAddresses []string `mapstructure:"master_http_addresses"`
	TServerHTTPAddress  string   `mapstructure:"tserver_http_address"`

	// Geo-distribution
	PlacementCloud  string   `mapstructure:"placement_cloud"`
	PlacementRegion string   `mapstructure:"placement_region"`
	PlacementZone   string   `mapstructure:"placement_zone"`
	PreferredZones  []string `mapstructure:"preferred_zones"`

	// Read replicas
	ReadReplicaEnabled bool     `mapstructure:"read_replica_enabled"`
	ReadReplicaHosts   []string `mapstructure:"read_replica_hosts"`

	// Migrations
	MigrationsPath string `mapstructure:"migrations_path"`
}

func (y YugabyteDBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		y.Username, y.Password, y.Host, y.Port, y.Database, y.SSLMode,
	)
}

// DragonflyDBConfig holds DragonflyDB (Redis-compatible) configuration
type DragonflyDBConfig struct {
	Enabled   bool     `mapstructure:"enabled"`
	Address   string   `mapstructure:"address"`   // Standalone mode
	Addresses []string `mapstructure:"addresses"` // Cluster mode
	Password  string   `mapstructure:"password"`
	DB        int      `mapstructure:"db"`

	// Connection settings
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxRetries      int           `mapstructure:"max_retries"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`

	// TLS
	TLSEnabled bool   `mapstructure:"tls_enabled"`
	TLSCert    string `mapstructure:"tls_cert"`
	TLSKey     string `mapstructure:"tls_key"`
	TLSCA      string `mapstructure:"tls_ca"`

	// DragonflyDB-specific
	ClusterMode bool   `mapstructure:"cluster_mode"`
	ReplicaRead bool   `mapstructure:"replica_read"`
	KeyPrefix   string `mapstructure:"key_prefix"`
}

func (d DragonflyDBConfig) Addr() string {
	if d.Address != "" {
		return d.Address
	}
	if len(d.Addresses) > 0 {
		return d.Addresses[0]
	}
	return "localhost:6379"
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	CORSOrigins     []string      `mapstructure:"cors_origins"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	TLSCertFile     string        `mapstructure:"tls_cert_file"`
	TLSKeyFile      string        `mapstructure:"tls_key_file"`
	CORSEnabled     bool          `mapstructure:"cors_enabled"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"` // Added for legacy
	MigrationsPath  string        `mapstructure:"migrations_path"`
	Name            string        `mapstructure:"name"` // Alias for Database
	User            string        `mapstructure:"user"` // Alias for Username
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.Username, d.Password, d.Host, d.Port, d.Database, d.SSLMode,
	)
}

type RedisConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Password    string        `mapstructure:"password"`
	Database    int           `mapstructure:"database"`
	MaxRetries  int           `mapstructure:"max_retries"`
	PoolSize    int           `mapstructure:"pool_size"`
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type NATSConfig struct {
	URL              string        `mapstructure:"url"`
	ClusterID        string        `mapstructure:"cluster_id"`
	ClientID         string        `mapstructure:"client_id"`
	Token            string        `mapstructure:"token"` // Added for compatibility
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	MaxReconnects    int           `mapstructure:"max_reconnects"`
	ReconnectWait    time.Duration `mapstructure:"reconnect_wait"`
	JetStreamEnabled bool          `mapstructure:"jetstream_enabled"`
	TLSEnabled       bool          `mapstructure:"tls_enabled"`
	TLSCertFile      string        `mapstructure:"tls_cert_file"`
	TLSKeyFile       string        `mapstructure:"tls_key_file"`
	TLSCAFile        string        `mapstructure:"tls_ca_file"`

	// Stream configuration
	Streams []StreamConfig `mapstructure:"streams"`
}

type StreamConfig struct {
	Name      string   `mapstructure:"name"`
	Subjects  []string `mapstructure:"subjects"`
	Retention string   `mapstructure:"retention"` // limits, interest, workqueue
	MaxAge    string   `mapstructure:"max_age"`
	MaxMsgs   int64    `mapstructure:"max_msgs"`
	MaxBytes  int64    `mapstructure:"max_bytes"`
	Replicas  int      `mapstructure:"replicas"`
}

type IntegrationsConfig struct {
	Coolify CoolifyConfig `mapstructure:"coolify"`
	Rancher RancherConfig `mapstructure:"rancher"`
	ArgoCD  ArgoCDConfig  `mapstructure:"argocd"`
	Vault   VaultConfig   `mapstructure:"vault"`
	RKE2    RKE2Config    `mapstructure:"rke2"`
	Hasura  HasuraConfig  `mapstructure:"hasura"`
}

// RKE2Config holds RKE2 cluster provisioning configuration
type RKE2Config struct {
	Enabled      bool   `mapstructure:"enabled"`
	RancherURL   string `mapstructure:"rancher_url"`
	RancherToken string `mapstructure:"rancher_token"`

	// Default cluster settings
	KubernetesVersion string `mapstructure:"kubernetes_version"`
	CNI               string `mapstructure:"cni"`

	// SSH for bare metal provisioning
	SSHUser    string        `mapstructure:"ssh_user"`
	SSHKeyPath string        `mapstructure:"ssh_key_path"`
	SSHTimeout time.Duration `mapstructure:"ssh_timeout"`

	// Cloud provider
	CloudProvider string `mapstructure:"cloud_provider"`

	// Security
	Profile string `mapstructure:"profile"`
	SELinux bool   `mapstructure:"selinux"`

	// Storage
	DefaultStorageClass string `mapstructure:"default_storage_class"`
}

// HasuraConfig holds Hasura GraphQL Engine configuration
type HasuraConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	AdminSecret string `mapstructure:"admin_secret"`

	// Console
	EnableConsole bool   `mapstructure:"enable_console"`
	ConsoleAssets string `mapstructure:"console_assets_dir"`

	// Database (primary)
	DatabaseURL string `mapstructure:"database_url"`

	// JWT Auth
	JWTSecret   string `mapstructure:"jwt_secret"`
	JWTIssuer   string `mapstructure:"jwt_issuer"`
	JWTAudience string `mapstructure:"jwt_audience"`

	// Performance
	QueryTimeout    time.Duration `mapstructure:"query_timeout"`
	EnableTelemetry bool          `mapstructure:"enable_telemetry"`

	// Features
	EnableRemoteSchemas     bool `mapstructure:"enable_remote_schemas"`
	EnableActions           bool `mapstructure:"enable_actions"`
	EnableEventTriggers     bool `mapstructure:"enable_event_triggers"`
	EnableScheduledTriggers bool `mapstructure:"enable_scheduled_triggers"`
}

type CoolifyConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	BaseURL       string        `mapstructure:"base_url"`
	APIToken      string        `mapstructure:"api_token"`
	WebhookSecret string        `mapstructure:"webhook_secret"`
	Timeout       time.Duration `mapstructure:"timeout"`

	// Default settings for new projects
	DefaultBuildPack string `mapstructure:"default_buildpack"`
	DefaultRegistry  string `mapstructure:"default_registry"`
	URL              string `mapstructure:"url"`     // Alias or Legacy
	APIKey           string `mapstructure:"api_key"` // Alias or Legacy
}

type RancherConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	BaseURL        string        `mapstructure:"base_url"`
	AccessKey      string        `mapstructure:"access_key"`
	SecretKey      string        `mapstructure:"secret_key"`
	ClusterID      string        `mapstructure:"cluster_id"`
	DefaultProject string        `mapstructure:"default_project"`
	Timeout        time.Duration `mapstructure:"timeout"`

	// Multi-cluster support
	Clusters      []ClusterConfig `mapstructure:"clusters"`
	TLSSkipVerify bool            `mapstructure:"tls_skip_verify"`
	URL           string          `mapstructure:"url"`
	Token         string          `mapstructure:"token"`
}

type ClusterConfig struct {
	ID      string `mapstructure:"id"`
	Name    string `mapstructure:"name"`
	Region  string `mapstructure:"region"`
	Primary bool   `mapstructure:"primary"`
}

type ArgoCDConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	ServerURL string        `mapstructure:"server_url"`
	Username  string        `mapstructure:"username"`
	Password  string        `mapstructure:"password"`
	AuthToken string        `mapstructure:"auth_token"`
	Insecure  bool          `mapstructure:"insecure"`
	GRPCWeb   bool          `mapstructure:"grpc_web"`
	Timeout   time.Duration `mapstructure:"timeout"`

	// Git repository for manifests
	ManifestRepo   string `mapstructure:"manifest_repo"`
	ManifestBranch string `mapstructure:"manifest_branch"`
	ManifestPath   string `mapstructure:"manifest_path"`

	// Application defaults
	SyncPolicy     SyncPolicyConfig `mapstructure:"sync_policy"`
	TLSSkipVerify  bool             `mapstructure:"tls_skip_verify"`
	URL            string           `mapstructure:"url"`
	Token          string           `mapstructure:"token"`
	AppProject     string           `mapstructure:"app_project"`
	RepoURL        string           `mapstructure:"repo_url"`
	TargetRevision string           `mapstructure:"target_revision"`
}

type SyncPolicyConfig struct {
	AutoSync   bool `mapstructure:"auto_sync"`
	SelfHeal   bool `mapstructure:"self_heal"`
	Prune      bool `mapstructure:"prune"`
	AllowEmpty bool `mapstructure:"allow_empty"`
	Automated  bool `mapstructure:"automated"`
}

type VaultConfig struct {
	Enabled    bool          `mapstructure:"enabled"`
	Address    string        `mapstructure:"address"`
	Token      string        `mapstructure:"token"`
	AuthMethod string        `mapstructure:"auth_method"` // token, kubernetes, approle
	MountPath  string        `mapstructure:"mount_path"`
	Timeout    time.Duration `mapstructure:"timeout"`

	// Kubernetes auth
	K8sRole     string `mapstructure:"k8s_role"`
	K8sAuthPath string `mapstructure:"k8s_auth_path"`

	// AppRole auth
	RoleID   string `mapstructure:"role_id"`
	SecretID string `mapstructure:"secret_id"`
}

type AuthConfig struct {
	JWTSecret         string        `mapstructure:"jwt_secret"`
	JWTExpiration     time.Duration `mapstructure:"jwt_expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`

	// OAuth/OIDC
	OIDCEnabled      bool     `mapstructure:"oidc_enabled"`
	OIDCIssuer       string   `mapstructure:"oidc_issuer"`
	OIDCClientID     string   `mapstructure:"oidc_client_id"`
	OIDCClientSecret string   `mapstructure:"oidc_client_secret"`
	OIDCScopes       []string `mapstructure:"oidc_scopes"`

	// API Keys
	APIKeyPrefix  string `mapstructure:"api_key_prefix"`
	APIKeyEnabled bool   `mapstructure:"api_key_enabled"`

	// Rate Limiting
	RateLimitEnabled  bool          `mapstructure:"rate_limit_enabled"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`
	RateLimitRequests int           `mapstructure:"rate_limit_requests"`

	// Session
	SessionCookieName   string        `mapstructure:"session_cookie_name"`
	SessionCookieSecure bool          `mapstructure:"session_cookie_secure"`
	SessionMaxAge       time.Duration `mapstructure:"session_max_age"`
}

type ObservabilityConfig struct {
	Metrics       MetricsConfig `mapstructure:"metrics"`
	Logging       LoggingConfig `mapstructure:"logging"`
	Tracing       TracingConfig `mapstructure:"tracing"`
	MetricsConfig MetricsConfig `mapstructure:"-"` // Alias
}

type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Path      string `mapstructure:"path"`
	Port      int    `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	Subsystem string `mapstructure:"subsystem"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, text
	Output     string `mapstructure:"output"` // stdout, file
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"` // MB
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"` // days
	Compress   bool   `mapstructure:"compress"`
}

type TracingConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	Exporter    string  `mapstructure:"exporter"` // jaeger, otlp, zipkin
	Endpoint    string  `mapstructure:"endpoint"`
	SampleRate  float64 `mapstructure:"sample_rate"`
	ServiceName string  `mapstructure:"service_name"`
}

// Load reads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/northflank-oss")
		v.AddConfigPath("$HOME/.northflank-oss")
	}

	// Environment variables
	v.SetEnvPrefix("NFOSS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("server.cors_origins", []string{"*"})

	// YugabyteDB defaults (primary database)
	v.SetDefault("yugabytedb.enabled", true)
	v.SetDefault("yugabytedb.host", "localhost")
	v.SetDefault("yugabytedb.port", 5433) // YugabyteDB YSQL port
	v.SetDefault("yugabytedb.database", "northstack")
	v.SetDefault("yugabytedb.username", "yugabyte")
	v.SetDefault("yugabytedb.ssl_mode", "disable")
	v.SetDefault("yugabytedb.max_conns", 50)
	v.SetDefault("yugabytedb.min_conns", 10)
	v.SetDefault("yugabytedb.max_conn_lifetime", "30m")
	v.SetDefault("yugabytedb.max_conn_idle_time", "5m")
	v.SetDefault("yugabytedb.health_check_period", "30s")
	v.SetDefault("yugabytedb.migrations_path", "migrations")

	// DragonflyDB defaults (Redis replacement)
	v.SetDefault("dragonflydb.enabled", true)
	v.SetDefault("dragonflydb.address", "localhost:6379")
	v.SetDefault("dragonflydb.db", 0)
	v.SetDefault("dragonflydb.pool_size", 50)
	v.SetDefault("dragonflydb.min_idle_conns", 10)
	v.SetDefault("dragonflydb.max_retries", 3)
	v.SetDefault("dragonflydb.dial_timeout", "5s")
	v.SetDefault("dragonflydb.read_timeout", "3s")
	v.SetDefault("dragonflydb.write_timeout", "3s")
	v.SetDefault("dragonflydb.pool_timeout", "4s")
	v.SetDefault("dragonflydb.conn_max_idle_time", "5m")
	v.SetDefault("dragonflydb.conn_max_lifetime", "30m")
	v.SetDefault("dragonflydb.cluster_mode", false)
	v.SetDefault("dragonflydb.replica_read", true)
	v.SetDefault("dragonflydb.key_prefix", "northstack")

	// Legacy Database defaults (fallback to PostgreSQL)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.database", "northflank_oss")
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("database.migrations_path", "migrations")

	// Legacy Redis defaults (fallback)
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.dial_timeout", "5s")

	// NATS defaults
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.cluster_id", "northflank-cluster")
	v.SetDefault("nats.client_id", "platform-orchestrator")
	v.SetDefault("nats.max_reconnects", 60)
	v.SetDefault("nats.reconnect_wait", "2s")
	v.SetDefault("nats.jetstream_enabled", true)

	// Integration defaults - Coolify
	v.SetDefault("integrations.coolify.enabled", true)
	v.SetDefault("integrations.coolify.timeout", "30s")
	v.SetDefault("integrations.coolify.default_buildpack", "nixpacks")

	// Integration defaults - Rancher
	v.SetDefault("integrations.rancher.enabled", true)
	v.SetDefault("integrations.rancher.timeout", "30s")

	// Integration defaults - ArgoCD
	v.SetDefault("integrations.argocd.enabled", true)
	v.SetDefault("integrations.argocd.timeout", "30s")
	v.SetDefault("integrations.argocd.manifest_branch", "main")
	v.SetDefault("integrations.argocd.manifest_path", "deployments")
	v.SetDefault("integrations.argocd.sync_policy.auto_sync", true)
	v.SetDefault("integrations.argocd.sync_policy.self_heal", true)
	v.SetDefault("integrations.argocd.sync_policy.prune", true)

	// Integration defaults - Vault
	v.SetDefault("integrations.vault.enabled", true)
	v.SetDefault("integrations.vault.auth_method", "kubernetes")
	v.SetDefault("integrations.vault.mount_path", "secret")
	v.SetDefault("integrations.vault.timeout", "10s")

	// Integration defaults - RKE2
	v.SetDefault("integrations.rke2.enabled", true)
	v.SetDefault("integrations.rke2.kubernetes_version", "v1.28.5+rke2r1")
	v.SetDefault("integrations.rke2.cni", "cilium")
	v.SetDefault("integrations.rke2.ssh_user", "root")
	v.SetDefault("integrations.rke2.ssh_timeout", "30s")
	v.SetDefault("integrations.rke2.cloud_provider", "none")
	v.SetDefault("integrations.rke2.profile", "cis-1.23")
	v.SetDefault("integrations.rke2.selinux", false)
	v.SetDefault("integrations.rke2.default_storage_class", "longhorn")

	// Integration defaults - Hasura
	v.SetDefault("integrations.hasura.enabled", true)
	v.SetDefault("integrations.hasura.endpoint", "http://localhost:8081")
	v.SetDefault("integrations.hasura.enable_console", true)
	v.SetDefault("integrations.hasura.query_timeout", "60s")
	v.SetDefault("integrations.hasura.enable_telemetry", false)
	v.SetDefault("integrations.hasura.enable_remote_schemas", true)
	v.SetDefault("integrations.hasura.enable_actions", true)
	v.SetDefault("integrations.hasura.enable_event_triggers", true)
	v.SetDefault("integrations.hasura.enable_scheduled_triggers", true)

	// Auth defaults
	v.SetDefault("auth.jwt_expiration", "24h")
	v.SetDefault("auth.refresh_expiration", "168h")
	v.SetDefault("auth.api_key_prefix", "nfoss_")
	v.SetDefault("auth.session_cookie_name", "nfoss_session")
	v.SetDefault("auth.session_cookie_secure", true)
	v.SetDefault("auth.session_max_age", "168h")

	// Observability defaults
	v.SetDefault("observability.metrics.enabled", true)
	v.SetDefault("observability.metrics.path", "/metrics")
	v.SetDefault("observability.metrics.port", 9090)
	v.SetDefault("observability.metrics.namespace", "northflank_oss")
	v.SetDefault("observability.metrics.subsystem", "orchestrator")

	v.SetDefault("observability.logging.level", "info")
	v.SetDefault("observability.logging.format", "json")
	v.SetDefault("observability.logging.output", "stdout")

	v.SetDefault("observability.tracing.enabled", false)
	v.SetDefault("observability.tracing.exporter", "otlp")
	v.SetDefault("observability.tracing.sample_rate", 0.1)
	v.SetDefault("observability.tracing.service_name", "northflank-oss-orchestrator")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Integrations.Coolify.Enabled && c.Integrations.Coolify.BaseURL == "" {
		return fmt.Errorf("coolify base_url is required when coolify is enabled")
	}

	if c.Integrations.Rancher.Enabled && c.Integrations.Rancher.BaseURL == "" {
		return fmt.Errorf("rancher base_url is required when rancher is enabled")
	}

	if c.Integrations.ArgoCD.Enabled && c.Integrations.ArgoCD.ServerURL == "" {
		return fmt.Errorf("argocd server_url is required when argocd is enabled")
	}

	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}

	return nil
}
