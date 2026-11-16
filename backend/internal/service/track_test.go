package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

func TestGetTrackAssemblesCreditsAndAllAlbums(t *testing.T) {
	id, main, guest := uuid.New(), uuid.New(), uuid.New()
	albumID, compilationID := uuid.New(), uuid.New()
	year := int16(2004)
	featured := "featured"
	created := time.Date(2026, 11, 16, 12, 0, 0, 0, time.FixedZone("local", 5*60*60))
	repo := &fakeRepo{
		tracks:  []catalog.Track{{ID: id, Title: "Song", DurationMs: 252000, CreatedAt: created, UpdatedAt: created}},
		artists: []catalog.Artist{{ID: main, Name: "Main"}, {ID: guest, Name: "Guest"}},
		trackArtists: []catalog.TrackArtist{
			{TrackID: id, ArtistID: guest, Position: 1, Role: &featured},
			{TrackID: id, ArtistID: main, Position: 0},
		},
		albums: []catalog.Album{
			{ID: albumID, Title: "Original", AlbumType: catalog.AlbumTypeAlbum, ReleaseYear: &year},
			{
				ID: compilationID, Title: "Compilation", AlbumType: catalog.AlbumTypeCompilation,
				CreatedAt: created,
			},
		},
		albumTracks: []catalog.AlbumTrack{
			{TrackID: id, AlbumID: compilationID, DiscNo: 2, TrackNo: 7},
			{TrackID: id, AlbumID: albumID, DiscNo: 1, TrackNo: 4},
		},
	}

	got, err := NewLibrary(repo).GetTrack(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, dto.TrackRef{ID: id, Title: "Song", DurationMs: 252000}, got.TrackRef)
	require.Len(t, got.Artists, 2)
	assert.Equal(t, main, got.Artists[0].ID)
	assert.Equal(t, "Main", got.Artists[0].DisplayName)
	assert.Equal(t, dto.ArtistRoleMain, got.Artists[0].Role)
	assert.Equal(t, dto.ArtistRoleFeatured, got.Artists[1].Role)
	require.Len(t, got.Albums, 2)
	assert.Equal(t, albumID, got.Albums[0].ID)
	assert.Equal(t, 2004, *got.Albums[0].ReleaseYear)
	assert.Equal(t, 4, got.Albums[0].TrackNumber)
	assert.Equal(t, 2, got.Albums[1].DiscNumber)
	assert.Nil(t, got.Albums[1].ReleaseYear)
	assert.Equal(t, created.UTC(), got.CreateTime)
	assert.Equal(t, created.UTC(), got.UpdateTime)
}

func TestGetTrackWithoutRelationsEncodesEmptyArrays(t *testing.T) {
	id := uuid.New()
	lib := NewLibrary(&fakeRepo{tracks: []catalog.Track{{ID: id, Title: "Solo"}}})

	got, err := lib.GetTrack(context.Background(), id)

	require.NoError(t, err)
	data, err := json.Marshal(got)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"artists":[]`)
	assert.Contains(t, string(data), `"albums":[]`)
	assert.NotContains(t, string(data), `"qualities"`)
	assert.NotContains(t, string(data), `"name"`)
}

func TestGetTrackNotFound(t *testing.T) {
	lib := NewLibrary(&fakeRepo{})

	_, err := lib.GetTrack(context.Background(), uuid.New())

	assert.ErrorIs(t, err, catalog.ErrNotFound)
}
