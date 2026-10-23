//go:build integration

package service

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"audio-streaming/internal/adapters/postgres"
	"audio-streaming/internal/catalog"
	"audio-streaming/internal/env"
)

const (
	testMigrationsDir = "../../migrations"
	testImage         = "postgres:18.6-alpine"
)

var testPool = sync.OnceValues(startTestDB)

func startTestDB() (*pgxpool.Pool, error) {
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, testImage,
		tcpostgres.WithDatabase(env.MustGet("TEST_POSTGRES_DB")),
		tcpostgres.WithUsername(env.MustGet("TEST_POSTGRES_USER")),
		tcpostgres.WithPassword(env.MustGet("TEST_POSTGRES_PASSWORD")),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("starting postgres container (is docker running?): %w", err)
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("connecting: %w", err)
	}

	if _, err := postgres.Migrate(ctx, pool, testMigrationsDir); err != nil {
		pool.Close()
		_ = testcontainers.TerminateContainer(ctr)
		return nil, fmt.Errorf("migrating: %w", err)
	}
	return pool, nil
}

// newDBLibrary returns a Library whose transactions commit to the test
// database, and a repository for reading back what it committed. The
// catalog tables are emptied in t.Cleanup, so tests must not run in
// parallel.
func newDBLibrary(t *testing.T) (*Library, *postgres.Repository) {
	t.Helper()

	pool, err := testPool()
	require.NoError(t, err, "test database")
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(),
			`TRUNCATE images, artists, albums, tracks CASCADE`)
		if err != nil {
			t.Logf("truncate: %v", err)
		}
	})

	repo := postgres.NewRepository(pool)
	return NewLibrary(repo), repo
}

// selectRow reads back a single row with plain SQL, for tables the
// repository has no reader for.
func selectRow[T any](t *testing.T, sql string, args ...any) T {
	t.Helper()
	pool, err := testPool()
	require.NoError(t, err, "test database")
	rows, err := pool.Query(context.Background(), sql, args...)
	require.NoError(t, err, "Query")
	v, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	require.NoError(t, err, "CollectOneRow")
	return v
}

func countRows(t *testing.T, table string) int {
	t.Helper()
	pool, err := testPool()
	require.NoError(t, err, "test database")
	var n int
	err = pool.QueryRow(context.Background(), `SELECT count(*) FROM `+table).Scan(&n)
	require.NoError(t, err, "count %s", table)
	return n
}

func TestDBSaveCommitsTrack(t *testing.T) {
	lib, repo := newDBLibrary(t)
	ctx := context.Background()
	rec := testRecord("Duality")
	cover := testCover()

	track, err := lib.Save(ctx, rec, cover)

	require.NoError(t, err)
	gotTrack, err := repo.GetTrack(ctx, rec.TrackID)
	require.NoError(t, err, "GetTrack")
	artist, err := repo.FindArtistByName(ctx, "Slipknot")
	require.NoError(t, err, "FindArtistByName")
	byArtist, err := repo.FindTrack(ctx, "Duality", artist.ID)
	require.NoError(t, err, "FindTrack")
	album, err := repo.FindAlbum(ctx, "Vol. 3", artist.ID)
	require.NoError(t, err, "FindAlbum")
	require.NotNil(t, album.CoverImageID, "CoverImageID")
	img, err := repo.GetImage(ctx, *album.CoverImageID)
	require.NoError(t, err, "GetImage")
	albumTrack := selectRow[catalog.AlbumTrack](t,
		`SELECT * FROM album_tracks WHERE track_id = $1`, rec.TrackID)

	assert.Equal(t, track, gotTrack, "track")
	assert.Equal(t, rec.TrackID, byArtist.ID, "track artist")
	assert.Equal(t, ptr(int16(2004)), album.ReleaseYear, "release year")
	assert.Equal(t, catalog.Image{
		ID:             img.ID,
		StoragePath:    cover.StoragePath,
		OriginalFormat: cover.OriginalFormat,
		Width:          cover.Width,
		Height:         cover.Height,
		CreatedAt:      img.CreatedAt,
		UpdatedAt:      img.UpdatedAt,
	}, img, "cover")
	assert.Equal(t, catalog.AlbumTrack{
		AlbumID: album.ID,
		TrackID: rec.TrackID,
		Quality: catalog.QualityLossless,
		DiscNo:  1,
		TrackNo: 2,
	}, albumTrack, "album track")
}

func TestDBSaveReusesCommittedArtistAndAlbum(t *testing.T) {
	lib, _ := newDBLibrary(t)
	mustSave(t, lib, testRecord("Duality"))
	rec := testRecord("Pulse of the Maggots")
	rec.Tags.TrackNumber = 3

	_, err := lib.Save(context.Background(), rec, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, countRows(t, "artists"))
	assert.Equal(t, 1, countRows(t, "albums"))
	assert.Equal(t, 2, countRows(t, "album_tracks"))
}

func TestDBSaveRejectsCommittedTrack(t *testing.T) {
	lib, _ := newDBLibrary(t)
	mustSave(t, lib, testRecord("Duality"))
	rec := testRecord("DUALITY")
	rec.Tags.TrackNumber = 3

	_, err := lib.Save(context.Background(), rec, nil)

	require.ErrorIs(t, err, ErrTrackExists)
	assert.Equal(t, 1, countRows(t, "tracks"))
}

func TestDBSaveRollsBackOnFailure(t *testing.T) {
	lib, repo := newDBLibrary(t)
	mustSave(t, lib, testRecord("Duality"))
	// Same album and track number as Duality, so AddAlbumTrack fails
	// after the artist and track were already inserted.
	rec := testRecord("Before I Forget")
	rec.Tags.Artist = "Corey Taylor"
	rec.Tags.AlbumArtist = "Slipknot"

	_, err := lib.Save(context.Background(), rec, nil)

	require.Error(t, err)
	_, err = repo.GetTrack(context.Background(), rec.TrackID)
	require.ErrorIs(t, err, catalog.ErrNotFound, "track")
	_, err = repo.FindArtistByName(context.Background(), "Corey Taylor")
	assert.ErrorIs(t, err, catalog.ErrNotFound, "artist")
}
