// Package git provides GitHub adapter implementation
package git

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubProvider implements GitProvider for GitHub
type GitHubProvider struct {
	config     OAuthConfig
	httpClient *http.Client
	apiBaseURL string
}

// NewGitHubProvider creates a new GitHub provider
func NewGitHubProvider(config OAuthConfig) *GitHubProvider {
	return &GitHubProvider{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiBaseURL: "https://api.github.com",
	}
}

// GetAuthURL returns the OAuth authorization URL
func (g *GitHubProvider) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":    {g.config.ClientID},
		"redirect_uri": {g.config.RedirectURL},
		"scope":        {strings.Join(g.config.Scopes, " ")},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
func (g *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*OAuthToken, error) {
	data := url.Values{
		"client_id":     {g.config.ClientID},
		"client_secret": {g.config.ClientSecret},
		"code":          {code},
		"redirect_uri":  {g.config.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("github oauth error: %s", result.Error)
	}

	return &OAuthToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		Scope:        result.Scope,
	}, nil
}

// RefreshToken refreshes an access token
func (g *GitHubProvider) RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	// GitHub doesn't use refresh tokens by default
	return nil, fmt.Errorf("github does not support token refresh")
}

// ListRepositories lists user's repositories
func (g *GitHubProvider) ListRepositories(ctx context.Context, token string) ([]Repository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBaseURL+"/user/repos?per_page=100&sort=updated", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var ghRepos []struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		CloneURL      string `json:"clone_url"`
		SSHURL        string `json:"ssh_url"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
		HTMLURL       string `json:"html_url"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghRepos); err != nil {
		return nil, err
	}

	repos := make([]Repository, len(ghRepos))
	for i, r := range ghRepos {
		createdAt, _ := time.Parse(time.RFC3339, r.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, r.UpdatedAt)
		repos[i] = Repository{
			ID:            fmt.Sprintf("%d", r.ID),
			Name:          r.Name,
			FullName:      r.FullName,
			Description:   r.Description,
			CloneURL:      r.CloneURL,
			SSHURL:        r.SSHURL,
			DefaultBranch: r.DefaultBranch,
			Private:       r.Private,
			Provider:      ProviderGitHub,
			WebURL:        r.HTMLURL,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		}
	}

	return repos, nil
}

// GetRepository gets a specific repository
func (g *GitHubProvider) GetRepository(ctx context.Context, token, owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", g.apiBaseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ghRepo struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		CloneURL      string `json:"clone_url"`
		SSHURL        string `json:"ssh_url"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
		HTMLURL       string `json:"html_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghRepo); err != nil {
		return nil, err
	}

	return &Repository{
		ID:            fmt.Sprintf("%d", ghRepo.ID),
		Name:          ghRepo.Name,
		FullName:      ghRepo.FullName,
		Description:   ghRepo.Description,
		CloneURL:      ghRepo.CloneURL,
		SSHURL:        ghRepo.SSHURL,
		DefaultBranch: ghRepo.DefaultBranch,
		Private:       ghRepo.Private,
		Provider:      ProviderGitHub,
		WebURL:        ghRepo.HTMLURL,
	}, nil
}

// ListBranches lists repository branches
func (g *GitHubProvider) ListBranches(ctx context.Context, token, owner, repo string) ([]Branch, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/branches", g.apiBaseURL, owner, repo)
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

	var ghBranches []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
		Protected bool `json:"protected"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghBranches); err != nil {
		return nil, err
	}

	branches := make([]Branch, len(ghBranches))
	for i, b := range ghBranches {
		branches[i] = Branch{
			Name:      b.Name,
			CommitSHA: b.Commit.SHA,
			Protected: b.Protected,
		}
	}

	return branches, nil
}

// GetCommit gets a specific commit
func (g *GitHubProvider) GetCommit(ctx context.Context, token, owner, repo, sha string) (*Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s", g.apiBaseURL, owner, repo, sha)
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

	var ghCommit struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name  string `json:"name"`
				Email string `json:"email"`
				Date  string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghCommit); err != nil {
		return nil, err
	}

	timestamp, _ := time.Parse(time.RFC3339, ghCommit.Commit.Author.Date)

	return &Commit{
		SHA:       ghCommit.SHA,
		Message:   ghCommit.Commit.Message,
		Author:    ghCommit.Commit.Author.Name,
		Email:     ghCommit.Commit.Author.Email,
		Timestamp: timestamp,
	}, nil
}

// ListCommits lists commits on a branch
func (g *GitHubProvider) ListCommits(ctx context.Context, token, owner, repo, branch string, limit int) ([]Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?sha=%s&per_page=%d", g.apiBaseURL, owner, repo, branch, limit)
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

	var ghCommits []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name  string `json:"name"`
				Email string `json:"email"`
				Date  string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghCommits); err != nil {
		return nil, err
	}

	commits := make([]Commit, len(ghCommits))
	for i, c := range ghCommits {
		timestamp, _ := time.Parse(time.RFC3339, c.Commit.Author.Date)
		commits[i] = Commit{
			SHA:       c.SHA,
			Message:   c.Commit.Message,
			Author:    c.Commit.Author.Name,
			Email:     c.Commit.Author.Email,
			Timestamp: timestamp,
		}
	}

	return commits, nil
}

// CreateDeployKey creates a deploy key for a repository
func (g *GitHubProvider) CreateDeployKey(ctx context.Context, token, owner, repo, title, publicKey string) (*DeployKey, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/keys", g.apiBaseURL, owner, repo)
	body := fmt.Sprintf(`{"title":"%s","key":"%s","read_only":true}`, title, publicKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
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
		ReadOnly  bool   `json:"read_only"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339, key.CreatedAt)

	return &DeployKey{
		ID:        key.ID,
		Title:     key.Title,
		Key:       key.Key,
		ReadOnly:  key.ReadOnly,
		CreatedAt: createdAt,
	}, nil
}

