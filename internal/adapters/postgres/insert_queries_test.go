//go:build integration

package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testMigrationsDir = "../../../migrations"
	testImage         = "postgres:18.6-alpine"
)

var testConn = sync.OnceValues(startTestDB)

func startTestDB() (*pgx.Conn, error) {
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, testImage,
		tcpostgres.WithDatabase("audio_test"),
		tcpostgres.WithUsername("audio"),
		tcpostgres.WithPassword("audio"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("starting postgres container "+
			"(is docker running?): %w", err)
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("connection string: %w", err)
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("connecting: %w", err)
	}

	if _, err := Migrate(ctx, conn, testMigrationsDir); err != nil {
		_ = conn.Close(ctx)
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("migrating: %w", err)
	}
	return conn, nil
}

// newTestTx opens a transaction on the shared test connection and rolls
// it back in t.Cleanup, so tests don't see each other's rows.
func newTestTx(t *testing.T) pgx.Tx {
	t.Helper()

	conn, err := testConn()
	require.NoError(t, err, "test database")

	tx, err := conn.Begin(context.Background())
	require.NoError(t, err, "Begin")
	t.Cleanup(func() {
		if err := tx.Rollback(context.Background()); err != nil {
			t.Logf("rollback: %v", err)
		}
	})
	return tx
}

// selectOne reads back a single row and maps it onto T by its db tags.
func selectOne[T any](
	t *testing.T,
	dbtx DBTX,
	sql string,
	args ...any,
) T {
	t.Helper()
	rows, err := dbtx.Query(context.Background(), sql, args...)
	require.NoError(t, err, "Query")
	v, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	require.NoError(t, err, "CollectOneRow")
	return v
}

func mustAddTrack(t *testing.T, dbtx DBTX) Track {
	t.Helper()
	track, err := AddTrack(context.Background(), dbtx, "Duality", 252_000)
	require.NoError(t, err, "AddTrack")
	return track
}

func mustAddArtist(t *testing.T, dbtx DBTX) Artist {
	t.Helper()
	artist, err := AddArtist(context.Background(), dbtx, "Slipknot",
		nil, nil)
	require.NoError(t, err, "AddArtist")
	return artist
}

func mustAddAlbum(t *testing.T, dbtx DBTX) Album {
	t.Helper()
	album, err := AddAlbum(context.Background(), dbtx,
		"Vol. 3: (The Subliminal Verses)", AlbumTypeAlbum, nil, nil)
	require.NoError(t, err, "AddAlbum")
	return album
}

func mustAddImage(t *testing.T, dbtx DBTX) Image {
	t.Helper()
	image, err := AddImage(context.Background(), dbtx,
		"covers/"+uuid.NewString()+".jpg", ImageFormatJPEG, nil, nil)
	require.NoError(t, err, "AddImage")
	return image
}

func ptr[T any](v T) *T { return &v }

func TestAddTrack(t *testing.T) {
	tx := newTestTx(t)

	track, err := AddTrack(context.Background(), tx, "Duality", 252_000)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, track.ID)
	assert.False(t, track.CreatedAt.IsZero(), "CreatedAt set")
	got := selectOne[Track](t, tx,
		`SELECT * FROM tracks WHERE id = $1`, track.ID)
	assert.Equal(t, track, got)
}

func TestAddTrackRejectsZeroDuration(t *testing.T) {
	tx := newTestTx(t)

	_, err := AddTrack(context.Background(), tx, "Duality", 0)

	assert.Error(t, err)
}

func TestAddImage(t *testing.T) {
	tx := newTestTx(t)

	image, err := AddImage(context.Background(), tx, "covers/a.png",
		ImageFormatPNG, ptr[int32](600), ptr[int32](400))

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, image.ID)
	got := selectOne[Image](t, tx,
		`SELECT * FROM images WHERE id = $1`, image.ID)
	assert.Equal(t, image, got)
}

func TestAddImageRejectsDuplicateStoragePath(t *testing.T) {
	tx := newTestTx(t)
	existing := mustAddImage(t, tx)

	_, err := AddImage(context.Background(), tx, existing.StoragePath,
		ImageFormatJPEG, nil, nil)

	assert.Error(t, err)
}

func TestAddArtist(t *testing.T) {
	tx := newTestTx(t)
	image := mustAddImage(t, tx)

	artist, err := AddArtist(context.Background(), tx, "The Beatles",
		ptr("Beatles, The"), &image.ID)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, artist.ID)
	got := selectOne[Artist](t, tx,
		`SELECT * FROM artists WHERE id = $1`, artist.ID)
	assert.Equal(t, artist, got)
}

