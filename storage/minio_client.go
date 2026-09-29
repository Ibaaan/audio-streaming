package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	mc     *minio.Client
	bucket string
}

func NewClient(ctx context.Context, cfg Config) (*MinioClient, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.AccessKey,
			cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}

	c := &MinioClient{mc: mc, bucket: cfg.Bucket}
	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *MinioClient) ensureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("checking bucket %q: %w", c.bucket, err)
	}
	if exists {
		return nil
	}
	err = c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("creating bucket %q: %w", c.bucket, err)
	}
	return nil
}

func (c *MinioClient) Upload(
	ctx context.Context, key string, r io.Reader,
	size int64, contentType string,
) (int64, error) {
	info, err := c.mc.PutObject(
		ctx, c.bucket, key, r, size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return 0, fmt.Errorf("uploading %q: %w", key, err)
	}
	return info.Size, nil
}

func (c *MinioClient) List(
	ctx context.Context, prefix string,
) ([]string, error) {
	var keys []string
	opts := minio.ListObjectsOptions{Prefix: prefix, Recursive: true}
	for obj := range c.mc.ListObjects(ctx, c.bucket, opts) {
		if obj.Err != nil {
			return nil, fmt.Errorf("listing objects: %w", obj.Err)
		}
		keys = append(keys, obj.Key)
	}
	return keys, nil
}

func (c *MinioClient) Delete(ctx context.Context, key string) error {
	err := c.mc.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("deleting %q: %w", key, err)
	}
	return nil
}

func (c *MinioClient) PresignedGetURL(
	ctx context.Context, key string, expiry time.Duration,
) (string, error) {
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigning %q: %w", key, err)
	}
	return u.String(), nil
}
