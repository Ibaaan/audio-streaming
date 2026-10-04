package dsp

import "math"

const (
	tpPhases = 4  // oversampling factor
	tpTaps   = 12 // taps per phase
)

// tpCoefs is the 4× interpolation filter from ITU-R BS.1770-4 Annex 2,
// split into phases. Phases 2 and 3 mirror 1 and 0.
var tpCoefs = [tpPhases][tpTaps]float64{
	{
		0.0017089843750, 0.0109863281250, -0.0196533203125,
		0.0332031250000, -0.0594482421875, 0.1373291015625,
		0.9721679687500, -0.1022949218750, 0.0476074218750,
		-0.0266113281250, 0.0148925781250, -0.0083007812500,
	},
	{
		-0.0291748046875, 0.0292968750000, -0.0517578125000,
		0.0891113281250, -0.1665039062500, 0.4650878906250,
		0.7797851562500, -0.2003173828125, 0.1015625000000,
		-0.0582275390625, 0.0330810546875, -0.0189208984375,
	},
	{
		-0.0189208984375, 0.0330810546875, -0.0582275390625,
		0.1015625000000, -0.2003173828125, 0.7797851562500,
		0.4650878906250, -0.1665039062500, 0.0891113281250,
		-0.0517578125000, 0.0292968750000, -0.0291748046875,
	},
	{
		-0.0083007812500, 0.0148925781250, -0.0266113281250,
		0.0476074218750, -0.1022949218750, 0.9721679687500,
		0.1373291015625, -0.0594482421875, 0.0332031250000,
		-0.0196533203125, 0.0109863281250, 0.0017089843750,
	},
}

// TruePeak measures the true peak of interleaved audio: the highest
// level the waveform reaches between samples, estimated by 4×
// oversampling (ITU-R BS.1770-4 Annex 2). Feed it blocks with Write;
// how the audio is split into blocks does not change the result.
//
// At 44.1 kHz, 4× can read up to about 0.5 dB low on worst-case
// signals; the −1 dBTP ceiling covers that.
type TruePeak struct {
	channels int
	// hist holds each channel's last tpTaps samples, newest first.
	hist [][tpTaps]float64
	peak float64
}

func NewTruePeak(channels int) *TruePeak {
	return &TruePeak{
		channels: channels,
		hist:     make([][tpTaps]float64, channels),
	}
}

// Write measures interleaved samples. len(samples) must be a multiple
// of the channel count.
func (m *TruePeak) Write(samples []float32) {
	for i, s := range samples {
		h := &m.hist[i%m.channels]
		m.peak = max(m.peak, push(h, float64(s)))
	}
}

// Peak returns the linear true peak so far (1.0 = full scale). The
// filter lags by a few samples, so Peak runs the history out with
// silence on a copy; the meter itself is unchanged.
func (m *TruePeak) Peak() float64 {
	peak := m.peak
	for _, h := range m.hist {
		for range tpTaps - 1 {
			peak = max(peak, push(&h, 0))
		}
	}
	return peak
}

// DBTP returns Peak in dB relative to full scale; -Inf for silence.
func (m *TruePeak) DBTP() float64 {
	return 20 * math.Log10(m.Peak())
}

// push adds x to the history and returns the largest magnitude among
// x and the four interpolated points that end at it.
func push(h *[tpTaps]float64, x float64) float64 {
	copy(h[1:], h[:tpTaps-1])
	h[0] = x

	peak := math.Abs(x)
	for p := range tpCoefs {
		var y float64
		for k, c := range tpCoefs[p] {
			y += c * h[k]
		}
		peak = max(peak, math.Abs(y))
	}
	return peak
}
