package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

func TestListArtistTracksIncludesFeaturedCreditsAndOffsets(t *testing.T) {
	artist, main := uuid.New(), uuid.New()
	track, album := uuid.New(), uuid.New()
	year := int16(2020)
	featured := "featured"
	repo := &fakeRepo{
		artists: []catalog.Artist{{ID: artist, Name: "Guest"}, {ID: main, Name: "Main"}},
		tracks:  []catalog.Track{{ID: track, Title: "Collaboration"}},
		trackArtists: []catalog.TrackArtist{
			{TrackID: track, ArtistID: main},
			{TrackID: track, ArtistID: artist, Position: 1, Role: &featured},
		},
		albums:      []catalog.Album{{ID: album, ReleaseYear: &year}},
		albumTracks: []catalog.AlbumTrack{{TrackID: track, AlbumID: album, DiscNo: 1, TrackNo: 3}},
	}
	lib := NewLibrary(repo)

	got, err := lib.ListArtistTracks(context.Background(), artist, 1, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, got.TotalSize)
	require.Len(t, got.Tracks, 1)
	assert.Equal(t, dto.ArtistRoleFeatured, got.Tracks[0].ArtistRole)
	assert.Equal(t, track, got.Tracks[0].ID)
	assert.Len(t, got.Tracks[0].Artists, 2)
	require.NotNil(t, got.Tracks[0].Album)
	assert.Equal(t, album, got.Tracks[0].Album.ID)
	assert.Equal(t, 3, *got.Tracks[0].TrackNumber)
	assert.Nil(t, got.NextOffset)
}

func TestListArtistTracksPagination(t *testing.T) {
	artist := uuid.New()
	repo := &fakeRepo{artists: []catalog.Artist{{ID: artist}}}
	for i := 0; i < 3; i++ {
		id := uuid.New()
		repo.tracks = append(repo.tracks, catalog.Track{ID: id})
		repo.trackArtists = append(repo.trackArtists, catalog.TrackArtist{TrackID: id, ArtistID: artist})
	}
	lib := NewLibrary(repo)

	first, err := lib.ListArtistTracks(context.Background(), artist, 2, 0)

	require.NoError(t, err)
	require.NotNil(t, first.NextOffset)
	assert.Equal(t, 2, *first.NextOffset)
	assert.Equal(t, 3, first.TotalSize)
	require.Len(t, first.Tracks, 2)
	last, err := lib.ListArtistTracks(context.Background(), artist, 2, *first.NextOffset)
	require.NoError(t, err)
	require.Len(t, last.Tracks, 1)
	assert.Nil(t, last.NextOffset)
	assert.NotEqual(t, first.Tracks[0].ID, last.Tracks[0].ID)
	assert.NotEqual(t, first.Tracks[1].ID, last.Tracks[0].ID)
	past, err := lib.ListArtistTracks(context.Background(), artist, 2, 100)
	require.NoError(t, err)
	assert.NotNil(t, past.Tracks)
	assert.Empty(t, past.Tracks)
	assert.Nil(t, past.NextOffset)
}

func TestListArtistTracksMissingArtist(t *testing.T) {
	lib := NewLibrary(&fakeRepo{})

	_, err := lib.ListArtistTracks(context.Background(), uuid.New(), 20, 0)

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