// ListDeployKeys lists deploy keys
func (g *GitHubProvider) ListDeployKeys(ctx context.Context, token, owner, repo string) ([]DeployKey, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/keys", g.apiBaseURL, owner, repo)
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

	var keys []DeployKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// DeleteDeployKey deletes a deploy key
func (g *GitHubProvider) DeleteDeployKey(ctx context.Context, token, owner, repo string, keyID int64) error {
	url := fmt.Sprintf("%s/repos/%s/%s/keys/%d", g.apiBaseURL, owner, repo, keyID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete deploy key: %d", resp.StatusCode)
	}

	return nil
}

// CreateWebhook creates a repository webhook
func (g *GitHubProvider) CreateWebhook(ctx context.Context, token, owner, repo string, webhook *Webhook) (*Webhook, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks", g.apiBaseURL, owner, repo)

	payload := map[string]interface{}{
		"name":   "web",
		"active": webhook.Active,
		"events": webhook.Events,
		"config": map[string]interface{}{
			"url":          webhook.URL,
			"content_type": "json",
			"secret":       webhook.Secret,
			"insecure_ssl": "0",
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
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
		ID     int64    `json:"id"`
		Events []string `json:"events"`
		Active bool     `json:"active"`
		Config struct {
			URL string `json:"url"`
		} `json:"config"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339, result.CreatedAt)

	return &Webhook{
		ID:        result.ID,
		URL:       result.Config.URL,
		Events:    result.Events,
		Active:    result.Active,
		CreatedAt: createdAt,
	}, nil
}

// ListWebhooks lists repository webhooks
func (g *GitHubProvider) ListWebhooks(ctx context.Context, token, owner, repo string) ([]Webhook, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks", g.apiBaseURL, owner, repo)
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

	var hooks []Webhook
	if err := json.NewDecoder(resp.Body).Decode(&hooks); err != nil {
		return nil, err
	}

	return hooks, nil
}

// DeleteWebhook deletes a webhook
func (g *GitHubProvider) DeleteWebhook(ctx context.Context, token, owner, repo string, webhookID int64) error {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d", g.apiBaseURL, owner, repo, webhookID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ValidateWebhookPayload validates a GitHub webhook signature
func (g *GitHubProvider) ValidateWebhookPayload(payload []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature[7:]), []byte(expectedMAC))
}

// ParseWebhookEvent parses a GitHub webhook event
func (g *GitHubProvider) ParseWebhookEvent(eventType string, payload []byte) (interface{}, error) {
	switch eventType {
	case "push":
		var event struct {
			Ref        string `json:"ref"`
			Before     string `json:"before"`
			After      string `json:"after"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			Commits []struct {
				ID      string `json:"id"`
				Message string `json:"message"`
				Author  struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"author"`
				Timestamp string `json:"timestamp"`
			} `json:"commits"`
			Sender struct {
				Login string `json:"login"`
			} `json:"sender"`
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
			Repository: event.Repository.FullName,
			Commits:    commits,
			Sender:     event.Sender.Login,
		}, nil

	case "pull_request":
		var event struct {
			Action      string `json:"action"`
			Number      int    `json:"number"`
			PullRequest struct {
				Title string `json:"title"`
				Head  struct {
					Ref string `json:"ref"`
					SHA string `json:"sha"`
				} `json:"head"`
				Base struct {
					Ref string `json:"ref"`
				} `json:"base"`
				MergeCommitSHA string `json:"merge_commit_sha"`
			} `json:"pull_request"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			Sender struct {
				Login string `json:"login"`
			} `json:"sender"`
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, err
		}

		return &PullRequestEvent{
			Action:      event.Action,
			Number:      event.Number,
			Title:       event.PullRequest.Title,
			HeadBranch:  event.PullRequest.Head.Ref,
			HeadSHA:     event.PullRequest.Head.SHA,
			BaseBranch:  event.PullRequest.Base.Ref,
			Repository:  event.Repository.FullName,
			Sender:      event.Sender.Login,
			MergeCommit: event.PullRequest.MergeCommitSHA,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}
