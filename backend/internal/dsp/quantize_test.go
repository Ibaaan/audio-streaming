package dsp

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuantizeOnGridIsExact(t *testing.T) {
	src := make([]float32, 1<<16)
	want := make([]int16, len(src))
	for i := range src {
		want[i] = int16(i + math.MinInt16)
		src[i] = float32(want[i]) / fullScale
	}
	got := make([]int16, len(src))

	NewQuantizer(false, 1).Quantize(got, src)

	assert.Equal(t, want, got)
}

func TestQuantizeClamps(t *testing.T) {
	src := []float32{1, 1.5, -1, -1.5}
	got := make([]int16, len(src))

	NewQuantizer(true, 1).Quantize(got, src)

	assert.Equal(t, []int16{32767, 32767, -32768, -32768}, got)
}

func TestQuantizeShortDst(t *testing.T) {
	got := make([]int16, 2)

	n := NewQuantizer(false, 1).Quantize(got, []float32{0, 0, 0})

	assert.Equal(t, 2, n)
}

func TestQuantizeDitherIsReproducible(t *testing.T) {
	src := make([]float32, 1000)
	for i := range src {
		src[i] = float32(math.Sin(float64(i) / 10))
	}
	a := make([]int16, len(src))
	b := make([]int16, len(src))
	NewQuantizer(true, 42).Quantize(a, src)

	NewQuantizer(true, 42).Quantize(b, src)

	assert.Equal(t, a, b)
}

func TestQuantizeDitherError(t *testing.T) {
	const n = 100_000
	src := make([]float32, n)
	for i := range src {
		src[i] = float32(math.Sin(float64(i)/7.3)) * 0.5
	}
	got := make([]int16, n)

	NewQuantizer(true, 7).Quantize(got, src)

	var sum, worst float64
	for i := range src {
		e := float64(got[i]) - float64(src[i])*fullScale
		sum += e
		worst = max(worst, math.Abs(e))
	}
	require.Positive(t, worst, "dither must change something")
	assert.LessOrEqual(t, worst, 1.5)
	assert.InDelta(t, 0, sum/n, 0.01)
}
