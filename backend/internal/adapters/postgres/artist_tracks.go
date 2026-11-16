package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"audio-streaming/internal/catalog"
)

// ListArtistTracks includes credits in any role and counts before pagination.
// The oldest album is primary, consistently with GetAlbumByTrackId.
func (r *Repository) ListArtistTracks(
	ctx context.Context, artistID uuid.UUID, limit, offset int,
) ([]catalog.ArtistTrack, int, error) {
	var total int
	err := r.dbtx.QueryRow(ctx,
		`SELECT count(*) FROM track_artists WHERE artist_id = $1`, artistID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.dbtx.Query(ctx, `
SELECT `+trackOverviewColumns+`, ta.role`+trackOverviewFrom+`
JOIN track_artists ta ON ta.track_id = t.id
WHERE ta.artist_id = $1
ORDER BY al.release_year DESC NULLS LAST,
         al.disc_no ASC NULLS LAST, al.track_no ASC NULLS LAST, t.id
LIMIT $2 OFFSET $3`, artistID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	tracks, err := pgx.CollectRows(rows, pgx.RowToStructByName[catalog.ArtistTrack])
	return tracks, total, err
}
