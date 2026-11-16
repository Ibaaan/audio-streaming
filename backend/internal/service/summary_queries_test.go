package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

// Summary endpoints must not fall back to per-result detail reads.
type summaryOnlyRepo struct{ *fakeRepo }

func (r *summaryOnlyRepo) GetTrack(context.Context, uuid.UUID) (catalog.Track, error) {
	panic("unexpected per-result track lookup")
}
func (r *summaryOnlyRepo) GetAlbumDetails(
	context.Context, uuid.UUID,
) (catalog.AlbumDetails, error) {
	panic("unexpected per-result album lookup")
}
func (r *summaryOnlyRepo) GetArtistByTrackId(
	context.Context, uuid.UUID,
) ([]catalog.TrackArtistCredit, error) {
	panic("unexpected per-result credit lookup")
}
func (r *summaryOnlyRepo) GetAlbumByTrackId(
	context.Context, uuid.UUID,
) ([]catalog.TrackAlbumAppearance, error) {
	panic("unexpected per-result album lookup")
}

func TestSearchUsesSummaryReadsOnly(t *testing.T) {
	artist, album, track := uuid.New(), uuid.New(), uuid.New()
	repo := &summaryOnlyRepo{fakeRepo: &fakeRepo{
		artists:      []catalog.Artist{{ID: artist, Name: "Match"}},
		albums:       []catalog.Album{{ID: album, Title: "Match"}},
		tracks:       []catalog.Track{{ID: track, Title: "Match"}},
		albumArtists: []catalog.AlbumArtist{{ArtistID: artist, AlbumID: album}},
		trackArtists: []catalog.TrackArtist{{ArtistID: artist, TrackID: track}},
		albumTracks:  []catalog.AlbumTrack{{AlbumID: album, TrackID: track, DiscNo: 1, TrackNo: 2}},
	}}

	got, err := NewLibrary(repo).Search(context.Background(), dto.SearchRequest{Query: "Match"})

	require.NoError(t, err)
	require.Len(t, got.Albums.Albums, 1)
	assert.Equal(t, 1, got.Albums.Albums[0].TrackCount)
	require.Len(t, got.Tracks.Tracks, 1)
	assert.Equal(t, album, got.Tracks.Tracks[0].Album.ID)
	assert.Equal(t, 2, *got.Tracks.Tracks[0].TrackNumber)
}

func TestArtistTrackListUsesSummaryReadsOnly(t *testing.T) {
	artist, track := uuid.New(), uuid.New()
	featured := "featured"
	repo := &summaryOnlyRepo{fakeRepo: &fakeRepo{
		artists:      []catalog.Artist{{ID: artist, Name: "Artist"}},
		tracks:       []catalog.Track{{ID: track, Title: "Song"}},
		trackArtists: []catalog.TrackArtist{{ArtistID: artist, TrackID: track, Role: &featured}},
	}}

	got, err := NewLibrary(repo).ListArtistTracks(context.Background(), artist, 20, 0)

	require.NoError(t, err)
	require.Len(t, got.Tracks, 1)
	assert.Equal(t, dto.ArtistRoleFeatured, got.Tracks[0].ArtistRole)
	assert.Nil(t, got.Tracks[0].Album)
	assert.NotNil(t, got.Tracks[0].Artists)
}
