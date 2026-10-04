package dsp

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// quarterSine is a full-scale sine at fs/4 shifted by 45°: every sample
// is ±0.7071 (−3 dB), while the wave itself peaks at 1.0 in between.
func quarterSine(frames int) []float32 {
	s := make([]float32, frames)
	for n := range s {
		s[n] = float32(math.Sin(math.Pi/2*float64(n) + math.Pi/4))
	}
	return s
}

// noise is a deterministic test signal with no simple structure.
func noise(n int) []float32 {
	s := make([]float32, n)
	for i := range s {
		s[i] = float32(math.Sin(float64(i)*0.37) * math.Cos(float64(i)*1.91))
	}
	return s
}

func TestTruePeakSilence(t *testing.T) {
	m := NewTruePeak(2)
	m.Write(make([]float32, 1000))

	db := m.DBTP()

	assert.True(t, math.IsInf(db, -1))
}

func TestTruePeakFindsInterSamplePeak(t *testing.T) {
	m := NewTruePeak(1)
	m.Write(quarterSine(400))

	db := m.DBTP()

	assert.InDelta(t, 0, db, 0.1, "sample peak is −3 dB")
}

func TestTruePeakNotBelowSamplePeak(t *testing.T) {
	m := NewTruePeak(1)
	m.Write([]float32{0, 0, -0.9, 0, 0})

	peak := m.Peak()

	assert.GreaterOrEqual(t, peak, float64(float32(0.9)))
}

// The peak between the last two samples is only seen once the filter's
// lag is run out with silence.
func TestTruePeakPeakAtEnd(t *testing.T) {
	s := make([]float32, 22)
	s[20], s[21] = 0.7071, 0.7071
	m := NewTruePeak(1)
	m.Write(s)

	peak := m.Peak()

	assert.Greater(t, peak, 0.8)
}

func TestTruePeakBlockSizeDoesNotMatter(t *testing.T) {
	s := noise(2 * 1000)
	whole := NewTruePeak(2)
	whole.Write(s)
	chunked := NewTruePeak(2)
	for i := 0; i < len(s); i += 14 {
		chunked.Write(s[i:min(i+14, len(s))])
		chunked.Peak()
	}

	got := chunked.Peak()

	assert.Equal(t, whole.Peak(), got)
}

func TestTruePeakChannelsAreSeparate(t *testing.T) {
	sine := quarterSine(400)
	stereo := make([]float32, 2*len(sine))
	for i, v := range sine {
		stereo[2*i+1] = v
	}
	mono := NewTruePeak(1)
	mono.Write(sine)
	m := NewTruePeak(2)
	m.Write(stereo)

	got := m.Peak()

	assert.Equal(t, mono.Peak(), got)
}

// BenchmarkTruePeak measures one second of 44.1 kHz stereo per op.
// Run: go test -run '^$' -bench TruePeak ./internal/dsp/
func BenchmarkTruePeak(b *testing.B) {
	const rate = 44100
	s := noise(2 * rate)
	m := NewTruePeak(2)
	b.SetBytes(int64(4 * len(s)))

	for b.Loop() {
		m.Write(s)
	}

	perSecond := b.Elapsed().Seconds() / float64(b.N)
	b.ReportMetric(1/perSecond, "x_realtime")
}
