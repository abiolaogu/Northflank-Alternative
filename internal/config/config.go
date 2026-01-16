// Package config provides configuration management for the Platform Orchestrator.
// It supports loading configuration from files, environment variables, and defaults.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the platform orchestrator
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	NATS       NATSConfig       `mapstructure:"nats"`
	Coolify    CoolifyConfig    `mapstructure:"coolify"`
	Rancher    RancherConfig    `mapstructure:"rancher"`
	ArgoCD     ArgoCDConfig     `mapstructure:"argocd"`
	Vault      VaultConfig      `mapstructure:"vault"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	TLSCertFile     string        `mapstructure:"tls_cert_file"`
	TLSKeyFile      string        `mapstructure:"tls_key_file"`
	CORSEnabled     bool          `mapstructure:"cors_enabled"`
	CORSOrigins     []string      `mapstructure:"cors_origins"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Name, c.User, c.Password, c.SSLMode,
	)
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL             string        `mapstructure:"url"`
	ClusterID       string        `mapstructure:"cluster_id"`
	ClientID        string        `mapstructure:"client_id"`
	Token           string        `mapstructure:"token"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	TLSCertFile     string        `mapstructure:"tls_cert_file"`
	TLSKeyFile      string        `mapstructure:"tls_key_file"`
	TLSCAFile       string        `mapstructure:"tls_ca_file"`
	ReconnectWait   time.Duration `mapstructure:"reconnect_wait"`
	MaxReconnects   int           `mapstructure:"max_reconnects"`
	JetStreamEnabled bool         `mapstructure:"jetstream_enabled"`
}

// CoolifyConfig holds Coolify CI/Build integration configuration
type CoolifyConfig struct {
	URL         string        `mapstructure:"url"`
	APIKey      string        `mapstructure:"api_key"`
	Timeout     time.Duration `mapstructure:"timeout"`
	WebhookURL  string        `mapstructure:"webhook_url"`
	RegistryURL string        `mapstructure:"registry_url"`
}

// RancherConfig holds Rancher cluster management configuration
type RancherConfig struct {
	URL             string        `mapstructure:"url"`
	AccessKey       string        `mapstructure:"access_key"`
	SecretKey       string        `mapstructure:"secret_key"`
	Token           string        `mapstructure:"token"`
	Timeout         time.Duration `mapstructure:"timeout"`
	TLSSkipVerify   bool          `mapstructure:"tls_skip_verify"`
	DefaultProject  string        `mapstructure:"default_project"`
}

// ArgoCDConfig holds ArgoCD GitOps configuration
type ArgoCDConfig struct {
	URL              string        `mapstructure:"url"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	Token            string        `mapstructure:"token"`
	Timeout          time.Duration `mapstructure:"timeout"`
	TLSSkipVerify    bool          `mapstructure:"tls_skip_verify"`
	AppProject       string        `mapstructure:"app_project"`
	RepoURL          string        `mapstructure:"repo_url"`
	TargetRevision   string        `mapstructure:"target_revision"`
	SyncPolicy       SyncPolicy    `mapstructure:"sync_policy"`
}

// SyncPolicy defines ArgoCD sync policy
type SyncPolicy struct {
	Automated      bool `mapstructure:"automated"`
	Prune          bool `mapstructure:"prune"`
	SelfHeal       bool `mapstructure:"self_heal"`
	AllowEmpty     bool `mapstructure:"allow_empty"`
}

// VaultConfig holds HashiCorp Vault configuration
type VaultConfig struct {
	Address       string        `mapstructure:"address"`
	Token         string        `mapstructure:"token"`
	RoleID        string        `mapstructure:"role_id"`
	SecretID      string        `mapstructure:"secret_id"`
	AuthMethod    string        `mapstructure:"auth_method"` // token, approle, kubernetes
	MountPath     string        `mapstructure:"mount_path"`
	Namespace     string        `mapstructure:"namespace"`
	TLSEnabled    bool          `mapstructure:"tls_enabled"`
	TLSSkipVerify bool          `mapstructure:"tls_skip_verify"`
	TLSCAFile     string        `mapstructure:"tls_ca_file"`
	Timeout       time.Duration `mapstructure:"timeout"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret"`
	JWTExpiration      time.Duration `mapstructure:"jwt_expiration"`
	RefreshExpiration  time.Duration `mapstructure:"refresh_expiration"`
	SessionTimeout     time.Duration `mapstructure:"session_timeout"`
	BCryptCost         int           `mapstructure:"bcrypt_cost"`
	OAuthEnabled       bool          `mapstructure:"oauth_enabled"`
	OAuthProviders     []OAuthProvider `mapstructure:"oauth_providers"`
	APIKeyEnabled      bool          `mapstructure:"api_key_enabled"`
	RateLimitEnabled   bool          `mapstructure:"rate_limit_enabled"`
	RateLimitRequests  int           `mapstructure:"rate_limit_requests"`
	RateLimitWindow    time.Duration `mapstructure:"rate_limit_window"`
}

