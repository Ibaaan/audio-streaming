package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"audio-streaming/internal/catalog"
)

func (r *Repository) GetArtistStats(
	ctx context.Context, id uuid.UUID,
) (catalog.ArtistStats, error) {
	return queryOne[catalog.ArtistStats](ctx, r.dbtx, `
SELECT (SELECT count(*) FROM album_artists WHERE artist_id = a.id) AS album_count,
       (SELECT count(*) FROM track_artists WHERE artist_id = a.id) AS track_count
FROM artists a WHERE a.id = $1`, id)
}

func (r *Repository) ListArtistAlbums(
	ctx context.Context, id uuid.UUID, limit, offset int,
) ([]catalog.AlbumOverview, int, error) {
	var total int
	err := r.dbtx.QueryRow(ctx,
		`SELECT count(*) FROM album_artists WHERE artist_id = $1`, id,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.dbtx.Query(ctx, `
SELECT `+albumOverviewColumns+`
FROM albums a JOIN album_artists credit ON credit.album_id = a.id
WHERE credit.artist_id = $1
ORDER BY a.release_year DESC NULLS LAST, a.id
LIMIT $2 OFFSET $3`, id, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	albums, err := pgx.CollectRows(rows, pgx.RowToStructByName[catalog.AlbumOverview])
	return albums, total, err
}
