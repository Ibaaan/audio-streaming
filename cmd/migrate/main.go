// Command migrate applies every pending SQL migration in a directory to
// the PostgreSQL database.
//
// Migration files are named <version>_<name>.sql (e.g. 1_init_model.sql)
// and applied in ascending version order. Each file runs in its own
// transaction together with the bookkeeping insert into
// schema_migrations, so files must not contain BEGIN/COMMIT.
//
// Usage: migrate [-dir migrations]
//
// The connection string is read from DATABASE_URL.
package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Matches postgres.env used by docker-compose.yml.
const defaultDatabaseURL = "postgres://audio:audio@localhost:5432/audio" +
	"?sslmode=disable"

// Arbitrary key shared by every migrate run, so concurrent runs
// serialize instead of applying the same file twice.
const advisoryLockKey = 7_246_183_001

type migration struct {
	version int64
	name    string
	path    string
}

func main() {
	dir := flag.String("dir", "migrations", "directory with .sql files")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, *dir); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dir string) error {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = defaultDatabaseURL
	}

	migrations, err := loadMigrations(dir)
	if err != nil {
		return err
	}

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = conn.Close(context.Background()) }()

	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    bigint      PRIMARY KEY,
		name       text        NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := 0
	for _, m := range migrations {
		ok, err := apply(ctx, conn, m)
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(m.path), err)
		}
		if ok {
			fmt.Println("applied", filepath.Base(m.path))
			applied++
		}
	}

	fmt.Printf("%d migration(s) applied, database up to date\n", applied)

	return nil
}

// loadMigrations returns the .sql files in dir sorted by version.
func loadMigrations(dir string) ([]migration, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no .sql files in %q", dir)
	}

	migrations := make([]migration, 0, len(paths))
	seen := make(map[int64]string, len(paths))
	for _, p := range paths {
		base := strings.TrimSuffix(filepath.Base(p), ".sql")
		prefix, name, _ := strings.Cut(base, "_")

		version, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"%s: name must start with <version>_", filepath.Base(p),
			)
		}
		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf(
				"version %d used by both %s and %s",
				version, other, filepath.Base(p),
			)
		}
		seen[version] = filepath.Base(p)

		migrations = append(migrations, migration{version, name, p})
	}

	slices.SortFunc(migrations, func(a, b migration) int {
		return cmp.Compare(a.version, b.version)
	})

	return migrations, nil
}

// apply runs m unless it is already recorded in schema_migrations.
// It reports whether the migration was applied by this call.
func apply(
	ctx context.Context, conn *pgx.Conn, m migration,
) (bool, error) {
	sql, err := os.ReadFile(m.path)
	if err != nil {
		return false, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return false, err
	}
	// No-op once the transaction is committed.
	defer func() { _ = tx.Rollback(context.Background()) }()

	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", advisoryLockKey)
	if err != nil {
		return false, fmt.Errorf("lock: %w", err)
	}

	var exists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)",
		m.version,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return false, tx.Commit(ctx)
	}

	// No arguments: pgx uses the simple protocol, which allows the
	// file to contain multiple statements.
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
		m.version, m.name,
	)
	if err != nil {
		return false, err
	}

	return true, tx.Commit(ctx)
}
