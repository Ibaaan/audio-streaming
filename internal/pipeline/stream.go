package pipeline

import "audio-streaming/internal/audio"

// Stage transforms blocks between decoding and encoding. A Stage keeps
// state across blocks, so it serves one track only.
type Stage interface {
	// Process transforms in. The result may hold fewer or more frames
	// than in, and may be empty while the stage buffers input.
	Process(in audio.Block) (audio.Block, error)
	// Flush returns whatever the stage still buffers. Call it once,
	// after the last Process.
	Flush() (audio.Block, error)
}
