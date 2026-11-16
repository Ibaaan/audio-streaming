package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

func TestSearchReturnsRequestedGroups(t *testing.T) {
	artist, album, track := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeRepo{
		artists:      []catalog.Artist{{ID: artist, Name: "Radiohead"}},
		albums:       []catalog.Album{{ID: album, Title: "Radio Collection"}},
		tracks:       []catalog.Track{{ID: track, Title: "Radio Song"}},
		trackArtists: []catalog.TrackArtist{{ArtistID: artist, TrackID: track}},
		albumArtists: []catalog.AlbumArtist{{ArtistID: artist, AlbumID: album}},
		albumTracks:  []catalog.AlbumTrack{{AlbumID: album, TrackID: track, DiscNo: 1, TrackNo: 2}},
	}

	got, err := NewLibrary(repo).Search(context.Background(), dto.SearchRequest{Query: "RADIO"})

	require.NoError(t, err)
	require.NotNil(t, got.Artists)
	require.Len(t, got.Artists.Artists, 1)
	assert.Equal(t, artist, got.Artists.Artists[0].ID)
	require.NotNil(t, got.Albums)
	require.Len(t, got.Albums.Albums, 1)
	assert.Equal(t, 1, got.Albums.Albums[0].TrackCount)
	require.NotNil(t, got.Tracks)
	require.Len(t, got.Tracks.Tracks, 1)
	assert.Equal(t, album, got.Tracks.Tracks[0].Album.ID)
	assert.Equal(t, 2, *got.Tracks.Tracks[0].TrackNumber)
}

func TestSearchPaginationAndOmittedGroups(t *testing.T) {
	repo := &fakeRepo{artists: []catalog.Artist{
		{ID: uuid.New(), Name: "Artist A"},
		{ID: uuid.New(), Name: "Artist B"},
		{ID: uuid.New(), Name: "Artist C"},
	}}

	got, err := NewLibrary(repo).Search(context.Background(), dto.SearchRequest{
		Query: "Artist", Types: []dto.SearchType{dto.SearchTypeArtist}, PageSize: 1, Offset: 1,
	})

	require.NoError(t, err)
	assert.Nil(t, got.Albums)
	assert.Nil(t, got.Tracks)
	require.NotNil(t, got.Artists)
	assert.Equal(t, 3, got.Artists.TotalSize)
	require.NotNil(t, got.Artists.NextOffset)
	assert.Equal(t, 2, *got.Artists.NextOffset)
	require.Len(t, got.Artists.Artists, 1)
	assert.Equal(t, "Artist B", got.Artists.Artists[0].DisplayName)
	data, err := json.Marshal(got)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"albums"`)
}

func TestSearchEmptyGroups(t *testing.T) {
	lib := NewLibrary(&fakeRepo{})

	got, err := lib.Search(context.Background(), dto.SearchRequest{Query: "absent"})

	require.NoError(t, err)
	data, err := json.Marshal(got)
	require.NoError(t, err)
	assert.JSONEq(t, `{"artists":{"artists":[],"total_size":0},
 "albums":{"albums":[],"total_size":0},"tracks":{"tracks":[],"total_size":0}}`, string(data))
}

func TestSearchRejectsInvalidParameters(t *testing.T) {
	for _, req := range []dto.SearchRequest{
		{Query: ""}, {Query: "   "}, {Query: strings.Repeat("я", 201)},
		{Query: "valid", Offset: -1}, {Query: "valid", PageSize: -1},
		{Query: "valid", Types: []dto.SearchType{"other"}},
		{Query: "valid", Types: []dto.SearchType{dto.SearchTypeTrack, dto.SearchTypeTrack}},
	} {
		_, err := NewLibrary(&fakeRepo{}).Search(context.Background(), req)

		assert.ErrorIs(t, err, ErrInvalidSearch)
	}
}
