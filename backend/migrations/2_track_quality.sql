-- Quality belongs to the stored track, including tracks without an album.
ALTER TABLE tracks ADD COLUMN quality text NOT NULL DEFAULT 'low'
  CHECK (quality IN ('low', 'optimal', 'lossless'));

-- Preserve existing classifications; original codec/bitrate was not persisted.
-- Albumless legacy tracks remain low until their files are re-probed.
UPDATE tracks t SET quality = q.quality
FROM (
 SELECT track_id,
        CASE max(CASE quality WHEN 'lossless' THEN 3 WHEN 'optimal' THEN 2 ELSE 1 END)
          WHEN 3 THEN 'lossless' WHEN 2 THEN 'optimal' ELSE 'low' END AS quality
 FROM album_tracks GROUP BY track_id
) q WHERE t.id = q.track_id;

-- Keep legacy album credits consistent with the one track-level classification.
UPDATE album_tracks at SET quality = t.quality FROM tracks t WHERE t.id = at.track_id;
