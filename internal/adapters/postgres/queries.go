package postgres

import (
	"context"

	"github.com/google/uuid"
)

var INSERT_TRACK_QUERY string = `
INSERT INTO tracks (id, title, duration_ms, status)
VALUES ($1, $2, $3, $4)
RETURNING created_at, updated_at`

func AddTrack(
	ctx context.Context,
	dbtx DBTX,
	title string,
	durationMs int32,
) (Track, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Track{}, err
	}

	track := Track{
		ID:         id,
		Title:      title,
		DurationMs: durationMs,
		Status:     StatusReady,
	}

	err = dbtx.QueryRow(
		ctx,
		INSERT_TRACK_QUERY,
		track.ID,
		track.Title,
		track.DurationMs,
		track.Status,
	).Scan(&track.CreatedAt, &track.UpdatedAt)
	if err != nil {
		return Track{}, err
	}
	return track, nil
}
