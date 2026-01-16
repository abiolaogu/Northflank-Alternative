// Package git provides GitLab adapter implementation
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitLabProvider implements GitProvider for GitLab
type GitLabProvider struct {
	config     OAuthConfig
	httpClient *http.Client
	apiBaseURL string
}

// NewGitLabProvider creates a new GitLab provider
func NewGitLabProvider(config OAuthConfig, selfHostedURL string) *GitLabProvider {
	baseURL := "https://gitlab.com/api/v4"
	if selfHostedURL != "" {
		baseURL = selfHostedURL + "/api/v4"
	}
	return &GitLabProvider{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiBaseURL: baseURL,
	}
}

// GetAuthURL returns the OAuth authorization URL
func (g *GitLabProvider) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":     {g.config.ClientID},
		"redirect_uri":  {g.config.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(g.config.Scopes, " ")},
		"state":         {state},
	}
	return "https://gitlab.com/oauth/authorize?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
func (g *GitLabProvider) ExchangeCode(ctx context.Context, code string) (*OAuthToken, error) {
	data := url.Values{
		"client_id":     {g.config.ClientID},
		"client_secret": {g.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {g.config.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://gitlab.com/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &OAuthToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		Scope:        result.Scope,
	}, nil
}

// RefreshToken refreshes an access token
func (g *GitLabProvider) RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	data := url.Values{
		"client_id":     {g.config.ClientID},
		"client_secret": {g.config.ClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://gitlab.com/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &OAuthToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}, nil
}

// ListRepositories lists user's projects
func (g *GitLabProvider) ListRepositories(ctx context.Context, token string) ([]Repository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBaseURL+"/projects?membership=true&per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projects []struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		PathWithNS    string `json:"path_with_namespace"`
		Description   string `json:"description"`
		HTTPURLToRepo string `json:"http_url_to_repo"`
		SSHURLToRepo  string `json:"ssh_url_to_repo"`
		DefaultBranch string `json:"default_branch"`
		Visibility    string `json:"visibility"`
		WebURL        string `json:"web_url"`
		CreatedAt     string `json:"created_at"`
		LastActivity  string `json:"last_activity_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	repos := make([]Repository, len(projects))
	for i, p := range projects {
		created, _ := time.Parse(time.RFC3339, p.CreatedAt)
		updated, _ := time.Parse(time.RFC3339, p.LastActivity)
		repos[i] = Repository{
			ID:            fmt.Sprintf("%d", p.ID),
			Name:          p.Name,
			FullName:      p.PathWithNS,
			Description:   p.Description,
			CloneURL:      p.HTTPURLToRepo,
			SSHURL:        p.SSHURLToRepo,
			DefaultBranch: p.DefaultBranch,
			Private:       p.Visibility == "private",
			Provider:      ProviderGitLab,
			WebURL:        p.WebURL,
			CreatedAt:     created,
			UpdatedAt:     updated,
		}
	}

	return repos, nil
}

// GetRepository gets a specific project
func (g *GitLabProvider) GetRepository(ctx context.Context, token, owner, repo string) (*Repository, error) {
	projectPath := url.PathEscape(owner + "/" + repo)
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBaseURL+"/projects/"+projectPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var project struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		PathWithNS    string `json:"path_with_namespace"`
		Description   string `json:"description"`
		HTTPURLToRepo string `json:"http_url_to_repo"`
		SSHURLToRepo  string `json:"ssh_url_to_repo"`
		DefaultBranch string `json:"default_branch"`
		Visibility    string `json:"visibility"`
		WebURL        string `json:"web_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &Repository{
		ID:            fmt.Sprintf("%d", project.ID),
		Name:          project.Name,
		FullName:      project.PathWithNS,
		Description:   project.Description,
		CloneURL:      project.HTTPURLToRepo,
		SSHURL:        project.SSHURLToRepo,
		DefaultBranch: project.DefaultBranch,
		Private:       project.Visibility == "private",
		Provider:      ProviderGitLab,
		WebURL:        project.WebURL,
	}, nil
}

// ListBranches lists project branches
func (g *GitLabProvider) ListBranches(ctx context.Context, token, owner, repo string) ([]Branch, error) {
	projectPath := url.PathEscape(owner + "/" + repo)
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBaseURL+"/projects/"+projectPath+"/repository/branches", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var glBranches []struct {
		Name   string `json:"name"`
		Commit struct {
			ID string `json:"id"`
		} `json:"commit"`
		Protected bool `json:"protected"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&glBranches); err != nil {
		return nil, err
	}

	branches := make([]Branch, len(glBranches))
	for i, b := range glBranches {
		branches[i] = Branch{
			Name:      b.Name,
			CommitSHA: b.Commit.ID,
			Protected: b.Protected,
		}
	}

	return branches, nil
}

// GetCommit gets a specific commit
func (g *GitLabProvider) GetCommit(ctx context.Context, token, owner, repo, sha string) (*Commit, error) {
	projectPath := url.PathEscape(owner + "/" + repo)
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBaseURL+"/projects/"+projectPath+"/repository/commits/"+sha, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var commit struct {
		ID          string `json:"id"`
		Message     string `json:"message"`
		AuthorName  string `json:"author_name"`
		AuthorEmail string `json:"author_email"`
		CreatedAt   string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return nil, err
	}

	ts, _ := time.Parse(time.RFC3339, commit.CreatedAt)

	return &Commit{
		SHA:       commit.ID,
		Message:   commit.Message,
		Author:    commit.AuthorName,
		Email:     commit.AuthorEmail,
		Timestamp: ts,
	}, nil
}

// ListCommits lists commits on a branch
func (g *GitLabProvider) ListCommits(ctx context.Context, token, owner, repo, branch string, limit int) ([]Commit, error) {
	projectPath := url.PathEscape(owner + "/" + repo)
	url := fmt.Sprintf("%s/projects/%s/repository/commits?ref_name=%s&per_page=%d", g.apiBaseURL, projectPath, branch, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var glCommits []struct {
		ID          string `json:"id"`
		Message     string `json:"message"`
		AuthorName  string `json:"author_name"`
		AuthorEmail string `json:"author_email"`
		CreatedAt   string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&glCommits); err != nil {
		return nil, err
	}

	commits := make([]Commit, len(glCommits))
	for i, c := range glCommits {
		ts, _ := time.Parse(time.RFC3339, c.CreatedAt)
		commits[i] = Commit{
			SHA:       c.ID,
			Message:   c.Message,
			Author:    c.AuthorName,
			Email:     c.AuthorEmail,
			Timestamp: ts,
		}
	}

	return commits, nil
}

// CreateDeployKey creates a deploy key
func (g *GitLabProvider) CreateDeployKey(ctx context.Context, token, owner, repo, title, publicKey string) (*DeployKey, error) {
	projectPath := url.PathEscape(owner + "/" + repo)
	body := fmt.Sprintf(`{"title":"%s","key":"%s","can_push":false}`, title, publicKey)

	req, err := http.NewRequestWithContext(ctx, "POST", g.apiBaseURL+"/projects/"+projectPath+"/deploy_keys", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var key struct {
		ID        int64  `json:"id"`
		Title     string `json:"title"`
		Key       string `json:"key"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, err
	}

	created, _ := time.Parse(time.RFC3339, key.CreatedAt)

	return &DeployKey{
		ID:        key.ID,
		Title:     key.Title,
		Key:       key.Key,
		ReadOnly:  true,
		CreatedAt: created,
	}, nil
}

// ListDeployKeys lists deploy keys
func (g *GitLabProvider) ListDeployKeys(ctx context.Context, token, owner, repo string) ([]DeployKey, error) {
	projectPath := url.PathEscape(owner + "/" + repo)
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBaseURL+"/projects/"+projectPath+"/deploy_keys", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keys []DeployKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// DeleteDeployKey deletes a deploy key
func (g *GitLabProvider) DeleteDeployKey(ctx context.Context, token, owner, repo string, keyID int64) error {
	projectPath := url.PathEscape(owner + "/" + repo)
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/projects/%s/deploy_keys/%d", g.apiBaseURL, projectPath, keyID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// CreateWebhook creates a project webhook
func (g *GitLabProvider) CreateWebhook(ctx context.Context, token, owner, repo string, webhook *Webhook) (*Webhook, error) {
	projectPath := url.PathEscape(owner + "/" + repo)

	// Map events to GitLab format
	payload := map[string]interface{}{
		"url":                     webhook.URL,
		"push_events":             contains(webhook.Events, "push"),
		"merge_requests_events":   contains(webhook.Events, "merge_request"),
		"token":                   webhook.Secret,
		"enable_ssl_verification": true,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", g.apiBaseURL+"/projects/"+projectPath+"/hooks", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ID        int64  `json:"id"`
		URL       string `json:"url"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	created, _ := time.Parse(time.RFC3339, result.CreatedAt)

	return &Webhook{
		ID:        result.ID,
		URL:       result.URL,
		Events:    webhook.Events,
		Active:    true,
		CreatedAt: created,
	}, nil
}

// ListWebhooks lists project webhooks
func (g *GitLabProvider) ListWebhooks(ctx context.Context, token, owner, repo string) ([]Webhook, error) {
	projectPath := url.PathEscape(owner + "/" + repo)
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBaseURL+"/projects/"+projectPath+"/hooks", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var hooks []Webhook
	if err := json.NewDecoder(resp.Body).Decode(&hooks); err != nil {
		return nil, err
	}

	return hooks, nil
}

// DeleteWebhook deletes a webhook
func (g *GitLabProvider) DeleteWebhook(ctx context.Context, token, owner, repo string, webhookID int64) error {
	projectPath := url.PathEscape(owner + "/" + repo)
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/projects/%s/hooks/%d", g.apiBaseURL, projectPath, webhookID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// ValidateWebhookPayload validates a GitLab webhook token
func (g *GitLabProvider) ValidateWebhookPayload(payload []byte, signature, secret string) bool {
	// GitLab uses X-Gitlab-Token header which is just the secret
	return signature == secret
}

// ParseWebhookEvent parses a GitLab webhook event
func (g *GitLabProvider) ParseWebhookEvent(eventType string, payload []byte) (interface{}, error) {
	switch eventType {
	case "Push Hook":
		var event struct {
			Ref     string `json:"ref"`
			Before  string `json:"before"`
			After   string `json:"after"`
			Project struct {
				PathWithNS string `json:"path_with_namespace"`
			} `json:"project"`
			Commits []struct {
				ID      string `json:"id"`
				Message string `json:"message"`
				Author  struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"author"`
				Timestamp string `json:"timestamp"`
			} `json:"commits"`
			UserUsername string `json:"user_username"`
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, err
		}

		commits := make([]Commit, len(event.Commits))
		for i, c := range event.Commits {
			ts, _ := time.Parse(time.RFC3339, c.Timestamp)
			commits[i] = Commit{
				SHA:       c.ID,
				Message:   c.Message,
				Author:    c.Author.Name,
				Email:     c.Author.Email,
				Timestamp: ts,
			}
		}

		return &PushEvent{
			Ref:        event.Ref,
			Before:     event.Before,
			After:      event.After,
			Repository: event.Project.PathWithNS,
			Commits:    commits,
			Sender:     event.UserUsername,
		}, nil

	case "Merge Request Hook":
		var event struct {
			ObjectAttributes struct {
				Action         string `json:"action"`
				IID            int    `json:"iid"`
				Title          string `json:"title"`
				SourceBranch   string `json:"source_branch"`
				TargetBranch   string `json:"target_branch"`
				LastCommitSHA  string `json:"last_commit_sha"`
				MergeCommitSHA string `json:"merge_commit_sha"`
			} `json:"object_attributes"`
			Project struct {
				PathWithNS string `json:"path_with_namespace"`
			} `json:"project"`
			User struct {
				Username string `json:"username"`
			} `json:"user"`
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, err
		}

		return &PullRequestEvent{
			Action:      event.ObjectAttributes.Action,
			Number:      event.ObjectAttributes.IID,
			Title:       event.ObjectAttributes.Title,
			HeadBranch:  event.ObjectAttributes.SourceBranch,
			HeadSHA:     event.ObjectAttributes.LastCommitSHA,
			BaseBranch:  event.ObjectAttributes.TargetBranch,
			Repository:  event.Project.PathWithNS,
			Sender:      event.User.Username,
			MergeCommit: event.ObjectAttributes.MergeCommitSHA,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
