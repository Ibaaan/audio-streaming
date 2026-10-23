package postgres

import (
	"context"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

const insertArtistQuery = `
INSERT INTO artists (id, name, sort_name, image_id)
VALUES ($1, $2, $3, $4)
RETURNING created_at, updated_at`

// AddArtist inserts a and fills in the timestamps the database sets.
func (r *Repository) AddArtist(ctx context.Context, a *catalog.Artist) error {
	return r.dbtx.QueryRow(ctx, insertArtistQuery,
		a.ID, a.Name, a.SortName, a.ImageID,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

const selectArtistQuery = `SELECT * FROM artists WHERE id = $1`

func (r *Repository) GetArtist(ctx context.Context, id uuid.UUID) (catalog.Artist, error) {
	return queryOne[catalog.Artist](ctx, r.dbtx, selectArtistQuery, id)
}

// Names are not unique, so the oldest match wins. lower(name) is
// indexed by artists_name_lower.
const selectArtistByNameQuery = `
SELECT * FROM artists
WHERE lower(name) = lower($1)
ORDER BY created_at, id
LIMIT 1`

// FindArtistByName matches name case-insensitively.
func (r *Repository) FindArtistByName(ctx context.Context, name string) (catalog.Artist, error) {
	return queryOne[catalog.Artist](ctx, r.dbtx, selectArtistByNameQuery, name)
}
