package postgres

// Shared list projections aggregate credits before joining, preventing duplicate
// tracks or albums when a resource has several credits or appearances.
const albumOverviewColumns = `a.*,
 (SELECT count(*) FROM album_tracks at WHERE at.album_id = a.id) AS track_count,
 (SELECT coalesce(jsonb_agg(ar ORDER BY aa.position), '[]'::jsonb)
  FROM album_artists aa JOIN artists ar ON ar.id = aa.artist_id
  WHERE aa.album_id = a.id) AS artists`

const trackOverviewColumns = `t.*,
 (SELECT coalesce(jsonb_agg(
   to_jsonb(ar) || jsonb_build_object('role', credit.role) ORDER BY credit.position
  ), '[]'::jsonb)
  FROM track_artists credit JOIN artists ar ON ar.id = credit.artist_id
  WHERE credit.track_id = t.id) AS artists,
 CASE WHEN al.id IS NULL THEN NULL ELSE to_jsonb(al) END AS album`

// The oldest album remains primary, with ID breaking creation-time ties.
const trackOverviewFrom = ` FROM tracks t
LEFT JOIN LATERAL (
 SELECT a.*, at.disc_no, at.track_no
 FROM album_tracks at JOIN albums a ON a.id = at.album_id
 WHERE at.track_id = t.id
 ORDER BY a.created_at, a.id LIMIT 1
) al ON true`
