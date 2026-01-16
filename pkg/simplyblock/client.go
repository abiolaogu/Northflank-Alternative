package simplyblock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Client provides integration with Simplyblock NVMe-oF storage
type Client struct {
	client    *resty.Client
	baseURL   string
	apiKey    string
	clusterID string
	log       *zap.SugaredLogger
}

// Config holds Simplyblock client configuration
type Config struct {
	BaseURL   string        `mapstructure:"base_url"`
	APIKey    string        `mapstructure:"api_key"`
	ClusterID string        `mapstructure:"cluster_id"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// NewClient creates a new Simplyblock client
func NewClient(cfg *Config, log *zap.SugaredLogger) (*Client, error) {
	client := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetHeader("Authorization", "Bearer "+cfg.APIKey).
		SetHeader("Content-Type", "application/json")

	return &Client{
		client:    client,
		baseURL:   cfg.BaseURL,
		apiKey:    cfg.APIKey,
		clusterID: cfg.ClusterID,
		log:       log,
	}, nil
}

// Volume represents a Simplyblock volume
type Volume struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	SizeGB        int               `json:"size_gb"`
	IOPS          int               `json:"iops"`
	ThroughputMB  int               `json:"throughput_mb"`
	Status        string            `json:"status"`
	StoragePool   string            `json:"storage_pool"`
	NVMeSubsystem string            `json:"nvme_subsystem"`
	NQN           string            `json:"nqn"`
	Namespace     int               `json:"namespace"`
	Labels        map[string]string `json:"labels"`
	CreatedAt     time.Time         `json:"created_at"`
}

// CreateVolumeInput holds parameters for creating a volume
type CreateVolumeInput struct {
	Name         string            `json:"name"`
	SizeGB       int               `json:"size_gb"`
	StoragePool  string            `json:"storage_pool"`
	IOPS         int               `json:"iops,omitempty"`
	ThroughputMB int               `json:"throughput_mb,omitempty"`
	Encryption   bool              `json:"encryption,omitempty"`
	Replication  int               `json:"replication,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// CreateVolume creates a new NVMe-oF volume
func (c *Client) CreateVolume(ctx context.Context, input *CreateVolumeInput) (*Volume, error) {
	var volume Volume
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(input).
		SetResult(&volume).
		Post(fmt.Sprintf("/api/v1/clusters/%s/volumes", c.clusterID))

	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create volume: %s", resp.String())
	}

	c.log.Infow("Volume created", "id", volume.ID, "name", volume.Name, "size_gb", volume.SizeGB)
	return &volume, nil
}

// DeleteVolume deletes a volume
func (c *Client) DeleteVolume(ctx context.Context, volumeID string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/v1/clusters/%s/volumes/%s", c.clusterID, volumeID))

	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete volume: %s", resp.String())
	}

	c.log.Infow("Volume deleted", "id", volumeID)
	return nil
}

// GetVolume retrieves a volume by ID
func (c *Client) GetVolume(ctx context.Context, volumeID string) (*Volume, error) {
	var volume Volume
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&volume).
		Get(fmt.Sprintf("/api/v1/clusters/%s/volumes/%s", c.clusterID, volumeID))

	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get volume: %s", resp.String())
	}

	return &volume, nil
}

// ResizeVolume resizes a volume (online)
func (c *Client) ResizeVolume(ctx context.Context, volumeID string, newSizeGB int) (*Volume, error) {
	var volume Volume
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{"size_gb": newSizeGB}).
		SetResult(&volume).
		Patch(fmt.Sprintf("/api/v1/clusters/%s/volumes/%s", c.clusterID, volumeID))

	if err != nil {
		return nil, fmt.Errorf("failed to resize volume: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to resize volume: %s", resp.String())
	}

	c.log.Infow("Volume resized", "id", volumeID, "new_size_gb", newSizeGB)
	return &volume, nil
}

// Snapshot represents a volume snapshot
type Snapshot struct {
	ID        string    `json:"id"`
	VolumeID  string    `json:"volume_id"`
	Name      string    `json:"name"`
	SizeGB    int       `json:"size_gb"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateSnapshot creates an instant snapshot
func (c *Client) CreateSnapshot(ctx context.Context, volumeID, name string) (*Snapshot, error) {
	var snapshot Snapshot
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"name": name}).
		SetResult(&snapshot).
		Post(fmt.Sprintf("/api/v1/clusters/%s/volumes/%s/snapshots", c.clusterID, volumeID))

	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create snapshot: %s", resp.String())
	}

	c.log.Infow("Snapshot created", "id", snapshot.ID, "volume_id", volumeID, "name", name)
	return &snapshot, nil
}

// RestoreSnapshot restores a volume from snapshot
func (c *Client) RestoreSnapshot(ctx context.Context, snapshotID, newVolumeName string) (*Volume, error) {
	var volume Volume
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"name": newVolumeName}).
		SetResult(&volume).
		Post(fmt.Sprintf("/api/v1/clusters/%s/snapshots/%s/restore", c.clusterID, snapshotID))

	if err != nil {
		return nil, fmt.Errorf("failed to restore snapshot: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to restore snapshot: %s", resp.String())
	}

	c.log.Infow("Snapshot restored", "snapshot_id", snapshotID, "new_volume", volume.ID)
	return &volume, nil
}

// NVMeConnectionInfo holds NVMe-oF connection details
type NVMeConnectionInfo struct {
	NQN       string   `json:"nqn"`
	Namespace int      `json:"namespace"`
	Targets   []string `json:"targets"`
	Transport string   `json:"transport"`
	HostNQN   string   `json:"host_nqn"`
}

// GetNVMeConnectionInfo returns NVMe-oF connection details
func (c *Client) GetNVMeConnectionInfo(ctx context.Context, volumeID string) (*NVMeConnectionInfo, error) {
	var info NVMeConnectionInfo
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&info).
		Get(fmt.Sprintf("/api/v1/clusters/%s/volumes/%s/connection", c.clusterID, volumeID))

	if err != nil {
		return nil, fmt.Errorf("failed to get connection info: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get connection info: %s", resp.String())
	}

	return &info, nil
}

// StorageClass represents a Simplyblock storage class template
type StorageClass struct {
	Name         string `json:"name"`
	StoragePool  string `json:"storage_pool"`
	IOPS         int    `json:"iops"`
	ThroughputMB int    `json:"throughput_mb"`
	Encryption   bool   `json:"encryption"`
	Replication  int    `json:"replication"`
}

// PredefinedStorageClasses returns NorthStack storage class definitions
func PredefinedStorageClasses() []StorageClass {
	return []StorageClass{
		{Name: "northstack-block-fast", StoragePool: "nvme-fast", IOPS: 100000, ThroughputMB: 1000, Encryption: true, Replication: 3},
		{Name: "northstack-block-standard", StoragePool: "nvme-standard", IOPS: 10000, ThroughputMB: 250, Encryption: true, Replication: 2},
		{Name: "northstack-block-economy", StoragePool: "nvme-economy", IOPS: 3000, ThroughputMB: 125, Encryption: true, Replication: 2},
	}
}
