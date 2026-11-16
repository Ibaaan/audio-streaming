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
	ListArtistTracks(
		ctx context.Context, artistID uuid.UUID, limit, offset int,
	) ([]ArtistTrack, int, error)
	AddTrack(ctx context.Context, t *Track) error
	GetTrack(ctx context.Context, id uuid.UUID) (Track, error)
	FindTrack(ctx context.Context, title string, artistID uuid.UUID) (Track, error)
	ListTracks(ctx context.Context) ([]TrackSummary, error)
	AddTrackArtist(ctx context.Context, ta TrackArtist) error
}

type ArtistRepository interface {
	GetArtistStats(ctx context.Context, id uuid.UUID) (ArtistStats, error)
	ListArtistAlbums(
		ctx context.Context, id uuid.UUID, limit, offset int,
	) ([]AlbumOverview, int, error)
	AddArtist(ctx context.Context, a *Artist) error
	GetArtist(ctx context.Context, id uuid.UUID) (Artist, error)
	FindArtistByName(ctx context.Context, name string) (Artist, error)
	GetArtistByTrackId(ctx context.Context, id uuid.UUID) ([]TrackArtistCredit, error)
}

type AlbumRepository interface {
	GetAlbumDetails(ctx context.Context, id uuid.UUID) (AlbumDetails, error)
	AddAlbum(ctx context.Context, a *Album) error
	GetAlbum(ctx context.Context, id uuid.UUID) (Album, error)
	FindAlbum(ctx context.Context, title string, artistID uuid.UUID) (Album, error)
	AddAlbumArtist(ctx context.Context, aa AlbumArtist) error
	AddAlbumTrack(ctx context.Context, at AlbumTrack) error
	GetAlbumByTrackId(ctx context.Context, id uuid.UUID) ([]TrackAlbumAppearance, error)
}

type ImageRepository interface {
	AddImage(ctx context.Context, img *Image) error
	GetImage(ctx context.Context, id uuid.UUID) (Image, error)
	FindImageByPath(ctx context.Context, storagePath string) (Image, error)
}

// SearchRepository performs literal, case-insensitive substring searches.
// Results are ordered by name/title, then ID, for stable offset pagination.
type SearchRepository interface {
	SearchArtists(ctx context.Context, query string, limit, offset int) ([]Artist, int, error)
	SearchAlbums(ctx context.Context, query string, limit, offset int) ([]Album, int, error)
	SearchTracks(ctx context.Context, query string, limit, offset int) ([]Track, int, error)
}

// Repository is every repository together, for code that touches the whole
// catalog, such as importing a track. It lives here rather than with its
// users so that implementations can name it in WithinTx's callback
// without importing them.
type Repository interface {
	SearchRepository
	TrackRepository
	ArtistRepository
	AlbumRepository
	ImageRepository

	// WithinTx runs fn with a Repository bound to a new transaction and
	// commits if fn returns nil, rolling back otherwise.
	WithinTx(ctx context.Context, fn func(tx Repository) error) error
}
