package rancher

import (
	"context"
	"net/http"

	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/internal/models"
)

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewClient(cfg config.RancherConfig) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		token:   cfg.AccessKey + ":" + cfg.SecretKey, // Basic auth simplified
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *Client) ListClusters(ctx context.Context) ([]*models.Cluster, error) {
	// GET /v3/clusters
	// Placeholder implementation
	return []*models.Cluster{}, nil
}

func (c *Client) CreateCluster(ctx context.Context, name string, provider string) (*models.Cluster, error) {
	return &models.Cluster{
		Name:     name,
		Provider: provider,
		Status:   "provisioning",
	}, nil
}

func (c *Client) GetKubeconfig(ctx context.Context, clusterID string) ([]byte, error) {
	return []byte("apiVersion: v1..."), nil
}

func (c *Client) Health(ctx context.Context) error {
	// Ping Rancher
	return nil
}
