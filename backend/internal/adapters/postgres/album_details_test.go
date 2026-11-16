//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
)

func TestGetAlbumDetailsReadsNestedCredits(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	album := catalog.Album{ID: uuid.New(), Title: "Album", AlbumType: catalog.AlbumTypeAlbum}
	require.NoError(t, repo.AddAlbum(ctx, &album))
	artist := catalog.Artist{ID: uuid.New(), Name: "Artist"}
	require.NoError(t, repo.AddArtist(ctx, &artist))
	require.NoError(t, repo.AddAlbumArtist(ctx, catalog.AlbumArtist{
		AlbumID: album.ID, ArtistID: artist.ID,
	}))
	featured := "featured"
	for _, disc := range []int16{2, 1} {
		track := catalog.Track{ID: uuid.New(), Title: "Song", DurationMs: 1000}
		require.NoError(t, repo.AddTrack(ctx, &track))
		require.NoError(t, repo.AddTrackArtist(ctx, catalog.TrackArtist{
			TrackID: track.ID, ArtistID: artist.ID, Role: &featured,
		}))
		require.NoError(t, repo.AddAlbumTrack(ctx, catalog.AlbumTrack{
			AlbumID: album.ID, TrackID: track.ID, DiscNo: disc, TrackNo: 1, Quality: catalog.QualityOptimal,
		}))
	}

	got, err := repo.GetAlbumDetails(ctx, album.ID)

	require.NoError(t, err)
	assert.Equal(t, album.ID, got.ID)
	require.Len(t, got.Artists, 1)
	assert.Equal(t, artist.Name, got.Artists[0].Name)
	require.Len(t, got.Tracks, 2)
	assert.Equal(t, int16(1), got.Tracks[0].DiscNo)
	assert.Equal(t, int16(2), got.Tracks[1].DiscNo)
	assert.Equal(t, int32(1000), got.Tracks[0].DurationMs)
	require.Len(t, got.Tracks[0].Artists, 1)
	assert.Equal(t, artist.ID, got.Tracks[0].Artists[0].ID)
	assert.Equal(t, &featured, got.Tracks[0].Artists[0].Role)
}

func TestGetAlbumDetailsEmpty(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	album := catalog.Album{ID: uuid.New(), Title: "Empty", AlbumType: catalog.AlbumTypeAlbum}
	require.NoError(t, repo.AddAlbum(ctx, &album))

	got, err := repo.GetAlbumDetails(ctx, album.ID)

	require.NoError(t, err)
	assert.Empty(t, got.Tracks)
	assert.Empty(t, got.Artists)
}

func TestGetAlbumDetailsMissing(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.GetAlbumDetails(context.Background(), uuid.New())

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
