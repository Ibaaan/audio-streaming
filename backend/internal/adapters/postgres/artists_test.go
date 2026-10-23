//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
)

func TestAddArtist(t *testing.T) {
	repo := newTestRepo(t)
	image := mustAddImage(t, repo)
	artist := catalog.Artist{
		ID:       newID(t),
		Name:     "The Beatles",
		SortName: ptr("Beatles, The"),
		ImageID:  &image.ID,
	}

	err := repo.AddArtist(context.Background(), &artist)

	require.NoError(t, err)
	got := selectOne[catalog.Artist](t, repo, `SELECT * FROM artists WHERE id = $1`, artist.ID)
	assert.Equal(t, artist, got)
}

func TestAddArtistWithoutOptionalFields(t *testing.T) {
	repo := newTestRepo(t)
	artist := catalog.Artist{ID: newID(t), Name: "Slipknot"}

	err := repo.AddArtist(context.Background(), &artist)

	require.NoError(t, err)
	got := selectOne[catalog.Artist](t, repo, `SELECT * FROM artists WHERE id = $1`, artist.ID)
	assert.Nil(t, got.SortName)
	assert.Nil(t, got.ImageID)
}

func TestAddArtistRejectsBlankName(t *testing.T) {
	repo := newTestRepo(t)
	artist := catalog.Artist{ID: newID(t), Name: "   "}

	err := repo.AddArtist(context.Background(), &artist)

	assert.Error(t, err)
}

func TestGetArtist(t *testing.T) {
	repo := newTestRepo(t)
	artist := mustAddArtist(t, repo, "Slipknot")

	got, err := repo.GetArtist(context.Background(), artist.ID)

	require.NoError(t, err)
	assert.Equal(t, artist, got)
}

func TestFindArtistByNameIgnoresCase(t *testing.T) {
	repo := newTestRepo(t)
	artist := mustAddArtist(t, repo, "Slipknot")

	got, err := repo.FindArtistByName(context.Background(), "SLIPKNOT")

	require.NoError(t, err)
	assert.Equal(t, artist, got)
}

func TestFindArtistByNameReturnsOldestDuplicate(t *testing.T) {
	repo := newTestRepo(t)
	first := mustAddArtist(t, repo, "Slipknot")
	mustAddArtist(t, repo, "Slipknot")

	got, err := repo.FindArtistByName(context.Background(), "Slipknot")

	require.NoError(t, err)
	assert.Equal(t, first.ID, got.ID)
}

func TestFindArtistByNameReturnsNotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.FindArtistByName(context.Background(), "Nobody")

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
