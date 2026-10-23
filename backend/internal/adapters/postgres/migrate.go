package postgres

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Arbitrary key shared by every Migrate call, so concurrent runs
// serialize instead of applying the same file twice.
const advisoryLockKey = 7_246_183_001

// Migrator is what Migrate needs: plain statements plus transactions.
// Both *pgx.Conn and *pgxpool.Pool satisfy it.
type Migrator interface {
	DBTX
	TxBeginner
}

type migration struct {
	version int64
	name    string
	path    string
}

// Migrate applies every pending SQL migration in dir and returns the file
// names it applied, in order.
//
// Migration files are named <version>_<name>.sql (e.g. 1_init_model.sql)
// and applied in ascending version order. Each file runs in its own
// transaction together with the bookkeeping insert into
// schema_migrations, so files must not contain BEGIN/COMMIT.
func Migrate(ctx context.Context, db Migrator, dir string) ([]string, error) {
	migrations, err := loadMigrations(dir)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    bigint      PRIMARY KEY,
		name       text        NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`)
	if err != nil {
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	var applied []string
	for _, m := range migrations {
		ok, err := apply(ctx, db, m)
		if err != nil {
			return applied, fmt.Errorf("%s: %w", filepath.Base(m.path), err)
		}
		if ok {
			applied = append(applied, filepath.Base(m.path))
		}
	}
	return applied, nil
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
			return nil, fmt.Errorf("%s: name must start with <version>_", filepath.Base(p))
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
func apply(ctx context.Context, db TxBeginner, m migration) (bool, error) {
	sql, err := os.ReadFile(m.path)
	if err != nil {
		return false, err
	}

	tx, err := db.Begin(ctx)
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
