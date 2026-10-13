package pipeline

import (
	"context"
	"io"

	"audio-streaming/internal/audio"
)

// Decoder reads a source file.
type Decoder interface {
	// Probe describes the first audio stream without decoding it.
	Probe(ctx context.Context, path string) (audio.SourceInfo, error)
	// Open streams the audio as float blocks shaped by spec.
	Open(ctx context.Context, path string, spec audio.DecodeSpec) (
		audio.BlockReader, error)
	// OpenPCM16 streams the audio unchanged as interleaved 16-bit
	// little-endian PCM. Only for sources that are already 16-bit.
	OpenPCM16(ctx context.Context, path string) (io.ReadCloser, error)
}

// MetaExtractor reads tags and cover art before the audio is transformed.
type MetaExtractor interface {
	Extract(ctx context.Context, path string) (audio.Tags, *audio.Cover, error)
}

// Encoder writes FLAC to dstPath. pcm is interleaved stereo 16-bit
// little-endian at 44.1 kHz. dstPath must be a file: FLAC rewrites its
// header once the stream ends.
type Encoder interface {
	Encode(ctx context.Context, pcm io.Reader, opt audio.EncodeOptions,
		dstPath string) error
}

// Verifier independently checks the encoded file.
type Verifier interface {
	Verify(path string, exp audio.Expectations) (audio.Report, error)
}

// Uploader stores audio and cover blobs. Satisfied by *s3store.Store.
type Uploader interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64,
		contentType string) (int64, error)
}

// RecordSaver persists the track record linking metadata and blob keys.
type RecordSaver interface {
	Save(ctx context.Context, rec audio.Record) error
}
