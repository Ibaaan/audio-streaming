package dsp

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGain(t *testing.T) {
	tests := []struct {
		tp   float64
		want float64
	}{
		{math.Inf(-1), 1},
		{-3, 1},
		{-1, 1},
		{0, 0.891251},
		{0.8, 0.812831},
		{3, 0.630957},
	}
	for _, tt := range tests {
		t.Run(strconv.FormatFloat(tt.tp, 'g', -1, 64), func(t *testing.T) {
			g := Gain(tt.tp)

			assert.InDelta(t, tt.want, g, 1e-6)
		})
	}
}

func TestApplyGain(t *testing.T) {
	buf := []float32{1, -0.5, 0}

	ApplyGain(buf, 0.5)

	assert.Equal(t, []float32{0.5, -0.25, 0}, buf)
}
