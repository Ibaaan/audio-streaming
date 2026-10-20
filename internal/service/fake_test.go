package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

// fakeRepo is an in-memory catalog.Repository. WithinTx rolls back by
// restoring a copy of the slices taken before fn ran.
type fakeRepo struct {
	tracks       []catalog.Track
	artists      []catalog.Artist
	albums       []catalog.Album
	images       []catalog.Image
	trackArtists []catalog.TrackArtist
	albumArtists []catalog.AlbumArtist
	albumTracks  []catalog.AlbumTrack
}

var _ catalog.Repository = (*fakeRepo)(nil)

func (f *fakeRepo) WithinTx(
	ctx context.Context,
	fn func(tx catalog.Repository) error,
) error {
	saved := *f
	if err := fn(f); err != nil {
		*f = saved
		return err
	}
	return nil
}

func first[T any](items []T, match func(T) bool) (T, error) {
	for _, v := range items {
		if match(v) {
			return v, nil
		}
	}
	var zero T
	return zero, catalog.ErrNotFound
}

func (f *fakeRepo) AddTrack(ctx context.Context, t *catalog.Track) error {
	f.tracks = append(f.tracks, *t)
	return nil
}

func (f *fakeRepo) GetTrack(
	ctx context.Context,
	id uuid.UUID,
) (catalog.Track, error) {
	return first(f.tracks, func(t catalog.Track) bool { return t.ID == id })
}

func (f *fakeRepo) FindTrack(
	ctx context.Context,
	title string,
	artistID uuid.UUID,
) (catalog.Track, error) {
	return first(f.tracks, func(t catalog.Track) bool {
		_, err := first(f.trackArtists, func(ta catalog.TrackArtist) bool {
			return ta.TrackID == t.ID && ta.ArtistID == artistID
		})
		return err == nil && strings.EqualFold(t.Title, title)
	})
}

func (f *fakeRepo) AddTrackArtist(
	ctx context.Context,
	ta catalog.TrackArtist,
) error {
	f.trackArtists = append(f.trackArtists, ta)
	return nil
}

func (f *fakeRepo) AddArtist(ctx context.Context, a *catalog.Artist) error {
	f.artists = append(f.artists, *a)
	return nil
}

func (f *fakeRepo) GetArtist(
	ctx context.Context,
	id uuid.UUID,
) (catalog.Artist, error) {
	return first(f.artists, func(a catalog.Artist) bool { return a.ID == id })
}

func (f *fakeRepo) FindArtistByName(
	ctx context.Context,
	name string,
) (catalog.Artist, error) {
	return first(f.artists, func(a catalog.Artist) bool {
		return strings.EqualFold(a.Name, name)
	})
}

func (f *fakeRepo) AddAlbum(ctx context.Context, a *catalog.Album) error {
	f.albums = append(f.albums, *a)
	return nil
}

func (f *fakeRepo) GetAlbum(
	ctx context.Context,
	id uuid.UUID,
) (catalog.Album, error) {
	return first(f.albums, func(a catalog.Album) bool { return a.ID == id })
}

func (f *fakeRepo) FindAlbum(
	ctx context.Context,
	title string,
	artistID uuid.UUID,
) (catalog.Album, error) {
	return first(f.albums, func(a catalog.Album) bool {
		_, err := first(f.albumArtists, func(aa catalog.AlbumArtist) bool {
			return aa.AlbumID == a.ID && aa.ArtistID == artistID
		})
		return err == nil && strings.EqualFold(a.Title, title)
	})
}

func (f *fakeRepo) AddAlbumArtist(
	ctx context.Context,
	aa catalog.AlbumArtist,
) error {
	f.albumArtists = append(f.albumArtists, aa)
	return nil
}

func (f *fakeRepo) AddAlbumTrack(
	ctx context.Context,
	at catalog.AlbumTrack,
) error {
	f.albumTracks = append(f.albumTracks, at)
	return nil
}

func (f *fakeRepo) AddImage(ctx context.Context, img *catalog.Image) error {
	f.images = append(f.images, *img)
	return nil
}

func (f *fakeRepo) GetImage(
	ctx context.Context,
	id uuid.UUID,
) (catalog.Image, error) {
	return first(f.images, func(i catalog.Image) bool { return i.ID == id })
}

func (f *fakeRepo) FindImageByPath(
	ctx context.Context,
	storagePath string,
) (catalog.Image, error) {
	return first(f.images, func(i catalog.Image) bool {
		return i.StoragePath == storagePath
	})
}
