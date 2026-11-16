package postgres

import (
	"context"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

// Read metadata, credits, and track positions in one database snapshot.
const albumDetailsQuery = `
SELECT a.*,
 (SELECT coalesce(jsonb_agg(ar ORDER BY aa.position), '[]'::jsonb)
  FROM album_artists aa JOIN artists ar ON ar.id = aa.artist_id
  WHERE aa.album_id = a.id) AS artists,
 (SELECT coalesce(jsonb_agg(
   to_jsonb(t) || jsonb_build_object(
    'disc_no', at.disc_no, 'track_no', at.track_no,
    'artists', (SELECT coalesce(jsonb_agg(
      to_jsonb(ar) || jsonb_build_object('role', ta.role) ORDER BY ta.position
     ), '[]'::jsonb)
     FROM track_artists ta JOIN artists ar ON ar.id = ta.artist_id
     WHERE ta.track_id = t.id)
   ) ORDER BY at.disc_no, at.track_no
  ), '[]'::jsonb)
  FROM album_tracks at JOIN tracks t ON t.id = at.track_id
  WHERE at.album_id = a.id) AS tracks
FROM albums a WHERE a.id = $1`

func (r *Repository) GetAlbumDetails(
	ctx context.Context, id uuid.UUID,
) (catalog.AlbumDetails, error) {
	return queryOne[catalog.AlbumDetails](ctx, r.dbtx, albumDetailsQuery, id)
}
