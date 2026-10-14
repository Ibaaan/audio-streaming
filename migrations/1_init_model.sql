CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-----------------------------------------------------------------------------
CREATE TABLE images (
  id               uuid        PRIMARY KEY,
  storage_path     text        NOT NULL UNIQUE,
  original_format  text        NOT NULL,
  width            int         CHECK (width  > 0),
  height           int         CHECK (height > 0),
  status           text        NOT NULL CHECK (status IN ('pending','processing','ready','failed')),
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

-----------------------------------------------------------------------------
CREATE TABLE artists (
  id             uuid        PRIMARY KEY,
  name           text        NOT NULL CHECK (btrim(name) <> ''),
  sort_name      text,                                   -- "Beatles, The"
  image_id       uuid        REFERENCES images(id) ON DELETE SET NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX artists_name_lower ON artists (lower(name));

-----------------------------------------------------------------------------
CREATE TABLE albums (
  id                      uuid        PRIMARY KEY,
  title                   text        NOT NULL CHECK (btrim(title) <> ''),
  album_type              text        NOT NULL,
  release_year            smallint    CHECK (release_year BETWEEN 1000 AND 9999),
  cover_image_id          uuid        REFERENCES images(id) ON DELETE SET NULL,
  created_at              timestamptz NOT NULL DEFAULT now(),
  updated_at              timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX albums_release_year ON albums (release_year);

CREATE TABLE album_artists (
  album_id   uuid     NOT NULL REFERENCES albums(id)  ON DELETE CASCADE,
  artist_id  uuid     NOT NULL REFERENCES artists(id) ON DELETE RESTRICT,
  position   smallint NOT NULL CHECK (position >= 0),
  PRIMARY KEY (album_id, artist_id),
  UNIQUE (album_id, position)
);
CREATE INDEX album_artists_artist_id ON album_artists (artist_id);


-----------------------------------------------------------------------------
CREATE TABLE tracks (
  id           uuid        PRIMARY KEY,
  title        text        NOT NULL CHECK (btrim(title) <> ''),
  duration_ms  int         NOT NULL CHECK (duration_ms > 0),
  status       text        NOT NULL CHECK (status IN ('pending','ready','hidden','failed')),
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tracks_status     ON tracks (status) WHERE status <> 'ready';

CREATE TABLE track_artists (
  track_id   uuid     NOT NULL REFERENCES tracks(id)  ON DELETE CASCADE,
  artist_id  uuid     NOT NULL REFERENCES artists(id) ON DELETE RESTRICT,
  role       text,
  position   smallint NOT NULL CHECK (position >= 0),
  PRIMARY KEY (track_id, artist_id),
  UNIQUE (track_id, position)
);
CREATE INDEX track_artists_artist_id ON track_artists (artist_id, role);

CREATE TABLE album_tracks (
  album_id  uuid     NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
  track_id  uuid     NOT NULL REFERENCES tracks(id) ON DELETE RESTRICT,
  quality   text     NOT NULL CHECK (quality IN ('low','optimal','lossless')),
  disc_no   smallint NOT NULL CHECK (disc_no  > 0),
  track_no  smallint NOT NULL           CHECK (track_no > 0),
  PRIMARY KEY (album_id, track_id),
  UNIQUE (album_id, disc_no, track_no)
);
CREATE INDEX album_tracks_track_id ON album_tracks (track_id);

-----------------------------------------------------------------------------
CREATE TRIGGER images_updated_at      BEFORE UPDATE ON images      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER artists_updated_at     BEFORE UPDATE ON artists     FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER albums_updated_at      BEFORE UPDATE ON albums      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER tracks_updated_at      BEFORE UPDATE ON tracks      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
