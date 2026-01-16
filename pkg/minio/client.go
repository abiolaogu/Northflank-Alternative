package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// Client provides integration with MinIO (forked version)
// Fork: https://github.com/abiolaogu/MinIO
type Client struct {
	client *minio.Client
	config *Config
	log    *zap.SugaredLogger
}

// Config holds MinIO client configuration
type Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKey       string `mapstructure:"access_key"`
	SecretKey       string `mapstructure:"secret_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	Region          string `mapstructure:"region"`
	BucketPrefix    string `mapstructure:"bucket_prefix"`
	DefaultLocation string `mapstructure:"default_location"`
}

// NewClient creates a new MinIO client
func NewClient(cfg *Config, log *zap.SugaredLogger) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &Client{client: minioClient, config: cfg, log: log}, nil
}

// BucketInfo represents bucket information
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// CreateBucket creates a new bucket with optional versioning
func (c *Client) CreateBucket(ctx context.Context, name string, versioning bool) error {
	bucketName := c.config.BucketPrefix + name

	err := c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: c.config.DefaultLocation})
	if err != nil {
		exists, errBucketExists := c.client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			c.log.Debugw("Bucket already exists", "bucket", bucketName)
			return nil
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	if versioning {
		err = c.client.EnableVersioning(ctx, bucketName)
		if err != nil {
			c.log.Warnw("Failed to enable versioning", "bucket", bucketName, "error", err)
		}
	}

	c.log.Infow("Bucket created", "bucket", bucketName, "versioning", versioning)
	return nil
}

// DeleteBucket deletes a bucket
func (c *Client) DeleteBucket(ctx context.Context, name string, force bool) error {
	bucketName := c.config.BucketPrefix + name

	if force {
		objectsCh := c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true})
		for object := range objectsCh {
			if object.Err != nil {
				return fmt.Errorf("error listing objects: %w", object.Err)
			}
			err := c.client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete object %s: %w", object.Key, err)
			}
		}
	}

	err := c.client.RemoveBucket(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	c.log.Infow("Bucket deleted", "bucket", bucketName)
	return nil
}

// ListBuckets lists all buckets
func (c *Client) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	buckets, err := c.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var result []BucketInfo
	for _, b := range buckets {
		result = append(result, BucketInfo{Name: b.Name, CreationDate: b.CreationDate})
	}
	return result, nil
}

// ObjectInfo represents object metadata
type ObjectInfo struct {
	Bucket       string
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
	VersionID    string
}

// PutObject uploads an object to a bucket
func (c *Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*ObjectInfo, error) {
	bucketName := c.config.BucketPrefix + bucket
	opts := minio.PutObjectOptions{ContentType: contentType}

	info, err := c.client.PutObject(ctx, bucketName, key, reader, size, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	c.log.Debugw("Object uploaded", "bucket", bucketName, "key", key, "size", info.Size)
	return &ObjectInfo{Bucket: bucket, Key: key, Size: info.Size, ETag: info.ETag, LastModified: time.Now()}, nil
}

// GetObject retrieves an object from a bucket
func (c *Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, *ObjectInfo, error) {
	bucketName := c.config.BucketPrefix + bucket

	obj, err := c.client.GetObject(ctx, bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, nil, fmt.Errorf("failed to stat object: %w", err)
	}

	info := &ObjectInfo{
		Bucket:       bucket,
		Key:          key,
		Size:         stat.Size,
		ETag:         stat.ETag,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
	}

	return obj, info, nil
}

// DeleteObject deletes an object from a bucket
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	bucketName := c.config.BucketPrefix + bucket
	err := c.client.RemoveObject(ctx, bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	c.log.Debugw("Object deleted", "bucket", bucketName, "key", key)
	return nil
}

// ListObjects lists objects in a bucket
func (c *Client) ListObjects(ctx context.Context, bucket, prefix string, recursive bool, maxKeys int) ([]ObjectInfo, error) {
	bucketName := c.config.BucketPrefix + bucket
	opts := minio.ListObjectsOptions{Prefix: prefix, Recursive: recursive, MaxKeys: maxKeys}

	var objects []ObjectInfo
	for obj := range c.client.ListObjects(ctx, bucketName, opts) {
		if obj.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", obj.Err)
		}
		objects = append(objects, ObjectInfo{
			Bucket: bucket, Key: obj.Key, Size: obj.Size, ETag: obj.ETag,
			ContentType: obj.ContentType, LastModified: obj.LastModified,
		})
	}
	return objects, nil
}

// GetPresignedURL generates a presigned URL for downloading
func (c *Client) GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	bucketName := c.config.BucketPrefix + bucket
	presignedURL, err := c.client.PresignedGetObject(ctx, bucketName, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// PutPresignedURL generates a presigned URL for uploading
func (c *Client) PutPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	bucketName := c.config.BucketPrefix + bucket
	presignedURL, err := c.client.PresignedPutObject(ctx, bucketName, key, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// SetBucketPolicy sets a bucket policy
func (c *Client) SetBucketPolicy(ctx context.Context, bucket, policy string) error {
	bucketName := c.config.BucketPrefix + bucket
	err := c.client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}
	return nil
}

// MakePublicReadOnly makes a bucket publicly readable
func (c *Client) MakePublicReadOnly(ctx context.Context, bucket string) error {
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":"*"},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, c.config.BucketPrefix+bucket)
	return c.SetBucketPolicy(ctx, bucket, policy)
}

// LifecycleRule represents a lifecycle rule
type LifecycleRule struct {
	ID             string
	Prefix         string
	ExpirationDays int
}

// SetLifecycleRule sets a lifecycle rule for a bucket
func (c *Client) SetLifecycleRule(ctx context.Context, bucket string, rule *LifecycleRule) error {
	bucketName := c.config.BucketPrefix + bucket

	// Note: In production, use minio-go's lifecycle.Configuration
	// This is a simplified version - full implementation would use the lifecycle package
	c.log.Infow("Lifecycle rule configured (pending full implementation)",
		"bucket", bucketName, "rule_id", rule.ID, "expiration_days", rule.ExpirationDays)
	return nil
}

// CreateApplicationBucket creates a bucket for an application with standard policies
func (c *Client) CreateApplicationBucket(ctx context.Context, appID string) error {
	bucket := "app-" + appID

	if err := c.CreateBucket(ctx, bucket, true); err != nil {
		return err
	}

	c.SetLifecycleRule(ctx, bucket, &LifecycleRule{ID: "cleanup-temp", Prefix: "tmp/", ExpirationDays: 7})
	c.SetLifecycleRule(ctx, bucket, &LifecycleRule{ID: "cleanup-logs", Prefix: "logs/", ExpirationDays: 30})

	return nil
}

// CreateBackupBucket creates a bucket for backups with appropriate settings
func (c *Client) CreateBackupBucket(ctx context.Context, projectID string) error {
	bucket := "backup-" + projectID

	if err := c.CreateBucket(ctx, bucket, true); err != nil {
		return err
	}

	c.SetLifecycleRule(ctx, bucket, &LifecycleRule{ID: "cleanup-daily", Prefix: "daily/", ExpirationDays: 30})
	return nil
}
