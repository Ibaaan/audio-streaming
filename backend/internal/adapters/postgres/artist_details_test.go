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

func TestArtistDetailReads(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	artist := catalog.Artist{ID: uuid.New(), Name: "Artist"}
	require.NoError(t, repo.AddArtist(ctx, &artist))
	for _, year := range []int16{2000, 2020} {
		album := catalog.Album{ID: uuid.New(), Title: "Album", AlbumType: catalog.AlbumTypeAlbum,
			ReleaseYear: &year}
		require.NoError(t, repo.AddAlbum(ctx, &album))
		require.NoError(t, repo.AddAlbumArtist(ctx, catalog.AlbumArtist{
			AlbumID: album.ID, ArtistID: artist.ID,
		}))
		track := catalog.Track{ID: uuid.New(), Title: "Song", DurationMs: 1000}
		require.NoError(t, repo.AddTrack(ctx, &track))
		require.NoError(t, repo.AddTrackArtist(ctx, catalog.TrackArtist{
			TrackID: track.ID, ArtistID: artist.ID,
		}))
		require.NoError(t, repo.AddAlbumTrack(ctx, catalog.AlbumTrack{
			AlbumID: album.ID, TrackID: track.ID, DiscNo: 1, TrackNo: 1, Quality: catalog.QualityOptimal,
		}))
	}

	stats, err := repo.GetArtistStats(ctx, artist.ID)

	require.NoError(t, err)
	assert.Equal(t, catalog.ArtistStats{AlbumCount: 2, TrackCount: 2}, stats)
}

func TestListArtistAlbumsReadsCredits(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	artist := catalog.Artist{ID: uuid.New(), Name: "Artist"}
	require.NoError(t, repo.AddArtist(ctx, &artist))
	album := catalog.Album{ID: uuid.New(), Title: "Album", AlbumType: catalog.AlbumTypeAlbum}
	require.NoError(t, repo.AddAlbum(ctx, &album))
	require.NoError(t, repo.AddAlbumArtist(ctx, catalog.AlbumArtist{
		AlbumID: album.ID, ArtistID: artist.ID,
	}))

	albums, total, err := repo.ListArtistAlbums(ctx, artist.ID, 20, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, albums, 1)
	assert.Equal(t, album.ID, albums[0].ID)
	require.Len(t, albums[0].Artists, 1)
	assert.Equal(t, artist.ID, albums[0].Artists[0].ID)
	assert.Equal(t, artist.Name, albums[0].Artists[0].Name)
	assert.Zero(t, albums[0].TrackCount)
}
