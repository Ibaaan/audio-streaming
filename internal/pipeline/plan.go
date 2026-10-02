package pipeline

import "audio-streaming/internal/track"

type Plan struct {
	Passthrough           bool
	Downmix               bool
	Resample              bool
	SourceRate            int
	Dither                bool
	SeekpointsDurationSec int
}

func MakePlan(src track.SourceInfo) Plan {
	p := Plan{
		SourceRate: src.SampleRate,
		Downmix:  src.Channels > 2,
		Resample: src.SampleRate != 44100,
	}
	p.Dither = p.Downmix || p.Resample || src.BitDepth > 16
	p.Passthrough = !p.Dither
	p.SeekpointsDurationSec = seekpointCount(src.TotalSamples, src.SampleRate)
	return p
}

const (
	minSeekIntervalSeconds = 10
	maxSeekpoints          = 100
)

func seekpointCount(totalSamples int64, sampleRate int) int {
	if totalSamples <= 0 || sampleRate <= 0 {
		return 0
	}

	needed := ceilDiv(totalSamples, int64(sampleRate)*maxSeekpoints)
	interval := max(minSeekIntervalSeconds, needed)

	return int(interval)
}

func ceilDiv(a, b int64) int64 { return (a + b - 1) / b }
