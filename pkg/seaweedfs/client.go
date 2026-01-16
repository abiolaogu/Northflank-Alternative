package seaweedfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Client provides integration with SeaweedFS
type Client struct {
	masterClient *resty.Client
	filerClient  *resty.Client
	masterURL    string
	filerURL     string
	log          *zap.SugaredLogger
}

// Config holds SeaweedFS client configuration
type Config struct {
	MasterURL string        `mapstructure:"master_url"`
	FilerURL  string        `mapstructure:"filer_url"`
	Timeout   time.Duration `mapstructure:"timeout"`
	JWT       string        `mapstructure:"jwt"`
}

// NewClient creates a new SeaweedFS client
func NewClient(cfg *Config, log *zap.SugaredLogger) (*Client, error) {
	masterClient := resty.New().SetBaseURL(cfg.MasterURL).SetTimeout(cfg.Timeout).SetHeader("Content-Type", "application/json")
	filerClient := resty.New().SetBaseURL(cfg.FilerURL).SetTimeout(cfg.Timeout)

	if cfg.JWT != "" {
		masterClient.SetHeader("Authorization", "Bearer "+cfg.JWT)
		filerClient.SetHeader("Authorization", "Bearer "+cfg.JWT)
	}

	return &Client{
		masterClient: masterClient,
		filerClient:  filerClient,
		masterURL:    cfg.MasterURL,
		filerURL:     cfg.FilerURL,
		log:          log,
	}, nil
}

// AssignResponse represents a file ID assignment from master
type AssignResponse struct {
	FileID    string `json:"fid"`
	URL       string `json:"url"`
	PublicURL string `json:"publicUrl"`
	Count     int    `json:"count"`
}

// Assign gets a file ID for upload
func (c *Client) Assign(ctx context.Context, count int, replication, collection string) (*AssignResponse, error) {
	var resp AssignResponse
	params := map[string]string{"count": fmt.Sprintf("%d", count)}
	if replication != "" {
		params["replication"] = replication
	}
	if collection != "" {
		params["collection"] = collection
	}

	r, err := c.masterClient.R().SetContext(ctx).SetQueryParams(params).SetResult(&resp).Get("/dir/assign")
	if err != nil {
		return nil, fmt.Errorf("failed to assign file ID: %w", err)
	}
	if r.IsError() {
		return nil, fmt.Errorf("failed to assign file ID: %s", r.String())
	}

	return &resp, nil
}

// ClusterStatus represents the cluster status
type ClusterStatus struct {
	IsLeader bool         `json:"IsLeader"`
	Leader   string       `json:"Leader"`
	Peers    []string     `json:"Peers"`
	Topology TopologyInfo `json:"Topology"`
}

// TopologyInfo contains cluster topology
type TopologyInfo struct {
	DataCenters []DataCenter `json:"DataCenters"`
	Free        int64        `json:"Free"`
	Max         int64        `json:"Max"`
}

// DataCenter represents a datacenter
type DataCenter struct {
	ID    string `json:"Id"`
	Racks []Rack `json:"Racks"`
}

// Rack represents a rack
type Rack struct {
	ID        string     `json:"Id"`
	DataNodes []DataNode `json:"DataNodes"`
}

// DataNode represents a data node
type DataNode struct {
	ID        string `json:"Id"`
	URL       string `json:"Url"`
	PublicURL string `json:"PublicUrl"`
	Volumes   int    `json:"Volumes"`
	Max       int    `json:"Max"`
	Free      int    `json:"Free"`
}

// GetClusterStatus retrieves cluster status from master
func (c *Client) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	var status ClusterStatus
	r, err := c.masterClient.R().SetContext(ctx).SetResult(&status).Get("/cluster/status")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster status: %w", err)
	}
	if r.IsError() {
		return nil, fmt.Errorf("failed to get cluster status: %s", r.String())
	}
	return &status, nil
}

// FileInfo represents file metadata
type FileInfo struct {
	FullPath    string            `json:"FullPath"`
	Mtime       time.Time         `json:"Mtime"`
	Crtime      time.Time         `json:"Crtime"`
	Mode        uint32            `json:"Mode"`
	UID         uint32            `json:"Uid"`
	GID         uint32            `json:"Gid"`
	Mime        string            `json:"Mime"`
	Replication string            `json:"Replication"`
	Collection  string            `json:"Collection"`
	TtlSec      int32             `json:"TtlSec"`
	FileSize    int64             `json:"FileSize"`
	Extended    map[string][]byte `json:"Extended"`
}