func TestAddArtistWithoutOptionalFields(t *testing.T) {
	tx := newTestTx(t)

	artist, err := AddArtist(context.Background(), tx, "Slipknot",
		nil, nil)

	require.NoError(t, err)
	got := selectOne[Artist](t, tx,
		`SELECT * FROM artists WHERE id = $1`, artist.ID)
	assert.Nil(t, got.SortName)
	assert.Nil(t, got.ImageID)
}

func TestAddArtistRejectsBlankName(t *testing.T) {
	tx := newTestTx(t)

	_, err := AddArtist(context.Background(), tx, "   ", nil, nil)

	assert.Error(t, err)
}

func TestAddAlbum(t *testing.T) {
	tx := newTestTx(t)
	cover := mustAddImage(t, tx)

	album, err := AddAlbum(context.Background(), tx, "Iowa",
		AlbumTypeAlbum, ptr[int16](2001), &cover.ID)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, album.ID)
	got := selectOne[Album](t, tx,
		`SELECT * FROM albums WHERE id = $1`, album.ID)
	assert.Equal(t, album, got)
}

func TestAddAlbumRejectsOutOfRangeYear(t *testing.T) {
	tx := newTestTx(t)

	_, err := AddAlbum(context.Background(), tx, "Iowa",
		AlbumTypeAlbum, ptr[int16](99), nil)

	assert.Error(t, err)
}

func TestAddAlbumArtist(t *testing.T) {
	tx := newTestTx(t)
	album := mustAddAlbum(t, tx)
	artist := mustAddArtist(t, tx)

	albumArtist, err := AddAlbumArtist(context.Background(), tx,
		album.ID, artist.ID, 0)

	require.NoError(t, err)
	got := selectOne[AlbumArtist](t, tx,
		`SELECT * FROM album_artists WHERE album_id = $1`, album.ID)
	assert.Equal(t, albumArtist, got)
}

func TestAddAlbumArtistRejectsUnknownAlbum(t *testing.T) {
	tx := newTestTx(t)
	artist := mustAddArtist(t, tx)

	_, err := AddAlbumArtist(context.Background(), tx,
		uuid.New(), artist.ID, 0)

	assert.Error(t, err)
}

func TestAddTrackArtist(t *testing.T) {
	tx := newTestTx(t)
	track := mustAddTrack(t, tx)
	artist := mustAddArtist(t, tx)

	trackArtist, err := AddTrackArtist(context.Background(), tx,
		track.ID, artist.ID, ptr("vocals"), 0)

	require.NoError(t, err)
	got := selectOne[TrackArtist](t, tx,
		`SELECT * FROM track_artists WHERE track_id = $1`, track.ID)
	assert.Equal(t, trackArtist, got)
}

func TestAddTrackArtistRejectsDuplicatePosition(t *testing.T) {
	tx := newTestTx(t)
	track := mustAddTrack(t, tx)
	first := mustAddArtist(t, tx)
	second := mustAddArtist(t, tx)
	_, err := AddTrackArtist(context.Background(), tx,
		track.ID, first.ID, nil, 0)
	require.NoError(t, err, "first AddTrackArtist")

	_, err = AddTrackArtist(context.Background(), tx,
		track.ID, second.ID, nil, 0)

	assert.Error(t, err)
}

func TestAddAlbumTrack(t *testing.T) {
	tx := newTestTx(t)
	album := mustAddAlbum(t, tx)
	track := mustAddTrack(t, tx)

	albumTrack, err := AddAlbumTrack(context.Background(), tx,
		album.ID, track.ID, QualityLossless, 1, 1)

	require.NoError(t, err)
	got := selectOne[AlbumTrack](t, tx,
		`SELECT * FROM album_tracks WHERE album_id = $1`, album.ID)
	assert.Equal(t, albumTrack, got)
}

func TestAddAlbumTrackRejectsDuplicateTrackNumber(t *testing.T) {
	tx := newTestTx(t)
	album := mustAddAlbum(t, tx)
	first := mustAddTrack(t, tx)
	second := mustAddTrack(t, tx)
	_, err := AddAlbumTrack(context.Background(), tx,
		album.ID, first.ID, QualityLossless, 1, 1)
	require.NoError(t, err, "first AddAlbumTrack")

	_, err = AddAlbumTrack(context.Background(), tx,
		album.ID, second.ID, QualityLossless, 1, 1)

	assert.Error(t, err)
}
