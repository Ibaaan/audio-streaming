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

func TestAddAlbum(t *testing.T) {
	repo := newTestRepo(t)
	cover := mustAddImage(t, repo)
	album := catalog.Album{
		ID:           newID(t),
		Title:        "Iowa",
		AlbumType:    catalog.AlbumTypeAlbum,
		ReleaseYear:  ptr[int16](2001),
		CoverImageID: &cover.ID,
	}

	err := repo.AddAlbum(context.Background(), &album)

	require.NoError(t, err)
	got := selectOne[catalog.Album](t, repo, `SELECT * FROM albums WHERE id = $1`, album.ID)
	assert.Equal(t, album, got)
}

func TestAddAlbumRejectsOutOfRangeYear(t *testing.T) {
	repo := newTestRepo(t)
	album := catalog.Album{
		ID:          newID(t),
		Title:       "Iowa",
		AlbumType:   catalog.AlbumTypeAlbum,
		ReleaseYear: ptr[int16](99),
	}

	err := repo.AddAlbum(context.Background(), &album)

	assert.Error(t, err)
}

func TestGetAlbum(t *testing.T) {
	repo := newTestRepo(t)
	album := mustAddAlbum(t, repo, "Iowa")

	got, err := repo.GetAlbum(context.Background(), album.ID)

	require.NoError(t, err)
	assert.Equal(t, album, got)
}

func TestFindAlbumIgnoresCase(t *testing.T) {
	repo := newTestRepo(t)
	artist := mustAddArtist(t, repo, "Slipknot")
	album := mustAddAlbum(t, repo, "Iowa")
	mustAddAlbumArtist(t, repo, album, artist)

	got, err := repo.FindAlbum(context.Background(), "IOWA", artist.ID)

	require.NoError(t, err)
	assert.Equal(t, album, got)
}

func TestFindAlbumIgnoresOtherArtists(t *testing.T) {
	repo := newTestRepo(t)
	owner := mustAddArtist(t, repo, "Slipknot")
	other := mustAddArtist(t, repo, "Stone Sour")
	album := mustAddAlbum(t, repo, "Iowa")
	mustAddAlbumArtist(t, repo, album, owner)

	_, err := repo.FindAlbum(context.Background(), "Iowa", other.ID)

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}

func TestAddAlbumArtist(t *testing.T) {
	repo := newTestRepo(t)
	album := mustAddAlbum(t, repo, "Iowa")
	artist := mustAddArtist(t, repo, "Slipknot")
	albumArtist := catalog.AlbumArtist{
		AlbumID:  album.ID,
		ArtistID: artist.ID,
		Position: 0,
	}

	err := repo.AddAlbumArtist(context.Background(), albumArtist)

	require.NoError(t, err)
	got := selectOne[catalog.AlbumArtist](t, repo,
		`SELECT * FROM album_artists WHERE album_id = $1`, album.ID)
	assert.Equal(t, albumArtist, got)
}

func TestAddAlbumArtistRejectsUnknownAlbum(t *testing.T) {
	repo := newTestRepo(t)
	artist := mustAddArtist(t, repo, "Slipknot")
	albumArtist := catalog.AlbumArtist{
		AlbumID:  uuid.New(),
		ArtistID: artist.ID,
	}

	err := repo.AddAlbumArtist(context.Background(), albumArtist)

	assert.Error(t, err)
}

func TestAddAlbumTrack(t *testing.T) {
	repo := newTestRepo(t)
	album := mustAddAlbum(t, repo, "Iowa")
	track := mustAddTrack(t, repo)
	albumTrack := catalog.AlbumTrack{
		AlbumID: album.ID,
		TrackID: track.ID,
		Quality: catalog.QualityLossless,
		DiscNo:  1,
		TrackNo: 1,
	}

	err := repo.AddAlbumTrack(context.Background(), albumTrack)

	require.NoError(t, err)
	got := selectOne[catalog.AlbumTrack](t, repo,
		`SELECT * FROM album_tracks WHERE album_id = $1`, album.ID)
	assert.Equal(t, albumTrack, got)
}

func TestAddAlbumTrackRejectsDuplicateTrackNumber(t *testing.T) {
	repo := newTestRepo(t)
	album := mustAddAlbum(t, repo, "Iowa")
	first := mustAddTrack(t, repo)
	second := mustAddTrack(t, repo)
	err := repo.AddAlbumTrack(context.Background(), catalog.AlbumTrack{
		AlbumID: album.ID,
		TrackID: first.ID,
		Quality: catalog.QualityLossless,
		DiscNo:  1,
		TrackNo: 1,
	})
	require.NoError(t, err, "first AddAlbumTrack")

	err = repo.AddAlbumTrack(context.Background(), catalog.AlbumTrack{
		AlbumID: album.ID,
		TrackID: second.ID,
		Quality: catalog.QualityLossless,
		DiscNo:  1,
		TrackNo: 1,
	})

	assert.Error(t, err)
}
