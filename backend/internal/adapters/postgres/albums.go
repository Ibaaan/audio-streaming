package postgres

import (
	"context"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

const insertAlbumQuery = `
INSERT INTO albums (id, title, album_type, release_year, cover_image_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at, updated_at`

// AddAlbum inserts a and fills in the timestamps the database sets.
func (r *Repository) AddAlbum(ctx context.Context, a *catalog.Album) error {
	return r.dbtx.QueryRow(ctx, insertAlbumQuery,
		a.ID, a.Title, a.AlbumType, a.ReleaseYear,
		a.CoverImageID,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

const selectAlbumQuery = `SELECT * FROM albums WHERE id = $1`

func (r *Repository) GetAlbum(
	ctx context.Context,
	id uuid.UUID,
) (catalog.Album, error) {
	return queryOne[catalog.Album](ctx, r.dbtx, selectAlbumQuery, id)
}

// Titles are not unique, even per artist, so the oldest match wins.
const selectAlbumByArtistQuery = `
SELECT al.* FROM albums al
JOIN album_artists aa ON aa.album_id = al.id
WHERE lower(al.title) = lower($1) AND aa.artist_id = $2
ORDER BY al.created_at, al.id
LIMIT 1`

// FindAlbum looks up an album of artistID by title, case-insensitively.
func (r *Repository) FindAlbum(
	ctx context.Context,
	title string,
	artistID uuid.UUID,
) (catalog.Album, error) {
	return queryOne[catalog.Album](ctx, r.dbtx, selectAlbumByArtistQuery,
		title, artistID)
}

const insertAlbumArtistQuery = `
INSERT INTO album_artists (album_id, artist_id, position)
VALUES ($1, $2, $3)`

func (r *Repository) AddAlbumArtist(
	ctx context.Context,
	aa catalog.AlbumArtist,
) error {
	_, err := r.dbtx.Exec(ctx, insertAlbumArtistQuery,
		aa.AlbumID, aa.ArtistID, aa.Position)
	return err
}

const insertAlbumTrackQuery = `
INSERT INTO album_tracks (album_id, track_id, quality, disc_no, track_no)
VALUES ($1, $2, $3, $4, $5)`

func (r *Repository) AddAlbumTrack(
	ctx context.Context,
	at catalog.AlbumTrack,
) error {
	_, err := r.dbtx.Exec(ctx, insertAlbumTrackQuery,
		at.AlbumID, at.TrackID, at.Quality,
		at.DiscNo, at.TrackNo)
	return err
}
