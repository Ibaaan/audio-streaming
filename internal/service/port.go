package service

import (
	"context"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

// The stores below are what the service needs from storage, one small
// interface per entity. Add methods store a value the caller built, ID
// included; for entities they also fill in the database timestamps.
// Lookups that match nothing return catalog.ErrNotFound.
// *postgres.Repository satisfies all of them.

type TrackStore interface {
	AddTrack(ctx context.Context, t *catalog.Track) error
	GetTrack(ctx context.Context, id uuid.UUID) (catalog.Track, error)
	AddTrackArtist(ctx context.Context, ta catalog.TrackArtist) error
}

type ArtistStore interface {
	AddArtist(ctx context.Context, a *catalog.Artist) error
	GetArtist(ctx context.Context, id uuid.UUID) (catalog.Artist, error)
	FindArtistByName(ctx context.Context, name string) (
		catalog.Artist, error)
}

type AlbumStore interface {
	AddAlbum(ctx context.Context, a *catalog.Album) error
	GetAlbum(ctx context.Context, id uuid.UUID) (catalog.Album, error)
	FindAlbum(ctx context.Context, title string, artistID uuid.UUID) (
		catalog.Album, error)
	AddAlbumArtist(ctx context.Context, aa catalog.AlbumArtist) error
	AddAlbumTrack(ctx context.Context, at catalog.AlbumTrack) error
}

type ImageStore interface {
	AddImage(ctx context.Context, img *catalog.Image) error
	GetImage(ctx context.Context, id uuid.UUID) (catalog.Image, error)
	FindImageByPath(ctx context.Context, storagePath string) (
		catalog.Image, error)
}

// Repository is every store together, for code that touches the whole
// catalog, such as importing a track.
type Repository interface {
	TrackStore
	ArtistStore
	AlbumStore
	ImageStore
}
