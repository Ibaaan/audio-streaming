//go:build integration

package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"testing"
)

const someRealFile = "/home/lwnds/projects/go/audio-streaming/test_data/Slipknot - Duality.flac"

func TestProbe(t *testing.T) {
	path := someRealFile
	info, err := newTestFFmpeg().Probe(context.Background(), path)

	if err != nil {
		return
	}

	t.Logf("\nprobe result: %+v", info)
}

func TestExtractReal(t *testing.T) {
	_ = RemoveCover("../../../test_data/")
	path := someRealFile
	tags, c, err := newTestFFmpeg().Extract(context.Background(), path)

	if err != nil {
		return
	}

	_, _ = WriteCover("../../../test_data", c.Data, c.MIMEType)

	t.Logf("\nprobe result: %#v", tags)
}

var coverExts = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg", // non-standard, but some taggers write it
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
	"image/bmp":  ".bmp",
}

func WriteCover(dir string, data []byte, mimeType string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return "", fmt.Errorf("parse mime type %q: %w", mimeType, err)
	}

	ext, ok := coverExts[mediaType]
	if !ok {
		return "", fmt.Errorf("unsupported cover mime type %q", mediaType)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	path := filepath.Join(dir, "cover"+ext)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write cover: %w", err)
	}
	return path, nil
}

func RemoveCover(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "cover.*"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
	}
	return nil
}
