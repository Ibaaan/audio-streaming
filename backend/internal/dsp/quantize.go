package dsp

import (
	"math"
	"math/rand/v2"
)

// fullScale maps float [-1, 1) onto int16. The same factor is used on
// decode, so a 16-bit sample survives the round trip unchanged.
const fullScale = 32768

// Quantizer converts float samples to int16, optionally with TPDF dither.
type Quantizer struct {
	dither bool
	rng    *rand.Rand
}

// NewQuantizer returns a quantizer. The seed makes the dither, and so the
// output MD5, reproducible for a given track.
func NewQuantizer(dither bool, seed uint64) *Quantizer {
	return &Quantizer{
		dither: dither,
		rng:    rand.New(rand.NewPCG(seed, seed)),
	}
}

// Quantize converts min(len(dst), len(src)) samples and returns how many
// it wrote. Without dither, samples that sit on the 16-bit grid come out
// exact. Out-of-range values are clamped.
func (q *Quantizer) Quantize(dst []int16, src []float32) int {
	n := min(len(dst), len(src))
	for i := range n {
		v := float64(src[i]) * fullScale
		if q.dither {
			// Two uniform [0, 1) draws give a triangle on (-1, 1) LSB.
			v += q.rng.Float64() - q.rng.Float64()
		}
		dst[i] = int16(max(math.MinInt16, min(math.MaxInt16, math.Round(v))))
	}
	return n
}
