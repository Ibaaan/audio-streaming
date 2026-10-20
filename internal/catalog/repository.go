package catalog

import (
	"context"

	"github.com/google/uuid"
)

// The repositories below are what code using the catalog needs from
// storage, one small interface per entity. Add methods store a value the
// caller built, ID included; for entities they also fill in the database
// timestamps. Lookups that match nothing return ErrNotFound.
// *postgres.Repository satisfies all of them.

type TrackRepository interface {
	AddTrack(ctx context.Context, t *Track) error
	GetTrack(ctx context.Context, id uuid.UUID) (Track, error)
	FindTrack(ctx context.Context, title string, artistID uuid.UUID) (
		Track, error)
	AddTrackArtist(ctx context.Context, ta TrackArtist) error
}

type ArtistRepository interface {
	AddArtist(ctx context.Context, a *Artist) error
	GetArtist(ctx context.Context, id uuid.UUID) (Artist, error)
	FindArtistByName(ctx context.Context, name string) (Artist, error)
}

type AlbumRepository interface {
	AddAlbum(ctx context.Context, a *Album) error
	GetAlbum(ctx context.Context, id uuid.UUID) (Album, error)
	FindAlbum(ctx context.Context, title string, artistID uuid.UUID) (
		Album, error)
	AddAlbumArtist(ctx context.Context, aa AlbumArtist) error
	AddAlbumTrack(ctx context.Context, at AlbumTrack) error
}

type ImageRepository interface {
	AddImage(ctx context.Context, img *Image) error
	GetImage(ctx context.Context, id uuid.UUID) (Image, error)
	FindImageByPath(ctx context.Context, storagePath string) (Image, error)
}

// Repository is every repository together, for code that touches the whole
// catalog, such as importing a track. It lives here rather than with its
// users so that implementations can name it in WithinTx's callback
// without importing them.
type Repository interface {
	TrackRepository
	ArtistRepository
	AlbumRepository
	ImageRepository

	// WithinTx runs fn with a Repository bound to a new transaction and
	// commits if fn returns nil, rolling back otherwise.
	WithinTx(ctx context.Context, fn func(tx Repository) error) error
}
