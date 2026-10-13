//go:build integration

package ffmpeg

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/audio"
	"audio-streaming/internal/dsp"
)

// Binaries come from FFMPEG_BIN / FFPROBE_BIN, or PATH.

func newTestFFmpeg() *FFmpeg {
	return New(os.Getenv("FFMPEG_BIN"), os.Getenv("FFPROBE_BIN"))
}

// gen runs ffmpeg to create a test file in a temp dir.
func gen(t *testing.T, name string, args ...string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), name)
	args = append([]string{"-nostdin", "-loglevel", "error", "-y"}, args...)
	cmd := exec.Command(newTestFFmpeg().ffmpeg, append(args, out)...)
	msg, err := cmd.CombinedOutput()
	require.NoError(t, err, string(msg))
	return out
}

// sines makes a 2-second source with one sine per channel.
func sines(t *testing.T, name, layout string, rate string,
	extra ...string) string {
	t.Helper()
	exprs := map[string]string{
		"mono":   "0.5*sin(440*2*PI*t)",
		"stereo": "0.5*sin(440*2*PI*t)|0.4*sin(660*2*PI*t)",
		"5.1": "0.5*sin(440*2*PI*t)|0.4*sin(550*2*PI*t)|" +
			"0.3*sin(660*2*PI*t)|0.2*sin(50*2*PI*t)|" +
			"0.2*sin(770*2*PI*t)|0.2*sin(880*2*PI*t)",
	}
	src := "aevalsrc=" + exprs[layout] +
		":c=" + layout + ":s=" + rate + ":d=2"
	return gen(t, name, append([]string{"-f", "lavfi", "-i", src},
		extra...)...)
}

func readFloats(t *testing.T, r audio.BlockReader) []float32 {
	t.Helper()
	defer func() { require.NoError(t, r.Close()) }()
	var all []float32
	for {
		b, err := r.ReadBlock(context.Background())
		if errors.Is(err, io.EOF) {
			return all
		}
		require.NoError(t, err)
		require.Equal(t, 2, b.Channels)
		all = append(all, b.Samples...)
	}
}

func readPCM16(t *testing.T, path string) []int16 {
	t.Helper()
	r, err := newTestFFmpeg().OpenPCM16(context.Background(), path)
	require.NoError(t, err)
	defer func() { require.NoError(t, r.Close()) }()
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	out := make([]int16, len(data)/2)
	require.NoError(t, binary.Read(bytes.NewReader(data),
		binary.LittleEndian, out))
	return out
}

func TestProbeFLAC24(t *testing.T) {
	path := sines(t, "in.flac", "5.1", "96000",
		"-c:a", "flac", "-sample_fmt", "s32", "-bits_per_raw_sample", "24")

	info, err := newTestFFmpeg().Probe(context.Background(), path)

	require.NoError(t, err)
	assert.Equal(t, "flac", info.Codec)
	assert.True(t, info.Lossless)
	assert.Equal(t, 96000, info.SampleRate)
	assert.Equal(t, 24, info.BitDepth)
	assert.Equal(t, 6, info.Channels)
	assert.Equal(t, "5.1", info.Layout)
	assert.Equal(t, int64(192000), info.TotalSamples)
}

func TestProbeMP3(t *testing.T) {
	path := sines(t, "in.mp3", "stereo", "44100", "-c:a", "libmp3lame")

	info, err := newTestFFmpeg().Probe(context.Background(), path)

	require.NoError(t, err)
	assert.False(t, info.Lossless)
	assert.Equal(t, 0, info.BitDepth)
	assert.Equal(t, 44100, info.SampleRate)
	assert.InDelta(t, 88200, info.TotalSamples, 3000)
}

// The float path must not change 16-bit samples: ffmpeg's int→float
// scale has to match dsp.Quantizer's, and aresample must be a no-op at
// equal rates. Plan.OnGrid relies on both.
func TestOpenStereo16IsBitExact(t *testing.T) {
	path := sines(t, "in.flac", "stereo", "44100",
		"-c:a", "flac", "-sample_fmt", "s16")
	want := readPCM16(t, path)
	spec := audio.DecodeSpec{SampleRate: 44100}

	r, err := newTestFFmpeg().Open(context.Background(), path, spec)

	require.NoError(t, err)
	floats := readFloats(t, r)
	got := make([]int16, len(floats))
	dsp.NewQuantizer(false, 0).Quantize(got, floats)
	require.Len(t, want, 2*88200)
	assert.Equal(t, want, got)
}

func TestOpenMonoIsDuplicated(t *testing.T) {
	path := sines(t, "in.flac", "mono", "44100",
		"-c:a", "flac", "-sample_fmt", "s16")
	mono := readPCM16(t, path)
	m, err := dsp.MixMatrix(1, "mono")
	require.NoError(t, err)
	spec := audio.DecodeSpec{Mix: m, SampleRate: 44100}

	r, err := newTestFFmpeg().Open(context.Background(), path, spec)

	require.NoError(t, err)
	floats := readFloats(t, r)
	got := make([]int16, len(floats))
	dsp.NewQuantizer(false, 0).Quantize(got, floats)
	want := make([]int16, 0, 2*len(mono))
	for _, s := range mono {
		want = append(want, s, s)
	}
	assert.Equal(t, want, got)
}

func TestOpenDownmixAndResample(t *testing.T) {
	path := sines(t, "in.flac", "5.1", "96000",
		"-c:a", "flac", "-sample_fmt", "s32", "-bits_per_raw_sample", "24")
	m, err := dsp.MixMatrix(6, "5.1")
	require.NoError(t, err)
	spec := audio.DecodeSpec{Mix: m, SampleRate: 44100}

	r, err := newTestFFmpeg().Open(context.Background(), path, spec)

	require.NoError(t, err)
	floats := readFloats(t, r)
	assert.InDelta(t, 88200, len(floats)/2, 2)
}

func TestOpenMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.flac")
	spec := audio.DecodeSpec{SampleRate: 44100}
	r, err := newTestFFmpeg().Open(context.Background(), path, spec)
	require.NoError(t, err)

	_, err = r.ReadBlock(context.Background())

	assert.ErrorContains(t, err, "No such file")
}

func TestExtract(t *testing.T) {
	audio := sines(t, "a.flac", "stereo", "44100", "-c:a", "flac")
	cover := gen(t, "c.png",
		"-f", "lavfi", "-i", "color=c=red:s=8x8", "-frames:v", "1")
	path := gen(t, "tagged.flac",
		"-i", audio, "-i", cover, "-map", "0:a", "-map", "1:v",
		"-c", "copy", "-disposition:v", "attached_pic",
		"-metadata", "title=Song", "-metadata", "artist=Singer",
		"-metadata", "track=3/12", "-metadata", "label=Indie")
	png, err := os.ReadFile(cover)
	require.NoError(t, err)

	tags, c, err := newTestFFmpeg().Extract(context.Background(), path)

	require.NoError(t, err)
	assert.Equal(t, "Song", tags.Title)
	assert.Equal(t, "Singer", tags.Artist)
	assert.Equal(t, 3, tags.TrackNumber)
	assert.Equal(t, []string{"Indie"}, tags.Extra["label"])
	require.NotNil(t, c)
	assert.Equal(t, "image/png", c.MIMEType)
	assert.Equal(t, png, c.Data)
}
