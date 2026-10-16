//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
)

func TestAddImage(t *testing.T) {
	repo := newTestRepo(t)
	image := catalog.Image{
		ID:             newID(t),
		StoragePath:    "covers/a.png",
		OriginalFormat: catalog.ImageFormatPNG,
		Width:          ptr[int32](600),
		Height:         ptr[int32](400),
	}

	err := repo.AddImage(context.Background(), &image)

	require.NoError(t, err)
	got := selectOne[catalog.Image](t, repo,
		`SELECT * FROM images WHERE id = $1`, image.ID)
	assert.Equal(t, image, got)
}

func TestAddImageRejectsDuplicateStoragePath(t *testing.T) {
	repo := newTestRepo(t)
	existing := mustAddImage(t, repo)
	image := catalog.Image{
		ID:             newID(t),
		StoragePath:    existing.StoragePath,
		OriginalFormat: catalog.ImageFormatJPEG,
	}

	err := repo.AddImage(context.Background(), &image)

	assert.Error(t, err)
}

func TestGetImage(t *testing.T) {
	repo := newTestRepo(t)
	image := mustAddImage(t, repo)

	got, err := repo.GetImage(context.Background(), image.ID)

	require.NoError(t, err)
	assert.Equal(t, image, got)
}

func TestFindImageByPath(t *testing.T) {
	repo := newTestRepo(t)
	image := mustAddImage(t, repo)

	got, err := repo.FindImageByPath(context.Background(),
		image.StoragePath)

	require.NoError(t, err)
	assert.Equal(t, image, got)
}

func TestFindImageByPathReturnsNotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.FindImageByPath(context.Background(), "covers/none")

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