// OAuthProvider defines an OAuth provider configuration
type OAuthProvider struct {
	Name         string   `mapstructure:"name"`
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
	AuthURL      string   `mapstructure:"auth_url"`
	TokenURL     string   `mapstructure:"token_url"`
	UserInfoURL  string   `mapstructure:"user_info_url"`
}

// MetricsConfig holds metrics and monitoring configuration
type MetricsConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Path           string `mapstructure:"path"`
	PrometheusURL  string `mapstructure:"prometheus_url"`
	GrafanaURL     string `mapstructure:"grafana_url"`
	AlertmanagerURL string `mapstructure:"alertmanager_url"`
	LokiURL        string `mapstructure:"loki_url"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, console
	Output     string `mapstructure:"output"` // stdout, stderr, file
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`    // megabytes
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`     // days
	Compress   bool   `mapstructure:"compress"`
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	InCluster      bool   `mapstructure:"in_cluster"`
	KubeConfigPath string `mapstructure:"kubeconfig_path"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Read from environment variables
	v.SetEnvPrefix("OPENPAAS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.shutdown_timeout", 10*time.Second)
	v.SetDefault("server.cors_enabled", true)
	v.SetDefault("server.cors_origins", []string{"*"})

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "openpaas")
	v.SetDefault("database.user", "openpaas")
	v.SetDefault("database.password", "")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 15*time.Minute)
	v.SetDefault("database.conn_max_idle_time", 5*time.Minute)

	// NATS defaults
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.cluster_id", "openpaas-cluster")
	v.SetDefault("nats.client_id", "platform-orchestrator")
	v.SetDefault("nats.reconnect_wait", 2*time.Second)
	v.SetDefault("nats.max_reconnects", -1)
	v.SetDefault("nats.jetstream_enabled", true)

	// Coolify defaults
	v.SetDefault("coolify.url", "http://localhost:3000")
	v.SetDefault("coolify.timeout", 60*time.Second)

	// Rancher defaults
	v.SetDefault("rancher.url", "https://localhost:8443")
	v.SetDefault("rancher.timeout", 60*time.Second)
	v.SetDefault("rancher.tls_skip_verify", false)

	// ArgoCD defaults
	v.SetDefault("argocd.url", "https://localhost:8080")
	v.SetDefault("argocd.timeout", 60*time.Second)
	v.SetDefault("argocd.tls_skip_verify", false)
	v.SetDefault("argocd.app_project", "default")
	v.SetDefault("argocd.target_revision", "HEAD")
	v.SetDefault("argocd.sync_policy.automated", true)
	v.SetDefault("argocd.sync_policy.prune", true)
	v.SetDefault("argocd.sync_policy.self_heal", true)

	// Vault defaults
	v.SetDefault("vault.address", "http://localhost:8200")
	v.SetDefault("vault.auth_method", "token")
	v.SetDefault("vault.mount_path", "secret")
	v.SetDefault("vault.timeout", 30*time.Second)

	// Auth defaults
	v.SetDefault("auth.jwt_expiration", 24*time.Hour)
	v.SetDefault("auth.refresh_expiration", 7*24*time.Hour)
	v.SetDefault("auth.session_timeout", 30*time.Minute)
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.api_key_enabled", true)
	v.SetDefault("auth.rate_limit_enabled", true)
	v.SetDefault("auth.rate_limit_requests", 100)
	v.SetDefault("auth.rate_limit_window", time.Minute)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.max_size", 100)
	v.SetDefault("logging.max_backups", 3)
	v.SetDefault("logging.max_age", 7)
	v.SetDefault("logging.compress", true)

	// Kubernetes defaults
	v.SetDefault("kubernetes.in_cluster", false)
	v.SetDefault("kubernetes.default_timeout", 30*time.Second)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Auth.JWTSecret == "" && c.Auth.JWTExpiration > 0 {
		return fmt.Errorf("jwt_secret is required when JWT is enabled")
	}

	if c.Auth.BCryptCost < 4 || c.Auth.BCryptCost > 31 {
		return fmt.Errorf("bcrypt_cost must be between 4 and 31")
	}

	return nil
}

// GetAddress returns the server address in host:port format
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
