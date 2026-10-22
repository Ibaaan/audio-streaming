package ffmpeg

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/audio"
)

func filterArg(t *testing.T, args []string) string {
	t.Helper()
	for i, a := range args {
		if a == "-af" {
			return args[i+1]
		}
	}
	t.Fatal("no -af argument")
	return ""
}

func TestDecodeArgs(t *testing.T) {
	tests := []struct {
		name string
		spec audio.DecodeSpec
		want string
	}{
		{
			"stereo",
			audio.DecodeSpec{SampleRate: 44100},
			"aformat=sample_fmts=flt," +
				"aresample=44100:resampler=soxr:precision=28",
		},
		{
			"mono",
			audio.DecodeSpec{
				Mix:        [][]float64{{1}, {1}},
				SampleRate: 44100,
			},
			"aformat=sample_fmts=flt,pan=stereo|c0=1*c0|c1=1*c0," +
				"aresample=44100:resampler=soxr:precision=28",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := decodeArgs("in.flac", tt.spec)

			assert.Equal(t, tt.want, filterArg(t, args))
		})
	}
}

func f32le(samples ...float32) io.ReadCloser {
	var b bytes.Buffer
	_ = binary.Write(&b, binary.LittleEndian, samples)
	return io.NopCloser(&b)
}

func TestFloatReaderBlock(t *testing.T) {
	r := newFloatReader(f32le(0.5, -0.5, 1, -1), 44100)

	b, err := r.ReadBlock(context.Background())

	require.NoError(t, err)
	assert.Equal(t, audio.Block{
		Samples:    []float32{0.5, -0.5, 1, -1},
		Channels:   2,
		SampleRate: 44100,
	}, b)
}

func TestFloatReaderEOF(t *testing.T) {
	r := newFloatReader(f32le(0.5, -0.5), 44100)
	_, err := r.ReadBlock(context.Background())
	require.NoError(t, err)

	_, err = r.ReadBlock(context.Background())

	assert.ErrorIs(t, err, io.EOF)
}

func TestFloatReaderHalfFrame(t *testing.T) {
	r := newFloatReader(f32le(0.5, -0.5, 0.25), 44100)

	_, err := r.ReadBlock(context.Background())

	assert.Error(t, err)
}

func TestPanFilter(t *testing.T) {
	tests := []struct {
		name string
		m    [][]float64
		want string
	}{
		{"mono", [][]float64{{1}, {1}}, "pan=stereo|c0=1*c0|c1=1*c0"},
		{
			"5.1",
			[][]float64{{1, 0, 0.5, 0, 0.5, 0}, {0, 1, 0.5, 0, 0, 0.5}},
			"pan=stereo|c0=1*c0+0.5*c2+0.5*c4|c1=1*c1+0.5*c2+0.5*c5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PanFilter(tt.m)

			assert.Equal(t, tt.want, got)
		})
	}
}
