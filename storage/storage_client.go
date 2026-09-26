package storage

import (
	"context"
	"io"
	"time"
)

type StorageClient interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64,
		contentType string) (int64, error)

	List(ctx context.Context, prefix string) ([]string, error)

	Delete(ctx context.Context, key string) error

	PresignedGetURL(ctx context.Context, key string,
		expiry time.Duration) (string, error)
}
