package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"audio-streaming/internal/catalog"
)

// Parameters are literal substrings; SQL wildcard characters have no special meaning.
func searchRows[T any](
	ctx context.Context, db DBTX, countQuery, listQuery, query string, limit, offset int,
) ([]T, int, error) {
	var total int
	if err := db.QueryRow(ctx, countQuery, query).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := db.Query(ctx, listQuery, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	return result, total, err
}

func (r *Repository) SearchArtists(
	ctx context.Context, query string, limit, offset int,
) ([]catalog.Artist, int, error) {
	return searchRows[catalog.Artist](ctx, r.dbtx,
		`SELECT count(*) FROM artists WHERE strpos(lower(name), lower($1)) > 0`,
		`SELECT * FROM artists WHERE strpos(lower(name), lower($1)) > 0
 ORDER BY lower(name), id LIMIT $2 OFFSET $3`, query, limit, offset)
}

func (r *Repository) SearchAlbums(
	ctx context.Context, query string, limit, offset int,
) ([]catalog.Album, int, error) {
	return searchRows[catalog.Album](ctx, r.dbtx,
		`SELECT count(*) FROM albums WHERE strpos(lower(title), lower($1)) > 0`,
		`SELECT * FROM albums WHERE strpos(lower(title), lower($1)) > 0
 ORDER BY lower(title), id LIMIT $2 OFFSET $3`, query, limit, offset)
}

func (r *Repository) SearchTracks(
	ctx context.Context, query string, limit, offset int,
) ([]catalog.Track, int, error) {
	return searchRows[catalog.Track](ctx, r.dbtx,
		`SELECT count(*) FROM tracks WHERE strpos(lower(title), lower($1)) > 0`,
		`SELECT * FROM tracks WHERE strpos(lower(title), lower($1)) > 0
 ORDER BY lower(title), id LIMIT $2 OFFSET $3`, query, limit, offset)
}
