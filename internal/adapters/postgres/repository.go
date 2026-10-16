package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"audio-streaming/internal/catalog"
)

// Repository runs catalog queries against whatever it wraps: a pool, a
// single connection or a transaction. Its methods are split by entity
// across artists.go, albums.go, images.go and tracks.go.
type Repository struct {
	dbtx DBTX
}

func NewRepository(dbtx DBTX) *Repository {
	return &Repository{dbtx: dbtx}
}

// WithinTx runs fn with a repository bound to a new transaction and
// commits if fn returns nil. On a repository that is already inside a
// transaction it opens a savepoint instead.
func (r *Repository) WithinTx(
	ctx context.Context,
	fn func(tx *Repository) error,
) error {
	return WithTransaction(ctx, r.dbtx, func(tx DBTX) error {
		return fn(NewRepository(tx))
	})
}

// queryOne runs sql and maps the single row it returns onto T by T's db
// tags. Inserts use it too, through RETURNING *. No row is reported as
// catalog.ErrNotFound.
func queryOne[T any](
	ctx context.Context,
	dbtx DBTX,
	sql string,
	args ...any,
) (T, error) {
	rows, err := dbtx.Query(ctx, sql, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	v, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if errors.Is(err, pgx.ErrNoRows) {
		return v, catalog.ErrNotFound
	}
	return v, err
}
