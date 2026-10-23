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

func TestAddTrack(t *testing.T) {
	repo := newTestRepo(t)
	track := catalog.Track{
		ID:         newID(t),
		Title:      "Duality",
		DurationMs: 252_000,
	}

	err := repo.AddTrack(context.Background(), &track)

	require.NoError(t, err)
	assert.False(t, track.CreatedAt.IsZero(), "CreatedAt filled in")
	got := selectOne[catalog.Track](t, repo, `SELECT * FROM tracks WHERE id = $1`, track.ID)
	assert.Equal(t, track, got)
}

func TestAddTrackRejectsZeroDuration(t *testing.T) {
	repo := newTestRepo(t)
	track := catalog.Track{ID: newID(t), Title: "Duality"}

	err := repo.AddTrack(context.Background(), &track)

	assert.Error(t, err)
}

func TestAddTrackRejectsDuplicateID(t *testing.T) {
	repo := newTestRepo(t)
	existing := mustAddTrack(t, repo)
	track := catalog.Track{
		ID:         existing.ID,
		Title:      "Vermilion",
		DurationMs: 316_000,
	}

	err := repo.AddTrack(context.Background(), &track)

	assert.Error(t, err)
}

func TestGetTrack(t *testing.T) {
	repo := newTestRepo(t)
	track := mustAddTrack(t, repo)

	got, err := repo.GetTrack(context.Background(), track.ID)

	require.NoError(t, err)
	assert.Equal(t, track, got)
}

func TestGetTrackReturnsNotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.GetTrack(context.Background(), uuid.New())

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}

func TestAddTrackArtist(t *testing.T) {
	repo := newTestRepo(t)
	track := mustAddTrack(t, repo)
	artist := mustAddArtist(t, repo, "Slipknot")
	trackArtist := catalog.TrackArtist{
		TrackID:  track.ID,
		ArtistID: artist.ID,
		Role:     ptr("vocals"),
		Position: 0,
	}

	err := repo.AddTrackArtist(context.Background(), trackArtist)

	require.NoError(t, err)
	got := selectOne[catalog.TrackArtist](t, repo,
		`SELECT * FROM track_artists WHERE track_id = $1`, track.ID)
	assert.Equal(t, trackArtist, got)
}

func TestAddTrackArtistRejectsDuplicatePosition(t *testing.T) {
	repo := newTestRepo(t)
	track := mustAddTrack(t, repo)
	first := mustAddArtist(t, repo, "Slipknot")
	second := mustAddArtist(t, repo, "Stone Sour")
	err := repo.AddTrackArtist(context.Background(), catalog.TrackArtist{
		TrackID:  track.ID,
		ArtistID: first.ID,
	})
	require.NoError(t, err, "first AddTrackArtist")

	err = repo.AddTrackArtist(context.Background(), catalog.TrackArtist{
		TrackID:  track.ID,
		ArtistID: second.ID,
	})

	assert.Error(t, err)
}

func mustAddTrackArtist(
	t *testing.T,
	repo *Repository,
	track catalog.Track,
	artist catalog.Artist,
) {
	t.Helper()
	trackArtist := catalog.TrackArtist{
		TrackID:  track.ID,
		ArtistID: artist.ID,
	}
	require.NoError(t, repo.AddTrackArtist(context.Background(), trackArtist), "AddTrackArtist")
}

func TestFindTrackIgnoresCase(t *testing.T) {
	repo := newTestRepo(t)
	artist := mustAddArtist(t, repo, "Slipknot")
	track := mustAddTrack(t, repo)
	mustAddTrackArtist(t, repo, track, artist)

	got, err := repo.FindTrack(context.Background(), "DUALITY", artist.ID)

	require.NoError(t, err)
	assert.Equal(t, track, got)
}

func TestFindTrackIgnoresOtherArtists(t *testing.T) {
	repo := newTestRepo(t)
	owner := mustAddArtist(t, repo, "Slipknot")
	other := mustAddArtist(t, repo, "Stone Sour")
	track := mustAddTrack(t, repo)
	mustAddTrackArtist(t, repo, track, owner)

	_, err := repo.FindTrack(context.Background(), "Duality", other.ID)

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}

func TestListTracksJoinsArtistsAndAlbum(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	track := mustAddTrack(t, repo)
	corey := mustAddArtist(t, repo, "Corey Taylor")
	slipknot := mustAddArtist(t, repo, "Slipknot")
	require.NoError(t, repo.AddTrackArtist(ctx, catalog.TrackArtist{
		TrackID: track.ID, ArtistID: slipknot.ID, Position: 0,
	}), "AddTrackArtist")
	require.NoError(t, repo.AddTrackArtist(ctx, catalog.TrackArtist{
		TrackID: track.ID, ArtistID: corey.ID, Position: 1,
	}), "AddTrackArtist")
	album := catalog.Album{
		ID:          newID(t),
		Title:       "Vol. 3",
		AlbumType:   catalog.AlbumTypeAlbum,
		ReleaseYear: ptr(int16(2004)),
	}
	require.NoError(t, repo.AddAlbum(ctx, &album), "AddAlbum")
	require.NoError(t, repo.AddAlbumTrack(ctx, catalog.AlbumTrack{
		AlbumID: album.ID, TrackID: track.ID,
		Quality: catalog.QualityLossless, DiscNo: 1, TrackNo: 2,
	}), "AddAlbumTrack")

	got, err := repo.ListTracks(ctx)

	require.NoError(t, err)
	assert.Equal(t, []catalog.TrackSummary{{
		ID:          track.ID,
		Title:       "Duality",
		Artist:      "Slipknot, Corey Taylor",
		Album:       ptr("Vol. 3"),
		ReleaseYear: ptr(int16(2004)),
		DurationMs:  252_000,
		CreatedAt:   track.CreatedAt,
	}}, got)
}

func TestListTracksWithoutAlbum(t *testing.T) {
	repo := newTestRepo(t)
	mustAddTrack(t, repo)

	got, err := repo.ListTracks(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Nil(t, got[0].Album)
	assert.Nil(t, got[0].ReleaseYear)
}

func TestListTracksEmpty(t *testing.T) {
	repo := newTestRepo(t)

	got, err := repo.ListTracks(context.Background())

	require.NoError(t, err)
	assert.Empty(t, got)
}
