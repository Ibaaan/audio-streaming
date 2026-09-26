package storage

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
)

type TestVolume struct {
	Client *MinioClient
	Keys   []string
}

func CreateTestVolume(
	ctx context.Context, cfg Config, dir string,
) (*TestVolume, error) {
	cfg.Bucket = fmt.Sprintf("test-%d", time.Now().UnixNano())
	c, err := NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	v := &TestVolume{Client: c}
	err = filepath.WalkDir(dir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			key := filepath.ToSlash(rel)
			if err := uploadFile(ctx, c, key, path); err != nil {
				return err
			}
			v.Keys = append(v.Keys, key)
			return nil
		})
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("seeding from %q: %w", dir, err), v.Delete(ctx))
	}
	return v, nil
}

func uploadFile(
	ctx context.Context, c *MinioClient, key, path string,
) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err = c.Upload(ctx, key, f, info.Size(), contentType)
	return err
}

func (v *TestVolume) Delete(ctx context.Context) error {
	c := v.Client
	opts := minio.ListObjectsOptions{Recursive: true}
	for obj := range c.mc.ListObjects(ctx, c.bucket, opts) {
		if obj.Err != nil {
			return fmt.Errorf("listing %q: %w", c.bucket, obj.Err)
		}
		if err := c.Delete(ctx, obj.Key); err != nil {
			return err
		}
	}
	if err := c.mc.RemoveBucket(ctx, c.bucket); err != nil {
		return fmt.Errorf("removing bucket %q: %w", c.bucket, err)
	}
	return nil
}
