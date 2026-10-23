package pipeline

import (
	"audio-streaming/internal/audio"
	"fmt"
)

const (
	targetRate     = 44100
	targetBitDepth = 16
	targetChannels = 2
)

// Mix says how source channels become stereo.
type Mix int

const (
	MixNone    Mix = iota // already stereo
	MixDupMono            // mono copied to both channels
	MixDownmix            // multichannel folded down per ITU-R BS.775
)

type Plan struct {
	Passthrough     bool
	Mix             Mix
	Resample        bool
	OnGrid          bool
	SeekIntervalSec int
}

func MakePlan(src audio.SourceInfo) (Plan, error) {
	if src.SampleRate <= 0 {
		return Plan{}, fmt.Errorf("plan: invalid sample rate %d", src.SampleRate)
	}

	mix, err := func(channels int) (Mix, error) {
		switch {
		case channels == 1:
			return MixDupMono, nil
		case channels == targetChannels:
			return MixNone, nil
		case channels > targetChannels:
			return MixDownmix, nil
		default:
			return 0, fmt.Errorf("plan: invalid channel count %d", channels)
		}
	}(src.Channels)

	if err != nil {
		return Plan{}, err
	}

	p := Plan{
		Mix:      mix,
		Resample: src.SampleRate != targetRate,
	}
	p.OnGrid = src.Lossless &&
		src.BitDepth > 0 && src.BitDepth <= targetBitDepth &&
		!p.Resample && p.Mix != MixDownmix
	p.Passthrough = src.Lossless &&
		src.BitDepth == targetBitDepth &&
		src.Channels == targetChannels &&
		!p.Resample
	p.SeekIntervalSec = seekInterval(src.Samples, src.SampleRate)
	return p, nil
}

const (
	minSeekIntervalSeconds = 10
	maxSeekpoints          = 100
)

func seekInterval(totalSamples int64, sampleRate int) int {
	if totalSamples <= 0 || sampleRate <= 0 {
		return 0
	}

	needed := ceilDiv(totalSamples, int64(sampleRate)*maxSeekpoints)
	interval := max(minSeekIntervalSeconds, needed)

	return int(interval)
}

func ceilDiv(a, b int64) int64 { return (a + b - 1) / b }
