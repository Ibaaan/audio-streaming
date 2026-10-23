//go:build integration

package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/env"
)

const (
	testMigrationsDir = "../../../migrations"
	testImage         = "postgres:18.6-alpine"
)

var testConn = sync.OnceValues(startTestDB)

func startTestDB() (*pgx.Conn, error) {
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

// newTestRepo wraps a transaction on the shared test connection and rolls
// it back in t.Cleanup, so tests don't see each other's rows.
func newTestRepo(t *testing.T) *Repository {
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
	return NewRepository(tx)
}

// selectOne reads back a single row with plain SQL, independent of the
// repository's own readers.
func selectOne[T any](t *testing.T, repo *Repository, sql string, args ...any) T {
	t.Helper()
	rows, err := repo.dbtx.Query(context.Background(), sql, args...)
	require.NoError(t, err, "Query")
	v, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	require.NoError(t, err, "CollectOneRow")
	return v
}

func newID(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	require.NoError(t, err, "NewV7")
	return id
}

func mustAddTrack(t *testing.T, repo catalog.TrackRepository) catalog.Track {
	t.Helper()
	track := catalog.Track{
		ID:         newID(t),
		Title:      "Duality",
		DurationMs: 252_000,
	}
	require.NoError(t, repo.AddTrack(context.Background(), &track), "AddTrack")
	return track
}

func mustAddArtist(t *testing.T, repo *Repository, name string) catalog.Artist {
	t.Helper()
	artist := catalog.Artist{ID: newID(t), Name: name}
	require.NoError(t, repo.AddArtist(context.Background(), &artist), "AddArtist")
	return artist
}

func mustAddAlbum(t *testing.T, repo *Repository, title string) catalog.Album {
	t.Helper()
	album := catalog.Album{
		ID:        newID(t),
		Title:     title,
		AlbumType: catalog.AlbumTypeAlbum,
	}
	require.NoError(t, repo.AddAlbum(context.Background(), &album), "AddAlbum")
	return album
}

func mustAddAlbumArtist(
	t *testing.T,
	repo *Repository,
	album catalog.Album,
	artist catalog.Artist,
) {
	t.Helper()
	albumArtist := catalog.AlbumArtist{
		AlbumID:  album.ID,
		ArtistID: artist.ID,
		Position: 0,
	}
	err := repo.AddAlbumArtist(context.Background(), albumArtist)
	require.NoError(t, err, "AddAlbumArtist")
}

func mustAddImage(t *testing.T, repo *Repository) catalog.Image {
	t.Helper()
	image := catalog.Image{
		ID:             newID(t),
		StoragePath:    "covers/" + uuid.NewString() + ".jpg",
		OriginalFormat: catalog.ImageFormatJPEG,
	}
	require.NoError(t, repo.AddImage(context.Background(), &image), "AddImage")
	return image
}

func ptr[T any](v T) *T { return &v }
