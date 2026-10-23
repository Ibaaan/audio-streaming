package server

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

// RecordSaver adapts *service.Library to pipeline.RecordSaver. Library
// only takes covers that are already stored, so RecordSaver uploads
// rec.Cover first, under its content hash so that identical covers share
// one blob.
type RecordSaver struct {
	Lib   *service.Library
	Blobs pipeline.Uploader
}

var _ pipeline.RecordSaver = RecordSaver{}

func (s RecordSaver) Save(ctx context.Context, rec audio.Record) error {
	cover, err := s.uploadCover(ctx, rec.Cover)
	if err != nil {
		return err
	}
	_, err = s.Lib.Save(ctx, rec, cover)
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
func (s RecordSaver) uploadCover(ctx context.Context, cover *audio.Cover) (*catalog.Image, error) {
	if cover == nil || len(cover.Data) == 0 {
		return nil, nil
	}
	f, ok := coverFormats[cover.MIMEType]
	if !ok {
		return nil, nil
	}

	sum := sha256.Sum256(cover.Data)
	key := "covers/" + hex.EncodeToString(sum[:]) + f.ext
	_, err := s.Blobs.Upload(ctx, key, bytes.NewReader(cover.Data),
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

type NoEncoder struct{}

func (NoEncoder) Encode(
	ctx context.Context, pcm io.Reader,
	opt audio.EncodeOptions, dstPath string,
) error {
	return errNotImplemented
}

type NoVerifier struct{}

func (NoVerifier) Verify(path string, exp audio.Expectations) (audio.Report, error) {
	return audio.Report{}, errNotImplemented
}
