package ffmpeg

import (
	"audio-streaming/internal/audio"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTags(t *testing.T) {
	p, err := parseProbe([]byte(`{
		"format": {"tags": {
			"TITLE": "Song", "track": "3/12", "ALBUMARTIST": "Band",
			"DATE": "2001", "LABEL": "Indie"
		}},
		"streams": [{"codec_type": "audio", "tags": {
			"artist": "Singer", "title": "ignored"
		}}]
	}`))
	require.NoError(t, err)

	got := parseTags(p)

	assert.Equal(t, audio.Tags{
		Title:       "Song",
		Artist:      "Singer",
		AlbumArtist: "Band",
		TrackNumber: 3,
		Date:        "2001",
		Extra:       map[string][]string{"label": {"Indie"}},
	}, got)
}

func TestLeadingInt(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"3/12", 3},
		{" 07", 7},
		{"12", 12},
		{"", 0},
		{"A1", 0},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := leadingInt(tt.in)

			assert.Equal(t, tt.want, got)
		})
	}
}
