package postgres

import (
	"context"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

const insertTrackQuery = `
INSERT INTO tracks (id, title, duration_ms)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at`

// AddTrack inserts t and fills in the timestamps the database sets.
func (r *Repository) AddTrack(ctx context.Context, t *catalog.Track) error {
	return r.dbtx.QueryRow(ctx, insertTrackQuery,
		t.ID, t.Title, t.DurationMs,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
}

const selectTrackQuery = `SELECT * FROM tracks WHERE id = $1`

func (r *Repository) GetTrack(
	ctx context.Context,
	id uuid.UUID,
) (catalog.Track, error) {
	return queryOne[catalog.Track](ctx, r.dbtx, selectTrackQuery, id)
}

const insertTrackArtistQuery = `
INSERT INTO track_artists (track_id, artist_id, role, position)
VALUES ($1, $2, $3, $4)`

func (r *Repository) AddTrackArtist(
	ctx context.Context,
	ta catalog.TrackArtist,
) error {
	_, err := r.dbtx.Exec(ctx, insertTrackArtistQuery,
		ta.TrackID, ta.ArtistID, ta.Role, ta.Position)
	return err
}

// Titles are not unique, even per artist, so the oldest match wins.
const selectTrackByArtistQuery = `
SELECT t.* FROM tracks t
JOIN track_artists ta ON ta.track_id = t.id
WHERE lower(t.title) = lower($1) AND ta.artist_id = $2
ORDER BY t.created_at, t.id
LIMIT 1`

// FindTrack looks up a track of artistID by title, case-insensitively.
func (r *Repository) FindTrack(
	ctx context.Context,
	title string,
	artistID uuid.UUID,
) (catalog.Track, error) {
	return queryOne[catalog.Track](ctx, r.dbtx, selectTrackByArtistQuery,
		title, artistID)
}
