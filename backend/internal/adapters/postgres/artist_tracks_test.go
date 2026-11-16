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

func TestListArtistTracksOrdersAndPaginates(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	artist := catalog.Artist{ID: uuid.New(), Name: "Guest"}
	require.NoError(t, repo.AddArtist(ctx, &artist))
	featured := "featured"
	ids := []uuid.UUID{}
	for _, year := range []int16{2000, 2020} {
		track := catalog.Track{ID: uuid.New(), Title: "Song", DurationMs: 1000}
		require.NoError(t, repo.AddTrack(ctx, &track))
		ids = append(ids, track.ID)
		require.NoError(t, repo.AddTrackArtist(ctx, catalog.TrackArtist{
			ArtistID: artist.ID, TrackID: track.ID, Role: &featured,
		}))
		album := catalog.Album{ID: uuid.New(), Title: "Album", AlbumType: catalog.AlbumTypeAlbum,
			ReleaseYear: &year}
		require.NoError(t, repo.AddAlbum(ctx, &album))
		require.NoError(t, repo.AddAlbumTrack(ctx, catalog.AlbumTrack{
			AlbumID: album.ID, TrackID: track.ID, DiscNo: 1, TrackNo: 1, Quality: catalog.QualityOptimal,
		}))
	}

	tracks, total, err := repo.ListArtistTracks(ctx, artist.ID, 1, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, tracks, 1)
	assert.Equal(t, ids[1], tracks[0].ID)
	require.NotNil(t, tracks[0].Role)
	assert.Equal(t, featured, *tracks[0].Role)
	second, total, err := repo.ListArtistTracks(ctx, artist.ID, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, second, 1)
	assert.Equal(t, ids[0], second[0].ID)
	past, total, err := repo.ListArtistTracks(ctx, artist.ID, 1, 100)
	require.NoError(t, err)
	assert.Empty(t, past)
	assert.Equal(t, 2, total)
}
