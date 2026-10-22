// Package audio holds the types that carry decoded audio between
// packages. It has no logic and no dependencies.
package audio

import (
	"context"
	"io"
)

// Block is a run of interleaved float samples, nominally in [-1, 1).
type Block struct {
	Samples    []float32
	Channels   int
	SampleRate int
}

// Frames is the number of per-channel samples in the block.
func (b Block) Frames() int {
	if b.Channels == 0 {
		return 0
	}
	return len(b.Samples) / b.Channels
}

// BlockReader yields a track's audio one block at a time.
type BlockReader interface {
	// ReadBlock returns the next block, or io.EOF after the last one.
	// The block's Samples may be reused by the next call.
	ReadBlock(ctx context.Context) (Block, error)
	// Close releases the underlying source (file, process).
	Close() error
}

type SimpleBlockReader struct{}

func (r *SimpleBlockReader) ReadBlock(ctx context.Context) (Block, error) {
	return Block{}, io.EOF
}

func (r *SimpleBlockReader) Close() error {
	return nil
}

// DecodeSpec says how a decoder shapes the audio. The output is always
// interleaved stereo float.
type DecodeSpec struct {
	// Mix folds the source channels to stereo (2×N, see dsp.MixMatrix).
	// nil means the source is already stereo.
	Mix [][]float64
	// SampleRate is the output rate. Audio is resampled only when it
	// differs from the source rate.
	SampleRate int
}
