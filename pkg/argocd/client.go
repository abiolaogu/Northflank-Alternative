package argocd

import (
	"context"
	"fmt"

	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/internal/models"
)

type Client struct {
	serverURL string
	token     string
	insecure  bool
}

func NewClient(cfg config.ArgoCDConfig) *Client {
	return &Client{
		serverURL: cfg.ServerURL,
		token:     cfg.AuthToken,
		insecure:  cfg.Insecure,
	}
}

func (c *Client) Sync(ctx context.Context, appName string) error {
	// gRPC call to ArgoCD would go here
	fmt.Printf("Syncing app %s\n", appName)
	return nil
}

func (c *Client) CreateApplication(ctx context.Context, app *models.Application) error {
	// gRPC create application
	return nil
}

func (c *Client) GetSyncStatus(ctx context.Context, appName string) (string, error) {
	return "Synced", nil
}

func (c *Client) Health(ctx context.Context) error {
	return nil
}
