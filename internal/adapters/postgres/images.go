package postgres

import (
	"context"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

const insertImageQuery = `
INSERT INTO images (id, storage_path, original_format, width, height)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at, updated_at`

// AddImage inserts img and fills in the timestamps the database sets.
func (r *Repository) AddImage(ctx context.Context, img *catalog.Image) error {
	return r.dbtx.QueryRow(ctx, insertImageQuery,
		img.ID, img.StoragePath, img.OriginalFormat,
		img.Width, img.Height,
	).Scan(&img.CreatedAt, &img.UpdatedAt)
}

const selectImageQuery = `SELECT * FROM images WHERE id = $1`

func (r *Repository) GetImage(
	ctx context.Context,
	id uuid.UUID,
) (catalog.Image, error) {
	return queryOne[catalog.Image](ctx, r.dbtx, selectImageQuery, id)
}

const selectImageByPathQuery = `
SELECT * FROM images WHERE storage_path = $1`

// FindImageByPath looks up an image by its unique storage path.
func (r *Repository) FindImageByPath(
	ctx context.Context,
	storagePath string,
) (catalog.Image, error) {
	return queryOne[catalog.Image](ctx, r.dbtx, selectImageByPathQuery,
		storagePath)
}
