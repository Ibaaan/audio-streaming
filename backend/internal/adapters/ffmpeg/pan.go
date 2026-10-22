package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
)

// PanFilter renders a mix matrix (see dsp.MixMatrix) as an ffmpeg pan
// filter. Channels are addressed by index (c0, c1, ...) and zero
// weights are left out. "=" keeps the weights as given; "<" would
// rescale them to sum to 1.
func PanFilter(m [][]float64) string {
	var b strings.Builder
	b.WriteString("pan=stereo")
	for o, row := range m {
		fmt.Fprintf(&b, "|c%d=", o)
		first := true
		for i, w := range row {
			if w == 0 {
				continue
			}
			if !first {
				b.WriteByte('+')
			}
			first = false
			b.WriteString(strconv.FormatFloat(w, 'g', -1, 64))
			fmt.Fprintf(&b, "*c%d", i)
		}
	}
	return b.String()
}
