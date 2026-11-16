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

func TestSearchCatalogCaseInsensitiveAndLiteral(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	for _, label := range []string{"Radio A", "radio B", "100%_Mix", "Other"} {
		artist := catalog.Artist{ID: uuid.New(), Name: label}
		require.NoError(t, repo.AddArtist(ctx, &artist))
		album := catalog.Album{ID: uuid.New(), Title: label, AlbumType: catalog.AlbumTypeAlbum}
		require.NoError(t, repo.AddAlbum(ctx, &album))
		track := catalog.Track{ID: uuid.New(), Title: label, DurationMs: 1000}
		require.NoError(t, repo.AddTrack(ctx, &track))
	}

	artists, artistTotal, err := repo.SearchArtists(ctx, "RADIO", 1, 1)

	require.NoError(t, err)
	assert.Equal(t, 2, artistTotal)
	require.Len(t, artists, 1)
	assert.Equal(t, "radio B", artists[0].Name)
	albums, albumTotal, err := repo.SearchAlbums(ctx, "RADIO", 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, albumTotal)
	require.Len(t, albums, 1)
	assert.Equal(t, "radio B", albums[0].Title)
	tracks, trackTotal, err := repo.SearchTracks(ctx, "%_", 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, trackTotal)
	require.Len(t, tracks, 1)
	assert.Equal(t, "100%_Mix", tracks[0].Title)
	past, total, err := repo.SearchArtists(ctx, "radio", 20, 100)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Empty(t, past)
}