// DirectoryEntry represents a directory listing entry
type DirectoryEntry struct {
	FullPath string  `json:"FullPath"`
	Entries  []Entry `json:"Entries"`
}

// Entry represents a single entry in a directory
type Entry struct {
	FullPath string    `json:"FullPath"`
	Mtime    time.Time `json:"Mtime"`
	Crtime   time.Time `json:"Crtime"`
	Mode     uint32    `json:"Mode"`
	UID      uint32    `json:"Uid"`
	GID      uint32    `json:"Gid"`
	Mime     string    `json:"Mime"`
	FileSize int64     `json:"FileSize"`
	IsDir    bool      `json:"isDir,omitempty"`
}

// Upload uploads a file to SeaweedFS filer
func (c *Client) Upload(ctx context.Context, path string, data io.Reader, contentType string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, data); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}
	writer.Close()

	r, err := c.filerClient.R().SetContext(ctx).SetHeader("Content-Type", writer.FormDataContentType()).SetBody(body.Bytes()).Post(path)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("failed to upload file: %s", r.String())
	}

	c.log.Debugw("File uploaded", "path", path)
	return nil
}

// Download downloads a file from SeaweedFS filer
func (c *Client) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	r, err := c.filerClient.R().SetContext(ctx).SetDoNotParseResponse(true).Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	if r.IsError() {
		r.RawBody().Close()
		return nil, fmt.Errorf("failed to download file: status %d", r.StatusCode())
	}
	return r.RawBody(), nil
}

// Delete deletes a file or directory
func (c *Client) Delete(ctx context.Context, path string, recursive bool) error {
	req := c.filerClient.R().SetContext(ctx)
	if recursive {
		req.SetQueryParam("recursive", "true")
	}

	r, err := req.Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("failed to delete: %s", r.String())
	}

	c.log.Debugw("Deleted", "path", path, "recursive", recursive)
	return nil
}

// List lists directory contents
func (c *Client) List(ctx context.Context, path string, limit int) (*DirectoryEntry, error) {
	var entry DirectoryEntry
	r, err := c.filerClient.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{"limit": fmt.Sprintf("%d", limit)}).
		SetHeader("Accept", "application/json").
		SetResult(&entry).
		Get(path + "?pretty=y")

	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}
	if r.IsError() {
		return nil, fmt.Errorf("failed to list directory: %s", r.String())
	}

	return &entry, nil
}

// Mkdir creates a directory
func (c *Client) Mkdir(ctx context.Context, path string) error {
	r, err := c.filerClient.R().SetContext(ctx).SetHeader("Content-Type", "application/x-directory").Post(path + "/")
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("failed to create directory: %s", r.String())
	}
	return nil
}

// Stat gets file/directory metadata
func (c *Client) Stat(ctx context.Context, path string) (*FileInfo, error) {
	var info FileInfo
	r, err := c.filerClient.R().
		SetContext(ctx).
		SetQueryParam("metadata", "true").
		SetHeader("Accept", "application/json").
		SetResult(&info).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("failed to stat: %w", err)
	}
	if r.IsError() {
		return nil, fmt.Errorf("failed to stat: %s", r.String())
	}

	return &info, nil
}

// CreateBucket creates a bucket (collection) for organizing files
func (c *Client) CreateBucket(ctx context.Context, bucket, replication string) error {
	params := map[string]string{}
	if replication != "" {
		params["replication"] = replication
	}

	r, err := c.filerClient.R().SetContext(ctx).SetQueryParams(params).Post("/buckets/" + bucket + "/")
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("failed to create bucket: %s", r.String())
	}

	c.log.Infow("Bucket created", "bucket", bucket, "replication", replication)
	return nil
}

// SetQuota sets storage quota for a path
func (c *Client) SetQuota(ctx context.Context, path string, sizeBytes int64) error {
	r, err := c.filerClient.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{"quota": sizeBytes}).
		Post(path + "?quota=set")

	if err != nil {
		return fmt.Errorf("failed to set quota: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("failed to set quota: %s", r.String())
	}

	return nil
}
