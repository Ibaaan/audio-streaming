package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"audio-streaming/internal/catalog"
)

const insertTrackQuery = `
INSERT INTO tracks (id, title, duration_ms, quality)
VALUES ($1, $2, $3, $4)
RETURNING created_at, updated_at`

// AddTrack inserts t and fills in the timestamps the database sets.
func (r *Repository) AddTrack(ctx context.Context, t *catalog.Track) error {
	if t.Quality == "" {
		t.Quality = catalog.QualityLow
	}
	return r.dbtx.QueryRow(ctx, insertTrackQuery,
		t.ID, t.Title, t.DurationMs, t.Quality,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
}

const selectTrackQuery = `SELECT * FROM tracks WHERE id = $1`

func (r *Repository) GetTrack(ctx context.Context, id uuid.UUID) (catalog.Track, error) {
	return queryOne[catalog.Track](ctx, r.dbtx, selectTrackQuery, id)
}

const insertTrackArtistQuery = `
INSERT INTO track_artists (track_id, artist_id, role, position)
VALUES ($1, $2, $3, $4)`

func (r *Repository) AddTrackArtist(ctx context.Context, ta catalog.TrackArtist) error {
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
	return queryOne[catalog.Track](ctx, r.dbtx, selectTrackByArtistQuery, title, artistID)
}

// A track with several artists is listed with the first by position. A
// track on several albums is listed with the oldest of them.
const listTracksQuery = `
SELECT t.id AS track_id, t.title AS track_title,
       t.duration_ms AS track_duration_ms, t.quality AS track_quality,
       t.created_at AS track_created_at, t.updated_at AS track_updated_at,
       ar.id AS artist_id, ar.name AS artist_name,
       ar.sort_name AS artist_sort_name, ar.image_id AS artist_image_id,
       ar.created_at AS artist_created_at, ar.updated_at AS artist_updated_at,
       al.id AS album_id, al.title AS album_title,
       al.album_type, al.release_year, al.cover_image_id,
       al.created_at AS album_created_at, al.updated_at AS album_updated_at
FROM tracks t
LEFT JOIN LATERAL (
    SELECT a.*
    FROM track_artists ta
    JOIN artists a ON a.id = ta.artist_id
    WHERE ta.track_id = t.id
    ORDER BY ta.position, a.created_at, a.id
    LIMIT 1
) ar ON true
LEFT JOIN LATERAL (
    SELECT a.*
    FROM album_tracks at
    JOIN albums a ON a.id = at.album_id
    WHERE at.track_id = t.id
    ORDER BY a.created_at, a.id
    LIMIT 1
) al ON true
ORDER BY t.created_at DESC, t.id DESC`

type trackSummaryRow struct {
	TrackQuality    catalog.Quality `db:"track_quality"`
	TrackID         uuid.UUID       `db:"track_id"`
	TrackTitle      string          `db:"track_title"`
	TrackDurationMs int32           `db:"track_duration_ms"`
	TrackCreatedAt  time.Time       `db:"track_created_at"`
	TrackUpdatedAt  time.Time       `db:"track_updated_at"`

	ArtistID        *uuid.UUID `db:"artist_id"`
	ArtistName      *string    `db:"artist_name"`
	ArtistSortName  *string    `db:"artist_sort_name"`
	ArtistImageID   *uuid.UUID `db:"artist_image_id"`
	ArtistCreatedAt *time.Time `db:"artist_created_at"`
	ArtistUpdatedAt *time.Time `db:"artist_updated_at"`

	AlbumID        *uuid.UUID         `db:"album_id"`
	AlbumTitle     *string            `db:"album_title"`
	AlbumType      *catalog.AlbumType `db:"album_type"`
	ReleaseYear    *int16             `db:"release_year"`
	CoverImageID   *uuid.UUID         `db:"cover_image_id"`
	AlbumCreatedAt *time.Time         `db:"album_created_at"`
	AlbumUpdatedAt *time.Time         `db:"album_updated_at"`
}

// ListTracks returns every track with its artists and album, newest
// first.
func (r *Repository) ListTracks(ctx context.Context) ([]catalog.TrackSummary, error) {
	rows, err := r.dbtx.Query(ctx, listTracksQuery)
	if err != nil {
		return nil, err
	}
	dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[trackSummaryRow])
	if err != nil {
		return nil, err
	}

	tracks := make([]catalog.TrackSummary, 0, len(dbRows))
	for _, row := range dbRows {
		summary := catalog.TrackSummary{Track: catalog.Track{
			ID:         row.TrackID,
			Quality:    row.TrackQuality,
			Title:      row.TrackTitle,
			DurationMs: row.TrackDurationMs,
			CreatedAt:  row.TrackCreatedAt,
			UpdatedAt:  row.TrackUpdatedAt,
		}}
		if row.ArtistID != nil {
			summary.Artist = catalog.Artist{
				ID:        *row.ArtistID,
				Name:      *row.ArtistName,
				SortName:  row.ArtistSortName,
				ImageID:   row.ArtistImageID,
				CreatedAt: *row.ArtistCreatedAt,
				UpdatedAt: *row.ArtistUpdatedAt,
			}
		}
		if row.AlbumID != nil {
			summary.Album = &catalog.Album{
				ID:           *row.AlbumID,
				Title:        *row.AlbumTitle,
				AlbumType:    *row.AlbumType,
				ReleaseYear:  row.ReleaseYear,
				CoverImageID: row.CoverImageID,
				CreatedAt:    *row.AlbumCreatedAt,
				UpdatedAt:    *row.AlbumUpdatedAt,
			}
		}
		tracks = append(tracks, summary)
	}
	return tracks, nil
}
