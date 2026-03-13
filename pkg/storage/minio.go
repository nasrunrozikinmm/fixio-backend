package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"fixio/pkg/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps MinIO operations for file uploads
type Client struct {
	minio    *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

// NewMinioClient creates and configures a new MinIO storage client
func NewMinioClient(env *config.Environment) (*Client, error) {
	client, err := minio.New(env.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(env.MinioAccessKey, env.MinioSecretKey, ""),
		Secure: env.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure bucket exists
	exists, err := client.BucketExists(ctx, env.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, env.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", env.MinioBucket, err)
		}

		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, env.MinioBucket)

		if err := client.SetBucketPolicy(ctx, env.MinioBucket, policy); err != nil {
			log.Printf("Warning: failed to set bucket policy: %v", err)
		}
	}

	log.Printf("📦 MinIO connected: %s (bucket: %s)", env.MinioEndpoint, env.MinioBucket)

	return &Client{minio: client, bucket: env.MinioBucket, endpoint: env.MinioEndpoint, useSSL: env.MinioUseSSL}, nil
}

// Upload stores a file in MinIO and returns its public URL
func (c *Client) Upload(ctx context.Context, objectName, contentType string, reader io.Reader, size int64) (string, error) {
	_, err := c.minio.PutObject(ctx, c.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object %s: %w", objectName, err)
	}

	return c.ObjectURL(objectName), nil
}

// Delete removes a file from MinIO
func (c *Client) Delete(ctx context.Context, objectName string) error {
	return c.minio.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}

// ObjectURL returns the public URL for an object
func (c *Client) ObjectURL(objectName string) string {
	protocol := "http"
	if c.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, c.endpoint, c.bucket, objectName)
}
