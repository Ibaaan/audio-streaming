//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
	"audio-streaming/internal/service"
)

type countedReads struct {
	DBTX
	calls int
}

func (c *countedReads) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	c.calls++
	return c.DBTX.Query(ctx, sql, args...)
}
func (c *countedReads) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	c.calls++
	return c.DBTX.QueryRow(ctx, sql, args...)
}

func TestSummaryQueriesHaveFixedRoundTrips(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	main := catalog.Artist{ID: uuid.New(), Name: "Match main"}
	guest := catalog.Artist{ID: uuid.New(), Name: "Match guest"}
	require.NoError(t, repo.AddArtist(ctx, &main))
	require.NoError(t, repo.AddArtist(ctx, &guest))
	old := catalog.Album{ID: uuid.New(), Title: "Match original", AlbumType: catalog.AlbumTypeAlbum}
	compilation := catalog.Album{ID: uuid.New(), Title: "Match compilation",
		AlbumType: catalog.AlbumTypeCompilation}
	require.NoError(t, repo.AddAlbum(ctx, &old))
	require.NoError(t, repo.AddAlbum(ctx, &compilation))
	_, err := repo.dbtx.Exec(ctx,
		`UPDATE albums SET created_at=$2 WHERE id=$1`, old.ID, time.Unix(0, 0),
	)
	require.NoError(t, err)
	for _, album := range []catalog.Album{old, compilation} {
		require.NoError(t, repo.AddAlbumArtist(ctx, catalog.AlbumArtist{
			AlbumID: album.ID, ArtistID: main.ID,
		}))
	}
	featured := "featured"
	for i := 1; i <= 3; i++ {
		track := catalog.Track{ID: uuid.New(), Title: "Match song", DurationMs: 1000}
		require.NoError(t, repo.AddTrack(ctx, &track))
		require.NoError(t, repo.AddTrackArtist(ctx, catalog.TrackArtist{
			TrackID: track.ID, ArtistID: guest.ID, Position: 1, Role: &featured,
		}))
		require.NoError(t, repo.AddTrackArtist(ctx, catalog.TrackArtist{
			TrackID: track.ID, ArtistID: main.ID, Position: 0,
		}))
		for _, album := range []catalog.Album{old, compilation} {
			require.NoError(t, repo.AddAlbumTrack(ctx, catalog.AlbumTrack{
				TrackID: track.ID, AlbumID: album.ID, DiscNo: 1, TrackNo: int16(i),
				Quality: catalog.QualityOptimal,
			}))
		}
	}
	reads := &countedReads{DBTX: repo.dbtx}
	lib := service.NewLibrary(NewRepository(reads))

	got, err := lib.Search(ctx, dto.SearchRequest{Query: "Match", PageSize: 100})

	require.NoError(t, err)
	assert.Equal(t, 6, reads.calls)
	require.Len(t, got.Tracks.Tracks, 3)
	for _, track := range got.Tracks.Tracks {
		require.Len(t, track.Artists, 2)
		assert.Equal(t, main.ID, track.Artists[0].ID)
		assert.Equal(t, dto.ArtistRoleFeatured, track.Artists[1].Role)
		require.NotNil(t, track.Album)
		assert.Equal(t, old.ID, track.Album.ID)
	}
	require.Len(t, got.Albums.Albums, 2)
	for _, album := range got.Albums.Albums {
		assert.Equal(t, 3, album.TrackCount)
	}
	reads.calls = 0
	page, err := lib.ListArtistTracks(ctx, guest.ID, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, reads.calls)
	assert.Equal(t, 3, page.TotalSize)
	require.Len(t, page.Tracks, 2)
	assert.Equal(t, dto.ArtistRoleFeatured, page.Tracks[0].ArtistRole)
	require.NotNil(t, page.NextOffset)
	assert.Equal(t, 2, *page.NextOffset)
}
