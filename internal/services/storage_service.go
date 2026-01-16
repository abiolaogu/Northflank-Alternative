package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/northstack/platform/pkg/minio"
	"github.com/northstack/platform/pkg/seaweedfs"
	"github.com/northstack/platform/pkg/simplyblock"
)

// StorageService provides unified access to all storage backends
type StorageService struct {
	block  *simplyblock.Client // NVMe-oF block storage
	file   *seaweedfs.Client   // Distributed file storage
	object *minio.Client       // S3-compatible object storage
	log    *zap.SugaredLogger
}

// NewStorageService creates a unified storage service
func NewStorageService(
	block *simplyblock.Client,
	file *seaweedfs.Client,
	object *minio.Client,
	log *zap.SugaredLogger,
) *StorageService {
	return &StorageService{block: block, file: file, object: object, log: log}
}

// StorageType defines the type of storage
type StorageType string

const (
	StorageTypeBlock  StorageType = "block"  // Simplyblock NVMe-oF
	StorageTypeFile   StorageType = "file"   // SeaweedFS
	StorageTypeObject StorageType = "object" // MinIO S3
)

// ============================================================================
// Block Storage Operations (Simplyblock)
// ============================================================================

// CreateBlockVolumeInput holds parameters for creating a block volume
type CreateBlockVolumeInput struct {
	Name          string
	SizeGB        int
	StorageClass  string
	IOPS          int
	ThroughputMB  int
	Replication   int
	ApplicationID string
	ProjectID     string
}

// BlockVolumeInfo represents block volume information
type BlockVolumeInfo struct {
	ID     string
	Name   string
	SizeGB int
	IOPS   int
	NQN    string
	Status string
}

// CreateBlockVolume creates a new block volume for databases/stateful apps
func (s *StorageService) CreateBlockVolume(ctx context.Context, input *CreateBlockVolumeInput) (*BlockVolumeInfo, error) {
	volume, err := s.block.CreateVolume(ctx, &simplyblock.CreateVolumeInput{
		Name:         input.Name,
		SizeGB:       input.SizeGB,
		StoragePool:  input.StorageClass,
		IOPS:         input.IOPS,
		ThroughputMB: input.ThroughputMB,
		Encryption:   true,
		Replication:  input.Replication,
		Labels: map[string]string{
			"northstack.io/application": input.ApplicationID,
			"northstack.io/project":     input.ProjectID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create block volume: %w", err)
	}

	s.log.Infow("Block volume created", "id", volume.ID, "name", volume.Name, "size_gb", volume.SizeGB)
	return &BlockVolumeInfo{
		ID: volume.ID, Name: volume.Name, SizeGB: volume.SizeGB,
		IOPS: volume.IOPS, NQN: volume.NQN, Status: volume.Status,
	}, nil
}

// ============================================================================
// File Storage Operations (SeaweedFS)
// ============================================================================

// UploadFile uploads a file to SeaweedFS
func (s *StorageService) UploadFile(ctx context.Context, path string, data io.Reader, contentType string) error {
	return s.file.Upload(ctx, path, data, contentType)
}

// DownloadFile downloads a file from SeaweedFS
func (s *StorageService) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.file.Download(ctx, path)
}

// DeleteFile deletes a file from SeaweedFS
func (s *StorageService) DeleteFile(ctx context.Context, path string) error {
	return s.file.Delete(ctx, path, false)
}

// CreateDirectory creates a directory in SeaweedFS
func (s *StorageService) CreateDirectory(ctx context.Context, path string) error {
	return s.file.Mkdir(ctx, path)
}

// FileInfo represents file metadata
type FileInfo struct {
	Path         string
	Size         int64
	IsDir        bool
	LastModified time.Time
}

// ListDirectory lists files in a directory
func (s *StorageService) ListDirectory(ctx context.Context, path string, limit int) ([]FileInfo, error) {
	entry, err := s.file.List(ctx, path, limit)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, e := range entry.Entries {
		files = append(files, FileInfo{Path: e.FullPath, Size: e.FileSize, IsDir: e.IsDir, LastModified: e.Mtime})
	}
	return files, nil
}

// ============================================================================
// Object Storage Operations (MinIO)
// ============================================================================

// ObjectInfo represents object metadata
type ObjectInfo struct {
	Bucket       string
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
}

// PutObject uploads an object to MinIO
func (s *StorageService) PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) (*ObjectInfo, error) {
	info, err := s.object.PutObject(ctx, bucket, key, data, size, contentType)
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{Bucket: info.Bucket, Key: info.Key, Size: info.Size, ETag: info.ETag, LastModified: info.LastModified}, nil
}

// GetObject downloads an object from MinIO
func (s *StorageService) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, *ObjectInfo, error) {
	reader, info, err := s.object.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, nil, err
	}
	return reader, &ObjectInfo{Bucket: info.Bucket, Key: info.Key, Size: info.Size, ETag: info.ETag, ContentType: info.ContentType, LastModified: info.LastModified}, nil
}

// DeleteObject deletes an object from MinIO
func (s *StorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	return s.object.DeleteObject(ctx, bucket, key)
}

// GetPresignedURL generates a presigned URL for object access
func (s *StorageService) GetPresignedURL(ctx context.Context, bucket, key string, expiry int64) (string, error) {
	return s.object.GetPresignedURL(ctx, bucket, key, time.Duration(expiry)*time.Second)
}

// ============================================================================
// Application Storage Provisioning
// ============================================================================

// ProvisionApplicationStorage provisions all storage for an application
func (s *StorageService) ProvisionApplicationStorage(ctx context.Context, appID, projectID string) error {
	// Create MinIO bucket for artifacts/uploads
	if err := s.object.CreateApplicationBucket(ctx, appID); err != nil {
		s.log.Warnw("Failed to create object bucket", "error", err)
	}

	// Create SeaweedFS directory structure
	basePath := fmt.Sprintf("/apps/%s", appID)
	dirs := []string{basePath + "/uploads", basePath + "/logs", basePath + "/cache"}

	for _, dir := range dirs {
		if err := s.file.Mkdir(ctx, dir); err != nil {
			s.log.Warnw("Failed to create directory", "path", dir, "error", err)
		}
	}

	s.log.Infow("Application storage provisioned", "app_id", appID, "project_id", projectID)
	return nil
}

// DeprovisionApplicationStorage removes all storage for an application
func (s *StorageService) DeprovisionApplicationStorage(ctx context.Context, appID string) error {
	if err := s.object.DeleteBucket(ctx, "app-"+appID, true); err != nil {
		s.log.Warnw("Failed to delete object bucket", "error", err)
	}

	if err := s.file.Delete(ctx, "/apps/"+appID, true); err != nil {
		s.log.Warnw("Failed to delete file directory", "error", err)
	}

	s.log.Infow("Application storage deprovisioned", "app_id", appID)
	return nil
}
