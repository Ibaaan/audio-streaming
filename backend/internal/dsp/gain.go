package dsp

import "math"

// CeilingDBTP is the highest true peak a track may have after gain.
const CeilingDBTP = -1.0

// Gain returns the linear factor that brings a true peak of tpDB down to
// CeilingDBTP. Tracks already below the ceiling get 1 (never boosted).
// Silence (tpDB = -Inf) also gets 1.
func Gain(tpDB float64) float64 {
	return min(1, math.Pow(10, (CeilingDBTP-tpDB)/20))
}

// ApplyGain multiplies every sample in buf by g in place.
func ApplyGain(buf []float32, g float32) {
	for i := range buf {
		buf[i] *= g
	}
}
