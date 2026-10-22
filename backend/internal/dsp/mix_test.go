package dsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const k = minus3dB

func TestMixMatrix(t *testing.T) {
	tests := []struct {
		layout string
		ch     int
		want   [][]float64
	}{
		{"mono", 1, [][]float64{{1}, {1}}},
		{"stereo", 2, [][]float64{{1, 0}, {0, 1}}},
		{"2.1", 3, [][]float64{{1, 0, 0}, {0, 1, 0}}},
		{"5.1", 6, [][]float64{
			{1, 0, k, 0, k, 0},
			{0, 1, k, 0, 0, k},
		}},
		{"5.1(side)", 6, [][]float64{
			{1, 0, k, 0, k, 0},
			{0, 1, k, 0, 0, k},
		}},
		{"7.1", 8, [][]float64{
			{1, 0, k, 0, k, 0, k, 0},
			{0, 1, k, 0, 0, k, 0, k},
		}},
		{"", 6, [][]float64{
			{1, 0, k, 0, k, 0},
			{0, 1, k, 0, 0, k},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.layout, func(t *testing.T) {
			m, err := MixMatrix(tt.ch, tt.layout)

			require.NoError(t, err)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestMixMatrixChannelMismatch(t *testing.T) {
	_, err := MixMatrix(2, "5.1")

	assert.Error(t, err)
}

func TestMixMatrixUnknownLayout(t *testing.T) {
	_, err := MixMatrix(3, "")

	assert.Error(t, err)
}
