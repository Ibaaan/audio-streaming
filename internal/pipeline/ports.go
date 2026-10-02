package pipeline

import (
	"context"
	"io"

	"audio-streaming/internal/track"
)

// Decoder opens a source file and streams it as float blocks.
type Decoder interface {
	Open(ctx context.Context, path string) (track.SourceInfo,
		BlockReader, error)
}

// MetaExtractor reads tags and cover art before the audio is transformed.
type MetaExtractor interface {
	Extract(ctx context.Context, path string) (track.Tags, *track.Cover, error)
}

// ResamplerFactory builds a fresh resampler for each track.
type ResamplerFactory interface {
	New(fromRate, toRate, channels int) (Stage, error)
}

// Encoder writes the processed stream as FLAC.
type Encoder interface {
	Encode(ctx context.Context, src BlockReader, opt track.EncodeOptions,
		dst io.Writer) error
}

// Verifier independently checks the encoded file.
type Verifier interface {
	Verify(f io.ReadSeeker, exp track.Expectations) (track.Report, error)
}

// Uploader stores audio and cover blobs. Satisfied by *s3store.Store.
type Uploader interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64,
		contentType string) (int64, error)
}

// RecordSaver persists the track record linking metadata and blob keys.
type RecordSaver interface {
	Save(ctx context.Context, rec track.Record) error
}
