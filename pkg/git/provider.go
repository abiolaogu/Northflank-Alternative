// Package git provides adapters for Git providers (GitHub, GitLab, Bitbucket)
// This implements integration similar to Coolify and Northflank.
package git

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Provider represents a Git provider type
type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderGitLab    Provider = "gitlab"
	ProviderBitbucket Provider = "bitbucket"
	ProviderGitea     Provider = "gitea"
)

// Repository represents a Git repository
type Repository struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	DefaultBranch string    `json:"default_branch"`
	Private       bool      `json:"private"`
	Provider      Provider  `json:"provider"`
	WebURL        string    `json:"web_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Branch represents a Git branch
type Branch struct {
	Name      string `json:"name"`
	CommitSHA string `json:"commit_sha"`
	Protected bool   `json:"protected"`
}

// Commit represents a Git commit
type Commit struct {
	SHA       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

// DeployKey represents a deploy key for repository access
type DeployKey struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Key       string    `json:"key"`
	ReadOnly  bool      `json:"read_only"`
	CreatedAt time.Time `json:"created_at"`
}

// Webhook represents a repository webhook
type Webhook struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Active    bool      `json:"active"`
	Secret    string    `json:"secret,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// GitProvider interface for all Git provider implementations
type GitProvider interface {
	// Authentication
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error)

	// Repositories
	ListRepositories(ctx context.Context, token string) ([]Repository, error)
	GetRepository(ctx context.Context, token, owner, repo string) (*Repository, error)
	ListBranches(ctx context.Context, token, owner, repo string) ([]Branch, error)
	GetCommit(ctx context.Context, token, owner, repo, sha string) (*Commit, error)
	ListCommits(ctx context.Context, token, owner, repo, branch string, limit int) ([]Commit, error)

	// Deploy Keys (for private repos)
	CreateDeployKey(ctx context.Context, token, owner, repo, title, publicKey string) (*DeployKey, error)
	ListDeployKeys(ctx context.Context, token, owner, repo string) ([]DeployKey, error)
	DeleteDeployKey(ctx context.Context, token, owner, repo string, keyID int64) error

	// Webhooks
	CreateWebhook(ctx context.Context, token, owner, repo string, webhook *Webhook) (*Webhook, error)
	ListWebhooks(ctx context.Context, token, owner, repo string) ([]Webhook, error)
	DeleteWebhook(ctx context.Context, token, owner, repo string, webhookID int64) error

	// Validation
	ValidateWebhookPayload(payload []byte, signature, secret string) bool
	ParseWebhookEvent(eventType string, payload []byte) (interface{}, error)
}

// OAuthToken represents OAuth tokens
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
}

// OAuthConfig holds OAuth configuration for a provider
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// WebhookEvent types for all providers
type PushEvent struct {
	Ref        string   `json:"ref"`
	Before     string   `json:"before"`
	After      string   `json:"after"`
	Repository string   `json:"repository"`
	Commits    []Commit `json:"commits"`
	Sender     string   `json:"sender"`
}

type PullRequestEvent struct {
	Action      string `json:"action"` // opened, closed, merged, synchronize
	Number      int    `json:"number"`
	Title       string `json:"title"`
	HeadBranch  string `json:"head_branch"`
	HeadSHA     string `json:"head_sha"`
	BaseBranch  string `json:"base_branch"`
	Repository  string `json:"repository"`
	Sender      string `json:"sender"`
	MergeCommit string `json:"merge_commit,omitempty"`
}

// ProviderRegistry manages Git provider instances
type ProviderRegistry struct {
	providers map[Provider]GitProvider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[Provider]GitProvider),
	}
}

// Register adds a provider to the registry
func (r *ProviderRegistry) Register(provider Provider, impl GitProvider) {
	r.providers[provider] = impl
}

// Get returns a provider by type
func (r *ProviderRegistry) Get(provider Provider) (GitProvider, error) {
	impl, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
	return impl, nil
}

// GitConnection represents a saved connection to a Git provider
type GitConnection struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Provider     Provider  `json:"provider"`
	ProviderID   string    `json:"provider_id"`
	Username     string    `json:"username"`
	AccessToken  string    `json:"-"` // Encrypted in storage
	RefreshToken string    `json:"-"` // Encrypted in storage
	Scopes       []string  `json:"scopes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GitConnectionRepository interface for storing connections
type GitConnectionRepository interface {
	Create(ctx context.Context, connection *GitConnection) error
	GetByID(ctx context.Context, id uuid.UUID) (*GitConnection, error)
	GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider Provider) (*GitConnection, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]GitConnection, error)
	Update(ctx context.Context, connection *GitConnection) error
	Delete(ctx context.Context, id uuid.UUID) error
}
