package pipeline

import (
	"context"
	"io"

	"audio-streaming/internal/track"
)

// DecodeSpec says how Decoder.Open shapes the audio. The output is
// always interleaved stereo float.
type DecodeSpec struct {
	// Mix folds the source channels to stereo (2×N, see dsp.MixMatrix).
	// nil means the source is already stereo.
	Mix [][]float64
	// SampleRate is the output rate. Audio is resampled only when it
	// differs from the source rate.
	SampleRate int
}

// Decoder reads a source file.
type Decoder interface {
	// Probe describes the first audio stream without decoding it.
	Probe(ctx context.Context, path string) (track.SourceInfo, error)
	// Open streams the audio as float blocks shaped by spec.
	Open(ctx context.Context, path string, spec DecodeSpec) (
		BlockReader, error)
	// OpenPCM16 streams the audio unchanged as interleaved 16-bit
	// little-endian PCM. Only for sources that are already 16-bit.
	OpenPCM16(ctx context.Context, path string) (io.ReadCloser, error)
}

// MetaExtractor reads tags and cover art before the audio is transformed.
type MetaExtractor interface {
	Extract(ctx context.Context, path string) (track.Tags, *track.Cover, error)
}

// Encoder writes FLAC to dstPath. pcm is interleaved stereo 16-bit
// little-endian at 44.1 kHz. dstPath must be a file: FLAC rewrites its
// header once the stream ends.
type Encoder interface {
	Encode(ctx context.Context, pcm io.Reader, opt track.EncodeOptions,
		dstPath string) error
}

// Verifier independently checks the encoded file.
type Verifier interface {
	Verify(path string, exp track.Expectations) (track.Report, error)
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
