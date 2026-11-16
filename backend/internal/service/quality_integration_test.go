//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

func TestDBTrackQualityWithoutAlbum(t *testing.T) {
	lib, repo := newDBLibrary(t)
	rec := testRecord("High bitrate AAC")
	rec.Tags.Album = ""
	rec.SourceInfo.Codec = "aac"
	rec.SourceInfo.Lossless = false
	rec.SourceInfo.BitRate = 256000
	ctx := context.Background()

	saved, err := lib.Save(ctx, rec, nil)

	require.NoError(t, err)
	assert.Equal(t, catalog.QualityLossless, saved.Quality)
	stored, err := repo.GetTrack(ctx, saved.ID)
	require.NoError(t, err)
	assert.Equal(t, catalog.QualityLossless, stored.Quality)
	quality, err := lib.GetTrackQuality(ctx, saved.ID)
	require.NoError(t, err)
	assert.Equal(t, dto.QualityLossless, quality)
	found, _, err := repo.SearchTracks(ctx, "High bitrate AAC", 20, 0)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, catalog.QualityLossless, found[0].Quality)
}
