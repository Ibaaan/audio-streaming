package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"audio-streaming/internal/audio"
	"audio-streaming/internal/catalog"
	"audio-streaming/internal/pipeline"
	"audio-streaming/internal/service"
)

// recordSaver adapts *service.Library to pipeline.RecordSaver. Library
// only takes covers that are already stored, so recordSaver uploads
// rec.Cover first, under its content hash so that identical covers share
// one blob.
type recordSaver struct {
	lib   *service.Library
	blobs pipeline.Uploader
}

var _ pipeline.RecordSaver = recordSaver{}

func (s recordSaver) Save(ctx context.Context, rec audio.Record) error {
	cover, err := s.uploadCover(ctx, rec.Cover)
	if err != nil {
		return err
	}
	_, err = s.lib.Save(ctx, rec, cover)
	return err
}

var coverFormats = map[string]struct {
	format catalog.ImageFormat
	ext    string
}{
	"image/jpeg": {catalog.ImageFormatJPEG, ".jpg"},
	"image/png":  {catalog.ImageFormatPNG, ".png"},
	"image/webp": {catalog.ImageFormatWEBP, ".webp"},
	"image/gif":  {catalog.ImageFormatGIF, ".gif"},
	"image/avif": {catalog.ImageFormatAVIF, ".avif"},
}

// uploadCover returns nil for no cover, or for one in a format the
// catalog does not know.
func (s recordSaver) uploadCover(
	ctx context.Context, cover *audio.Cover,
) (*catalog.Image, error) {
	if cover == nil || len(cover.Data) == 0 {
		return nil, nil
	}
	f, ok := coverFormats[cover.MIMEType]
	if !ok {
		return nil, nil
	}

	sum := sha256.Sum256(cover.Data)
	key := "covers/" + hex.EncodeToString(sum[:]) + f.ext
	_, err := s.blobs.Upload(ctx, key, bytes.NewReader(cover.Data),
		int64(len(cover.Data)), cover.MIMEType)
	if err != nil {
		return nil, err
	}

	img := &catalog.Image{StoragePath: key, OriginalFormat: f.format}
	// Formats the standard library cannot decode keep no size.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(cover.Data))
	if err == nil && cfg.Width > 0 && cfg.Height > 0 {
		w, h := int32(cfg.Width), int32(cfg.Height)
		img.Width, img.Height = &w, &h
	}
	return img, nil
}

// errNotImplemented fills the pipeline's Encoder and Verifier until they
// exist. Run does not call them yet.
var errNotImplemented = errors.New("not implemented")

type noEncoder struct{}

func (noEncoder) Encode(
	ctx context.Context, pcm io.Reader,
	opt audio.EncodeOptions, dstPath string,
) error {
	return errNotImplemented
}

type noVerifier struct{}

func (noVerifier) Verify(path string, exp audio.Expectations) (
	audio.Report, error,
) {
	return audio.Report{}, errNotImplemented
}
