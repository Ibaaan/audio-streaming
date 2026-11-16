package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
)

func TestListArtistAlbumsOffset(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{artists: []catalog.Artist{{ID: id}}}
	for year := 2000; year < 2004; year++ {
		album := uuid.New()
		y := int16(year)
		repo.albums = append(repo.albums, catalog.Album{ID: album, ReleaseYear: &y})
		repo.albumArtists = append(repo.albumArtists, catalog.AlbumArtist{AlbumID: album, ArtistID: id})
	}

	got, err := NewLibrary(repo).ListArtistAlbums(context.Background(), id, 2, 1)

	require.NoError(t, err)
	assert.Equal(t, 4, got.TotalSize)
	require.Len(t, got.Albums, 2)
	assert.Equal(t, 2002, *got.Albums[0].ReleaseYear)
	assert.Equal(t, 2001, *got.Albums[1].ReleaseYear)
	require.NotNil(t, got.NextOffset)
	assert.Equal(t, 3, *got.NextOffset)
}

func TestListArtistAlbumsPastEnd(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{artists: []catalog.Artist{{ID: id}}}

	got, err := NewLibrary(repo).ListArtistAlbums(context.Background(), id, 20, 100)

	require.NoError(t, err)
	assert.NotNil(t, got.Albums)
	assert.Empty(t, got.Albums)
	assert.Nil(t, got.NextOffset)
	assert.Zero(t, got.TotalSize)
}

func TestListArtistAlbumsMissingArtist(t *testing.T) {
	_, err := NewLibrary(&fakeRepo{}).ListArtistAlbums(context.Background(), uuid.New(), 20, 0)

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
