package pipeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/track"
)

var seekTests = []struct {
	length   time.Duration
	interval int
	points   int
}{
	{5 * time.Second, 10, 1},
	{3*time.Minute + 30*time.Second, 10, 21},
	{16*time.Minute + 40*time.Second, 10, 100},
	{20 * time.Minute, 12, 100},
	{3 * time.Hour, 108, 100},
}

const rate = 44100

func TestSeekInterval(t *testing.T) {
	for _, tt := range seekTests {
		t.Run(tt.length.String(), func(t *testing.T) {
			secs := int64(tt.length / time.Second)
			samples := secs * rate

			interval := seekInterval(samples, rate)

			assert.Equal(t, tt.interval, interval, "interval")
			assert.EqualValues(t, tt.points,
				ceilDiv(secs, int64(interval)), "points")
		})
	}
}

func source(codec string, lossless bool, rate, bits, ch int,
) track.SourceInfo {
	return track.SourceInfo{
		Codec:        codec,
		Lossless:     lossless,
		SampleRate:   rate,
		BitDepth:     bits,
		Channels:     ch,
		TotalSamples: int64(rate) * 60,
	}
}

var planTests = []struct {
	name string
	src  track.SourceInfo
	want Plan
}{
	{"flac 16/44.1 stereo", source("flac", true, 44100, 16, 2),
		Plan{Passthrough: true, Mix: MixNone, OnGrid: true,
			SeekIntervalSec: 10}},
	{"flac 16/44.1 mono", source("flac", true, 44100, 16, 1),
		Plan{Mix: MixDupMono, OnGrid: true, SeekIntervalSec: 10}},
	{"mp3 44.1 stereo", source("mp3", false, 44100, 0, 2),
		Plan{Mix: MixNone, SeekIntervalSec: 10}},
	{"mp3 44.1 mono", source("mp3", false, 44100, 0, 1),
		Plan{Mix: MixDupMono, SeekIntervalSec: 10}},
	{"wav 8/44.1 stereo", source("pcm_u8", true, 44100, 8, 2),
		Plan{Mix: MixNone, OnGrid: true, SeekIntervalSec: 10}},
	{"flac 24/44.1 stereo", source("flac", true, 44100, 24, 2),
		Plan{Mix: MixNone, SeekIntervalSec: 10}},
	{"flac 16/48 stereo", source("flac", true, 48000, 16, 2),
		Plan{Mix: MixNone, Resample: true, SeekIntervalSec: 10}},
	{"flac 24/96 5.1", source("flac", true, 96000, 24, 6),
		Plan{Mix: MixDownmix, Resample: true, SeekIntervalSec: 10}},
	{"flac 16/44.1 5.1", source("flac", true, 44100, 16, 6),
		Plan{Mix: MixDownmix, SeekIntervalSec: 10}},
}

func TestMakePlan(t *testing.T) {
	for _, tt := range planTests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := MakePlan(tt.src)

			require.NoError(t, err)
			assert.Equal(t, tt.want, p)
		})
	}
}

func TestMakePlanNoChannels(t *testing.T) {
	src := source("flac", true, 44100, 16, 0)

	_, err := MakePlan(src)

	assert.Error(t, err)
}

func TestMakePlanNoSampleRate(t *testing.T) {
	src := source("flac", true, 0, 16, 2)

	_, err := MakePlan(src)

	assert.Error(t, err)
}
