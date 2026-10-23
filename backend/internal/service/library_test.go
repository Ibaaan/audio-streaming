package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/audio"
	"audio-streaming/internal/catalog"
)

func newTestLibrary() (*Library, *fakeRepo) {
	repo := &fakeRepo{}
	return NewLibrary(repo), repo
}

func testRecord(title string) audio.Record {
	return audio.Record{
		TrackID: uuid.Must(uuid.NewV7()),
		SourceInfo: audio.SourceInfo{
			Lossless:   true,
			SampleRate: 44100,
			Samples:    44100 * 252,
		},
		Tags: audio.Tags{
			Title:       title,
			Artist:      "Slipknot",
			Album:       "Vol. 3",
			TrackNumber: 2,
			Date:        "2004-05-25",
		},
	}
}

func testCover() *catalog.Image {
	return &catalog.Image{
		StoragePath:    "covers/vol3.png",
		OriginalFormat: catalog.ImageFormatPNG,
		Width:          ptr(int32(600)),
		Height:         ptr(int32(600)),
	}
}

func mustSave(t *testing.T, lib *Library, rec audio.Record) catalog.Track {
	t.Helper()
	track, err := lib.Save(context.Background(), rec, nil)
	require.NoError(t, err, "Save")
	return track
}

func TestSaveCreatesTrackArtistAndAlbum(t *testing.T) {
	lib, repo := newTestLibrary()

	track, err := lib.Save(context.Background(), testRecord("Duality"), nil)

	require.NoError(t, err)
	require.Len(t, repo.artists, 1)
	require.Len(t, repo.albums, 1)
	assert.Equal(t, "Duality", track.Title)
	assert.Equal(t, int32(252_000), track.DurationMs)
	assert.Equal(t, []catalog.TrackArtist{
		{TrackID: track.ID, ArtistID: repo.artists[0].ID},
	}, repo.trackArtists)
	assert.Equal(t, []catalog.AlbumArtist{
		{AlbumID: repo.albums[0].ID, ArtistID: repo.artists[0].ID},
	}, repo.albumArtists)
	assert.Equal(t, []catalog.AlbumTrack{{
		AlbumID: repo.albums[0].ID,
		TrackID: track.ID,
		Quality: catalog.QualityLossless,
		DiscNo:  1,
		TrackNo: 2,
	}}, repo.albumTracks)
}

func TestSaveUsesRecordTrackID(t *testing.T) {
	lib, repo := newTestLibrary()
	rec := testRecord("Duality")

	track, err := lib.Save(context.Background(), rec, nil)

	require.NoError(t, err)
	assert.Equal(t, rec.TrackID, track.ID)
	assert.Equal(t, rec.TrackID, repo.tracks[0].ID)
}

func TestSaveReusesArtistAndAlbum(t *testing.T) {
	lib, repo := newTestLibrary()
	mustSave(t, lib, testRecord("Duality"))
	rec := testRecord("Pulse of the Maggots")
	rec.Tags.Artist = "SLIPKNOT"
	rec.Tags.TrackNumber = 3

	_, err := lib.Save(context.Background(), rec, nil)

	require.NoError(t, err)
	assert.Len(t, repo.artists, 1)
	assert.Len(t, repo.albums, 1)
	assert.Len(t, repo.albumTracks, 2)
}

func TestSaveRejectsExistingTrack(t *testing.T) {
	lib, repo := newTestLibrary()
	mustSave(t, lib, testRecord("Duality"))
	rec := testRecord("duality")
	rec.Tags.TrackNumber = 3

	_, err := lib.Save(context.Background(), rec, nil)

	assert.ErrorIs(t, err, ErrTrackExists)
	assert.Len(t, repo.tracks, 1)
}

func TestSaveAllowsSameTitleByOtherArtist(t *testing.T) {
	lib, repo := newTestLibrary()
	mustSave(t, lib, testRecord("Duality"))
	rec := testRecord("Duality")
	rec.Tags.Artist = "Stone Sour"

	_, err := lib.Save(context.Background(), rec, nil)

	require.NoError(t, err)
	assert.Len(t, repo.tracks, 2)
}

func TestSaveUsesAlbumArtistForAlbum(t *testing.T) {
	lib, repo := newTestLibrary()
	rec := testRecord("Duality")
	rec.Tags.AlbumArtist = "Various Artists"

	_, err := lib.Save(context.Background(), rec, nil)

	require.NoError(t, err)
	albumArtist, err := repo.GetArtist(context.Background(), repo.albumArtists[0].ArtistID)
	require.NoError(t, err, "GetArtist")
	assert.Equal(t, "Various Artists", albumArtist.Name)
}

func TestSaveWithoutAlbumSkipsAlbum(t *testing.T) {
	lib, repo := newTestLibrary()
	rec := testRecord("Duality")
	rec.Tags.Album = ""

	_, err := lib.Save(context.Background(), rec, nil)

	require.NoError(t, err)
	assert.Empty(t, repo.albums)
	assert.Empty(t, repo.albumTracks)
}

func TestSaveAddsCoverWithNewAlbum(t *testing.T) {
	lib, repo := newTestLibrary()
	cover := testCover()

	_, err := lib.Save(context.Background(), testRecord("Duality"), cover)

	require.NoError(t, err)
	require.Len(t, repo.images, 1)
	img := repo.images[0]
	assert.Equal(t, cover.StoragePath, img.StoragePath)
	assert.Equal(t, cover.OriginalFormat, img.OriginalFormat)
	assert.Equal(t, cover.Width, img.Width)
	assert.Equal(t, cover.Height, img.Height)
	assert.Equal(t, &img.ID, repo.albums[0].CoverImageID)
}

func TestSaveSharesCoverBetweenAlbums(t *testing.T) {
	lib, repo := newTestLibrary()
	_, err := lib.Save(context.Background(), testRecord("Duality"), testCover())
	require.NoError(t, err, "first Save")
	rec := testRecord("Duality (Live)")
	rec.Tags.Album = "Vol. 3 (Special Edition)"

	_, err = lib.Save(context.Background(), rec, testCover())

	require.NoError(t, err)
	require.Len(t, repo.albums, 2)
	assert.Len(t, repo.images, 1)
	assert.Equal(t, repo.albums[0].CoverImageID, repo.albums[1].CoverImageID)
}

func TestSaveSkipsCoverForExistingAlbum(t *testing.T) {
	lib, repo := newTestLibrary()
	mustSave(t, lib, testRecord("Duality"))
	rec := testRecord("Pulse of the Maggots")
	rec.Tags.TrackNumber = 3

	_, err := lib.Save(context.Background(), rec, testCover())

	require.NoError(t, err)
	assert.Empty(t, repo.images)
	assert.Nil(t, repo.albums[0].CoverImageID)
}

func TestSaveRejectsCoverWithoutPath(t *testing.T) {
	lib, repo := newTestLibrary()
	cover := testCover()
	cover.StoragePath = ""

	_, err := lib.Save(context.Background(), testRecord("Duality"), cover)

	assert.ErrorIs(t, err, ErrInvalidRecord)
	assert.Empty(t, repo.tracks)
}

func TestSaveRejectsInvalidRecord(t *testing.T) {
	lib, repo := newTestLibrary()
	rec := testRecord("  ")

	_, err := lib.Save(context.Background(), rec, nil)

	assert.ErrorIs(t, err, ErrInvalidRecord)
	assert.Empty(t, repo.artists)
}
