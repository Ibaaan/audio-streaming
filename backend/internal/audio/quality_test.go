package audio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyQuality(t *testing.T) {
	for _, tc := range []struct {
		name, codec string
		bitrate     int64
		want        Quality
	}{
		{"flac", "flac", 0, QualityLossless},
		{"alac", "alac", 0, QualityLossless},
		{"wav", "pcm_s16le", 64000, QualityLossless},
		{"aac high", "aac", 256000, QualityLossless},
		{"aac middle", "aac", 192000, QualityOptimal},
		{"aac below middle", "aac", 191999, QualityLow},
		{"aac below high", "aac", 255999, QualityOptimal},
		{"mp3 high", "mp3", 320000, QualityLossless},
		{"mp3 below high", "mp3", 319999, QualityOptimal},
		{"mp3 middle", "mp3", 192000, QualityOptimal},
		{"mp3 low", "mp3", 64000, QualityLow},
		{"opus high", "opus", 192000, QualityLossless},
		{"opus middle", "opus", 128000, QualityOptimal},
		{"opus low", "opus", 64000, QualityLow},
		{"vorbis high", "vorbis", 256000, QualityLossless},
		{"vorbis middle", "vorbis", 192000, QualityOptimal},
		{"wma high", "wmav2", 256000, QualityLossless},
		{"unknown bitrate", "mp3", 0, QualityLow},
		{"negative bitrate", "mp3", -1, QualityLow},
		{"other codec high", "ac3", 384000, QualityLossless},
		{"other codec middle", "ac3", 192000, QualityOptimal},
		{"case normalized", " AAC ", 256000, QualityLossless},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyQuality(SourceInfo{Codec: tc.codec, BitRate: tc.bitrate})

			assert.Equal(t, tc.want, got)
		})
	}
}
