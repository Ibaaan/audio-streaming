//go:build integration

package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T) *Store {
	t.Helper()

	cfg := ConfigFromEnv()

	cfg.Bucket = fmt.Sprintf("it-%d", time.Now().UnixNano())

	ctx, cancel := context.WithTimeout(context.Background(),
		10*time.Second)
	defer cancel()
	c, err := New(ctx, cfg)
	require.NoError(t, err, "NewClient")

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		opts := minio.ListObjectsOptions{Recursive: true}
		for obj := range c.mc.ListObjects(ctx, c.bucket, opts) {
			if obj.Err == nil {
				_ = c.mc.RemoveObject(ctx, c.bucket, obj.Key,
					minio.RemoveObjectOptions{})
			}
		}
		if err := c.mc.RemoveBucket(ctx, c.bucket); err != nil {
			t.Logf("removing bucket %q: %v", c.bucket, err)
		}
	})
	return c
}

func upload(t *testing.T, c *Store, key, body string) {
	t.Helper()
	n, err := c.Upload(context.Background(), key, strings.NewReader(body),
		int64(len(body)), "text/plain")
	require.NoError(t, err, "Upload(%q)", key)
	require.EqualValues(t, len(body), n, "Upload(%q) size", key)
}

func TestUpload(t *testing.T) {
	c := newTestClient(t)
	const body = "audio bytes"

	n, err := c.Upload(context.Background(), "songs/a.mp3",
		strings.NewReader(body), int64(len(body)), "audio/mpeg")

	require.NoError(t, err)
	assert.EqualValues(t, len(body), n)
}

func TestList(t *testing.T) {
	c := newTestClient(t)
	upload(t, c, "songs/a.mp3", "aaa")
	upload(t, c, "songs/b.mp3", "bbbb")
	upload(t, c, "covers/a.jpg", "c")

	keys, err := c.List(context.Background(), "")

	require.NoError(t, err)
	assert.ElementsMatch(t,
		[]string{"covers/a.jpg", "songs/a.mp3", "songs/b.mp3"}, keys)
}

func TestListWithPrefix(t *testing.T) {
	c := newTestClient(t)
	upload(t, c, "songs/a.mp3", "aaa")
	upload(t, c, "songs/b.mp3", "bbbb")
	upload(t, c, "covers/a.jpg", "c")

	keys, err := c.List(context.Background(), "songs/")

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"songs/a.mp3", "songs/b.mp3"}, keys)
}

func TestListEmpty(t *testing.T) {
	c := newTestClient(t)

	keys, err := c.List(context.Background(), "")

	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestDelete(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	upload(t, c, "keep.txt", "keep")
	upload(t, c, "gone.txt", "gone")

	err := c.Delete(ctx, "gone.txt")

	require.NoError(t, err)
	keys, err := c.List(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"keep.txt"}, keys)
}

func TestDeleteMissingKey(t *testing.T) {
	c := newTestClient(t)

	err := c.Delete(context.Background(), "missing.txt")

	assert.NoError(t, err)
}

func TestPresignedGetURL(t *testing.T) {
	c := newTestClient(t)
	const body = "presigned content"
	upload(t, c, "track.txt", body)

	u, err := c.PresignedGetURL(context.Background(), "track.txt", time.Minute)

	require.NoError(t, err)
	resp, err := http.Get(u)
	require.NoError(t, err)
	defer resp.Body.Close()
	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", got)
	assert.Equal(t, body, string(got))
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
}

func TestUploadContextCanceled(t *testing.T) {
	c := newTestClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Upload(ctx, "x", strings.NewReader("x"), 1, "text/plain")

	assert.ErrorIs(t, err, context.Canceled)
}

func TestListContextCanceled(t *testing.T) {
	c := newTestClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.List(ctx, "")

	assert.ErrorIs(t, err, context.Canceled)
}
