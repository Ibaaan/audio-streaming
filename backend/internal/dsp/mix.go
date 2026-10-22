// Package dsp holds pure audio-processing functions with no I/O.
package dsp

import (
	"fmt"
	"math"
)

// minus3dB is the ITU-R BS.775 weight for centre and surround channels.
const minus3dB = 1 / math.Sqrt2

// layouts lists channel order as ffmpeg/ffprobe report it.
var layouts = map[string][]string{
	"mono":      {"FC"},
	"stereo":    {"FL", "FR"},
	"2.1":       {"FL", "FR", "LFE"},
	"3.0":       {"FL", "FR", "FC"},
	"4.0":       {"FL", "FR", "FC", "BC"},
	"quad":      {"FL", "FR", "BL", "BR"},
	"5.0":       {"FL", "FR", "FC", "BL", "BR"},
	"5.0(side)": {"FL", "FR", "FC", "SL", "SR"},
	"5.1":       {"FL", "FR", "FC", "LFE", "BL", "BR"},
	"5.1(side)": {"FL", "FR", "FC", "LFE", "SL", "SR"},
	"7.1":       {"FL", "FR", "FC", "LFE", "BL", "BR", "SL", "SR"},
}

// defaultLayouts is used when the source reports no layout.
var defaultLayouts = map[int]string{
	1: "mono",
	2: "stereo",
	6: "5.1",
	8: "7.1",
}

// weights gives each channel's contribution to the left and right output.
var weights = map[string][2]float64{
	"FL":  {1, 0},
	"FR":  {0, 1},
	"FC":  {minus3dB, minus3dB},
	"BC":  {minus3dB, minus3dB},
	"LFE": {0, 0},
	"BL":  {minus3dB, 0},
	"SL":  {minus3dB, 0},
	"BR":  {0, minus3dB},
	"SR":  {0, minus3dB},
}

// MixMatrix returns a 2×ch matrix that folds the layout down to stereo:
// out[o] = Σ m[o][i]·in[i]. Mono is copied to both channels at full
// level. The result is not normalised, so the sum can exceed 1.0; the
// gain stage takes care of that.
func MixMatrix(ch int, layout string) ([][]float64, error) {
	if layout == "" {
		layout = defaultLayouts[ch]
	}
	names, ok := layouts[layout]
	if !ok {
		return nil, fmt.Errorf("dsp: unsupported channel layout %q (%d ch)",
			layout, ch)
	}
	if len(names) != ch {
		return nil, fmt.Errorf("dsp: layout %q has %d channels, got %d",
			layout, len(names), ch)
	}

	m := [][]float64{make([]float64, ch), make([]float64, ch)}
	if layout == "mono" {
		m[0][0], m[1][0] = 1, 1
		return m, nil
	}
	for i, name := range names {
		w := weights[name]
		m[0][i], m[1][i] = w[0], w[1]
	}
	return m, nil
}
