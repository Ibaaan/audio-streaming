package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

func TestGetArtistDetails(t *testing.T) {
	id := uuid.New()
	created := time.Now().UTC()
	repo := &fakeRepo{artists: []catalog.Artist{
		{ID: id, Name: "Artist", CreatedAt: created, UpdatedAt: created},
	}}

	got, err := NewLibrary(repo).GetArtist(context.Background(), id, nil)

	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "Artist", got.DisplayName)
	assert.Equal(t, created, got.CreateTime)
	assert.Equal(t, dto.ArtistStats{}, got.Stats)
	assert.Nil(t, got.Albums)
	assert.Nil(t, got.Tracks)
	data, err := json.Marshal(got)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"albums"`)
	assert.NotContains(t, string(data), `"tracks"`)
}

func TestGetArtistIncludesFirstPages(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{artists: []catalog.Artist{{ID: id, Name: "Artist"}}}
	for year := 2000; year < 2021; year++ {
		album, track := uuid.New(), uuid.New()
		y := int16(year)
		repo.albums = append(repo.albums, catalog.Album{ID: album, Title: "Album",
			AlbumType: catalog.AlbumTypeAlbum, ReleaseYear: &y})
		repo.tracks = append(repo.tracks, catalog.Track{ID: track})
		repo.albumArtists = append(repo.albumArtists, catalog.AlbumArtist{
			AlbumID: album, ArtistID: id,
		})
		repo.trackArtists = append(repo.trackArtists, catalog.TrackArtist{
			TrackID: track, ArtistID: id,
		})
		repo.albumTracks = append(repo.albumTracks, catalog.AlbumTrack{
			AlbumID: album, TrackID: track, DiscNo: 1, TrackNo: 1,
		})
	}

	got, err := NewLibrary(repo).GetArtist(context.Background(), id, []string{"albums", "tracks"})

	require.NoError(t, err)
	assert.Equal(t, dto.ArtistStats{AlbumCount: 21, TrackCount: 21}, got.Stats)
	require.NotNil(t, got.Albums)
	require.Len(t, got.Albums.Albums, 20)
	assert.Equal(t, 21, got.Albums.TotalSize)
	require.NotNil(t, got.Albums.NextOffset)
	assert.Equal(t, 20, *got.Albums.NextOffset)
	assert.Equal(t, 2020, *got.Albums.Albums[0].ReleaseYear)
	assert.Equal(t, 1, got.Albums.Albums[0].TrackCount)
	assert.Equal(t, id, got.Albums.Albums[0].Artists[0].ID)
	require.NotNil(t, got.Tracks)
	assert.Len(t, got.Tracks.Tracks, 20)
	require.NotNil(t, got.Tracks.NextOffset)
	assert.Equal(t, 20, *got.Tracks.NextOffset)
}

func TestGetArtistMissing(t *testing.T) {
	_, err := NewLibrary(&fakeRepo{}).GetArtist(context.Background(), uuid.New(), nil)

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}

func TestGetArtistIncludesEmptyCollections(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{artists: []catalog.Artist{{ID: id}}}

	got, err := NewLibrary(repo).GetArtist(context.Background(), id, []string{"albums", "tracks"})

	require.NoError(t, err)
	require.NotNil(t, got.Albums)
	assert.NotNil(t, got.Albums.Albums)
	assert.Empty(t, got.Albums.Albums)
	require.NotNil(t, got.Tracks)
	assert.NotNil(t, got.Tracks.Tracks)
	assert.Empty(t, got.Tracks.Tracks)
}
