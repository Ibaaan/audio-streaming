package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
)

func TestGetAlbumBuildsOrderedTracklist(t *testing.T) {
	id, artist := uuid.New(), uuid.New()
	year := int16(2004)
	featured := "featured"
	repo := &fakeRepo{
		albums: []catalog.Album{{ID: id, Title: "Album", ReleaseYear: &year,
			AlbumType: catalog.AlbumTypeAlbum}},
		artists:      []catalog.Artist{{ID: artist, Name: "Artist"}},
		albumArtists: []catalog.AlbumArtist{{AlbumID: id, ArtistID: artist}},
	}
	for _, disc := range []int16{2, 1} {
		track := uuid.New()
		repo.tracks = append(repo.tracks, catalog.Track{ID: track, Title: "Song", DurationMs: 1000})
		repo.albumTracks = append(repo.albumTracks, catalog.AlbumTrack{
			AlbumID: id, TrackID: track, DiscNo: disc, TrackNo: 1,
		})
		repo.trackArtists = append(repo.trackArtists, catalog.TrackArtist{
			TrackID: track, ArtistID: artist, Role: &featured,
		})
	}

	got, err := NewLibrary(repo).GetAlbum(context.Background(), id, []string{"tracks"})

	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "Album", got.Title)
	assert.Equal(t, 2004, *got.ReleaseYear)
	assert.Equal(t, int64(2000), got.DurationMs)
	assert.Equal(t, 2, got.DiscCount)
	assert.Equal(t, 2, got.TrackCount)
	require.Len(t, got.Artists, 1)
	assert.Equal(t, artist, got.Artists[0].ID)
	require.NotNil(t, got.Tracks)
	require.Len(t, *got.Tracks, 2)
	assert.Equal(t, 1, (*got.Tracks)[0].DiscNumber)
	assert.Equal(t, 2, (*got.Tracks)[1].DiscNumber)
	assert.Equal(t, "featured", string((*got.Tracks)[0].Artists[0].Role))
}

func TestGetAlbumOmitsTracklist(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{albums: []catalog.Album{{ID: id}}}

	got, err := NewLibrary(repo).GetAlbum(context.Background(), id, nil)

	require.NoError(t, err)
	assert.Nil(t, got.Tracks)
	assert.NotNil(t, got.Artists)
	assert.Zero(t, got.TrackCount)
	assert.Zero(t, got.DiscCount)
	data, err := json.Marshal(got)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"tracks"`)
}

func TestGetAlbumIncludesEmptyTracklist(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{albums: []catalog.Album{{ID: id}}}

	got, err := NewLibrary(repo).GetAlbum(context.Background(), id, []string{"tracks"})

	require.NoError(t, err)
	require.NotNil(t, got.Tracks)
	data, err := json.Marshal(got)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"tracks":[]`)
}

func TestGetAlbumMissing(t *testing.T) {
	_, err := NewLibrary(&fakeRepo{}).GetAlbum(context.Background(), uuid.New(), nil)

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
