package ffmpeg

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"audio-streaming/internal/audio"
)

const (
	blockFrames = 4096
	outChannels = 2
	frameBytes  = outChannels * 4 // float32 per channel
)

// Open streams path as interleaved stereo float32 shaped by spec.
func (f *FFmpeg) Open(
	ctx context.Context, path string, spec audio.DecodeSpec,
) (audio.BlockReader, error) {
	if spec.SampleRate <= 0 {
		return nil, fmt.Errorf("ffmpeg: invalid sample rate %d",
			spec.SampleRate)
	}
	p, err := start(ctx, f.ffmpeg, decodeArgs(path, spec))
	if err != nil {
		return nil, err
	}
	return newFloatReader(p, spec.SampleRate), nil
}

// OpenPCM16 streams path unchanged as 16-bit little-endian PCM.
func (f *FFmpeg) OpenPCM16(ctx context.Context, path string) (
	io.ReadCloser, error,
) {
	args := append(inputArgs(path),
		"-map", "0:a:0", "-f", "s16le", "-c:a", "pcm_s16le", "-")
	return start(ctx, f.ffmpeg, args)
}

// decodeArgs builds the filter chain. The first aformat makes ffmpeg
// convert to float before mixing and resampling; otherwise pan and
// aresample could work on (and round to) 16-bit integers. aresample
// does nothing when the rates already match.
func decodeArgs(path string, spec audio.DecodeSpec) []string {
	filters := []string{"aformat=sample_fmts=flt"}
	if spec.Mix != nil {
		filters = append(filters, PanFilter(spec.Mix))
	}
	filters = append(filters, fmt.Sprintf(
		"aresample=%d:resampler=soxr:precision=28", spec.SampleRate))

	return append(inputArgs(path),
		"-map", "0:a:0",
		"-af", strings.Join(filters, ","),
		"-f", "f32le", "-c:a", "pcm_f32le", "-")
}

// floatReader cuts raw f32le stereo into blocks.
type floatReader struct {
	src  io.ReadCloser
	rate int
	buf  []byte
	out  []float32
}

func newFloatReader(src io.ReadCloser, rate int) *floatReader {
	return &floatReader{
		src:  src,
		rate: rate,
		buf:  make([]byte, blockFrames*frameBytes),
		out:  make([]float32, blockFrames*outChannels),
	}
}

func (r *floatReader) ReadBlock(ctx context.Context) (
	audio.Block, error,
) {
	if err := ctx.Err(); err != nil {
		return audio.Block{}, err
	}
	n, err := io.ReadFull(r.src, r.buf)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil // short final block
	}
	if err != nil {
		return audio.Block{}, err
	}
	if n%frameBytes != 0 {
		return audio.Block{}, fmt.Errorf(
			"ffmpeg: stream ends mid-frame (%d bytes)", n)
	}

	samples := r.out[:n/4]
	for i := range samples {
		bits := binary.LittleEndian.Uint32(r.buf[4*i:])
		samples[i] = math.Float32frombits(bits)
	}
	return audio.Block{
		Samples:    samples,
		Channels:   outChannels,
		SampleRate: r.rate,
	}, nil
}

func (r *floatReader) Close() error { return r.src.Close() }
