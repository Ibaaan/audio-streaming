package pipeline

import (
	"context"
)

// BlockReader yields a track's audio one block at a time. Decoder
// produces one; Encoder drains one.
type BlockReader interface {
	// ReadBlock returns the next block, or io.EOF after the last one.
	// The block's Samples may be reused by the next call.
	ReadBlock(ctx context.Context) (Block, error)
	// Close releases the underlying source (file, decoder state).
	Close() error
}

// Stage transforms blocks between decoding and encoding, e.g.
// resampling. A Stage keeps state across blocks, so it serves one
// track only.
type Stage interface {
	// Process transforms in. The result may hold fewer or more frames
	// than in, and may be empty while the stage buffers input.
	Process(in Block) (Block, error)
	// Flush returns whatever the stage still buffers. Call it once,
	// after the last Process.
	Flush() (Block, error)
}

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
