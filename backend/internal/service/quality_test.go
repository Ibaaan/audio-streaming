package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/audio"
	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

func TestSaveQualityWithoutAlbum(t *testing.T) {
	rec := testRecord("Standalone")
	rec.Tags.Album = ""
	rec.SourceInfo.Codec = "aac"
	rec.SourceInfo.Lossless = false
	rec.SourceInfo.BitRate = 192000
	repo := &fakeRepo{}
	lib := NewLibrary(repo)

	track, err := lib.Save(context.Background(), rec, nil)

	require.NoError(t, err)
	assert.Equal(t, catalog.QualityOptimal, track.Quality)
	require.Len(t, repo.tracks, 1)
	assert.Equal(t, catalog.QualityOptimal, repo.tracks[0].Quality)
	got, err := lib.GetTrack(context.Background(), track.ID)
	require.NoError(t, err)
	assert.Equal(t, dto.QualityOptimal, got.Quality)
	data, err := json.Marshal(got)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"quality":"optimal"`)
	assert.NotContains(t, string(data), `"qualities"`)
}

func TestSaveUsesPipelineQuality(t *testing.T) {
	rec := testRecord("Low bitrate")
	rec.SourceInfo = audio.SourceInfo{
		Codec: "mp3", BitRate: 64000, SampleRate: 44100, Samples: 44100,
	}
	rec.Quality = audio.ClassifyQuality(rec.SourceInfo)
	repo := &fakeRepo{}

	track, err := NewLibrary(repo).Save(context.Background(), rec, nil)

	require.NoError(t, err)
	assert.Equal(t, catalog.QualityLow, track.Quality)
	require.Len(t, repo.albumTracks, 1)
	assert.Equal(t, catalog.QualityLow, repo.albumTracks[0].Quality)
}
