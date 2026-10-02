package pipeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
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

func TestSeekpointCount(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.length.String(), func(t *testing.T) {
			secs := int64(tt.length / time.Second)
			samples := secs * rate

			interval := seekpointCount(samples, rate)

			assert.Equal(t, tt.interval, interval, "interval")
			assert.EqualValues(t, tt.points,
				ceilDiv(secs, int64(interval)), "points")
		})
	}
}
