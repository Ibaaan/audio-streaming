package ffmpeg

import (
	"audio-streaming/internal/audio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// probeOutput is the part of `ffprobe -print_format json` we use.
type probeOutput struct {
	Streams []probeStream `json:"streams"`
	Format  struct {
		Duration string            `json:"duration"`
		Tags     map[string]string `json:"tags"`
	} `json:"format"`
}

type probeStream struct {
	Index            int    `json:"index"`
	CodecType        string `json:"codec_type"`
	CodecName        string `json:"codec_name"`
	SampleFmt        string `json:"sample_fmt"`
	SampleRate       string `json:"sample_rate"`
	Channels         int    `json:"channels"`
	ChannelLayout    string `json:"channel_layout"`
	BitsPerSample    int    `json:"bits_per_sample"`
	BitsPerRawSample string `json:"bits_per_raw_sample"`
	TimeBase         string `json:"time_base"`
	DurationTS       int64  `json:"duration_ts"`
	Duration         string `json:"duration"`
	Disposition      struct {
		AttachedPic int `json:"attached_pic"`
	} `json:"disposition"`
	Tags map[string]string `json:"tags"`
}

// losslessCodecs are decoded to exact integers. pcm_* is handled
// separately.
var losslessCodecs = map[string]bool{
	"flac": true, "alac": true, "wavpack": true, "ape": true,
	"tta": true, "tak": true, "truehd": true, "mlp": true,
}

// Probe describes the first audio stream of path.
func (f *FFmpeg) Probe(ctx context.Context, path string) (
	audio.SourceInfo, error,
) {
	p, err := f.probe(ctx, path)
	if err != nil {
		return audio.SourceInfo{}, err
	}
	return sourceInfo(p)
}

func (f *FFmpeg) probe(ctx context.Context, path string) (probeOutput, error) {
	args := []string{
		"-v", "error", "-protocol_whitelist", "file",
		"-print_format", "json", "-show_format", "-show_streams",
		"-i", "file:" + path,
	}
	out, err := exec.CommandContext(ctx, f.ffprobe, args...).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return probeOutput{}, fmt.Errorf("ffprobe %s: %w: %s",
				path, err, bytes.TrimSpace(ee.Stderr))
		}
		return probeOutput{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}
	return parseProbe(out)
}

func parseProbe(data []byte) (probeOutput, error) {
	var p probeOutput
	if err := json.Unmarshal(data, &p); err != nil {
		return probeOutput{}, fmt.Errorf("parsing ffprobe output: %w", err)
	}
	return p, nil
}

func (p probeOutput) audio() (probeStream, bool) {
	for _, s := range p.Streams {
		if s.CodecType == "audio" {
			return s, true
		}
	}
	return probeStream{}, false
}

func sourceInfo(p probeOutput) (audio.SourceInfo, error) {
	s, ok := p.audio()
	if !ok {
		return audio.SourceInfo{}, errors.New("ffprobe: no audio stream")
	}
	rate, err := strconv.Atoi(s.SampleRate)
	if err != nil {
		return audio.SourceInfo{}, fmt.Errorf("ffprobe: sample rate %q: %w",
			s.SampleRate, err)
	}

	info := audio.SourceInfo{
		Codec:        s.CodecName,
		Lossless:     isLossless(s.CodecName),
		SampleRate:   rate,
		Channels:     s.Channels,
		Layout:       layout(s.ChannelLayout),
		TotalSamples: totalSamples(s, p.Format.Duration, rate),
	}
	if info.Lossless {
		info.BitDepth = bitDepth(s)
	}
	return info, nil
}

func isLossless(codec string) bool {
	return losslessCodecs[codec] || strings.HasPrefix(codec, "pcm_")
}

// bitDepth returns the integer sample size, or 0 for float samples.
func bitDepth(s probeStream) int {
	if n, _ := strconv.Atoi(s.BitsPerRawSample); n > 0 {
		return n
	}
	sampleFmt := strings.TrimSuffix(s.SampleFmt, "p") // planar → packed
	if sampleFmt == "flt" || sampleFmt == "dbl" {
		return 0
	}
	if s.BitsPerSample > 0 {
		return s.BitsPerSample
	}
	switch sampleFmt {
	case "u8":
		return 8
	case "s16":
		return 16
	case "s32":
		return 32
	}
	return 0
}

// layout drops the placeholder ffprobe reports when the channels have
// no assigned speakers, e.g. "6 channels".
func layout(l string) string {
	if l == "unknown" || strings.HasSuffix(l, " channels") {
		return ""
	}
	return l
}

// totalSamples converts the stream duration to samples per channel.
func totalSamples(s probeStream, formatDur string, rate int) int64 {
	if s.DurationTS > 0 {
		var num, den int64
		_, err := fmt.Sscanf(s.TimeBase, "%d/%d", &num, &den)
		if err == nil && den > 0 {
			return s.DurationTS * num * int64(rate) / den
		}
	}
	dur := s.Duration
	if dur == "" {
		dur = formatDur
	}
	sec, err := strconv.ParseFloat(dur, 64)
	if err != nil {
		return 0
	}
	return int64(math.Round(sec * float64(rate)))
}
