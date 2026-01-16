package coolify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/internal/models"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(cfg config.CoolifyConfig) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIToken,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Application deployment request
type DeployRequest struct {
	RepoURL   string `json:"repository"`
	Branch    string `json:"branch"`
	BuildPack string `json:"build_pack"`
}

func (c *Client) TriggerBuild(ctx context.Context, appID string, commitSHA string) (string, error) {
	// Implementation placeholder matching prompt
	// POST /api/v1/applications/{id}/build
	endpoint := fmt.Sprintf("%s/api/v1/applications/%s/build?commit=%s", c.baseURL, appID, commitSHA)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response to get build ID
	return "build-id-placeholder", nil
}

func (c *Client) CreateApplication(ctx context.Context, app *models.Application) error {
	payload := map[string]interface{}{
		"name":       app.Name,
		"repository": app.Repository,
		"branch":     app.Branch,
		"build_pack": app.BuildPack,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/applications", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("coolify api error: %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *Client) GetBuildLogs(ctx context.Context, appID, buildID string) (io.ReadCloser, error) {
	endpoint := fmt.Sprintf("%s/api/v1/applications/%s/builds/%s/logs", c.baseURL, appID, buildID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	// Caller is responsible for closing body
	return resp.Body, err
}
