package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/track"
)

func TestSourceInfo(t *testing.T) {
	tests := []struct {
		name string
		json string
		want track.SourceInfo
	}{
		{
			"flac 24/96 5.1",
			`{"streams":[{"codec_type":"audio","codec_name":"flac",
			"sample_fmt":"s32","sample_rate":"96000","channels":6,
			"channel_layout":"5.1(side)","bits_per_raw_sample":"24",
			"time_base":"1/96000","duration_ts":192000}]}`,
			track.SourceInfo{
				Codec: "flac", Lossless: true, SampleRate: 96000,
				BitDepth: 24, Channels: 6, Layout: "5.1(side)",
				TotalSamples: 192000,
			},
		},
		{
			"mp3 has no bit depth",
			`{"streams":[{"codec_type":"audio","codec_name":"mp3",
			"sample_fmt":"fltp","sample_rate":"44100","channels":2,
			"channel_layout":"stereo","time_base":"1/14112000",
			"duration_ts":28224000}]}`,
			track.SourceInfo{
				Codec: "mp3", SampleRate: 44100, Channels: 2,
				Layout: "stereo", TotalSamples: 88200,
			},
		},
		{
			"wav s16 without raw bits, duration from format",
			`{"streams":[{"codec_type":"audio","codec_name":"pcm_s16le",
			"sample_fmt":"s16","sample_rate":"44100","channels":1,
			"channel_layout":"mono","bits_per_sample":16}],
			"format":{"duration":"1.500000"}}`,
			track.SourceInfo{
				Codec: "pcm_s16le", Lossless: true, SampleRate: 44100,
				BitDepth: 16, Channels: 1, Layout: "mono",
				TotalSamples: 66150,
			},
		},
		{
			"float wav has no bit depth",
			`{"streams":[{"codec_type":"audio","codec_name":"pcm_f32le",
			"sample_fmt":"flt","sample_rate":"48000","channels":2,
			"channel_layout":"stereo","bits_per_sample":32}]}`,
			track.SourceInfo{
				Codec: "pcm_f32le", Lossless: true, SampleRate: 48000,
				Channels: 2, Layout: "stereo",
			},
		},
		{
			"unassigned layout, cover stream first",
			`{"streams":[{"codec_type":"video","codec_name":"png"},
			{"codec_type":"audio","codec_name":"pcm_u8",
			"sample_fmt":"u8","sample_rate":"8000","channels":6,
			"channel_layout":"6 channels"}]}`,
			track.SourceInfo{
				Codec: "pcm_u8", Lossless: true, SampleRate: 8000,
				BitDepth: 8, Channels: 6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parseProbe([]byte(tt.json))
			require.NoError(t, err)

			got, err := sourceInfo(p)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSourceInfoNoAudio(t *testing.T) {
	p, err := parseProbe([]byte(
		`{"streams":[{"codec_type":"video","codec_name":"png"}]}`))
	require.NoError(t, err)

	_, err = sourceInfo(p)

	assert.Error(t, err)
}
