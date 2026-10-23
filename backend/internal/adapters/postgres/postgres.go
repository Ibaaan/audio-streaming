package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	user   string
	pass   string
	host   string
	port   string
	dbname string
}

func ConfigFromEnv() Config {
	return Config{
		user:   getEnv("POSTGRES_USER", "audio"),
		pass:   getEnv("POSTGRES_PASSWORD", "audio"),
		host:   getEnv("POSTGRES_HOST", "localhost"),
		port:   getEnv("POSTGRES_PORT", "5432"),
		dbname: getEnv("POSTGRES_DB", "audio"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, generateConnStringPGSQL(cfg))
	if err != nil {
		return nil, fmt.Errorf("creating postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return pool, nil
}

func generateConnStringPGSQL(cfg Config) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.user, cfg.pass),
		Host:     cfg.host + ":" + cfg.port,
		Path:     cfg.dbname,
		RawQuery: url.Values{"sslmode": {"disable"}}.Encode(),
	}
	return u.String()
}

// DBTX is a pool, a connection or a transaction. Begin on a transaction
// opens a savepoint, so transactions nest.
type DBTX interface {
	TxBeginner
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TxBeginner is anything that can start a transaction, e.g. *pgxpool.Pool.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

var (
	_ DBTX       = (*pgxpool.Pool)(nil)
	_ DBTX       = (*pgx.Conn)(nil)
	_ DBTX       = pgx.Tx(nil)
	_ TxBeginner = (*pgxpool.Pool)(nil)
)

// WithTransaction runs functions in order inside one transaction. If any fn
// returns an error, the remaining fns are skipped and the transaction is
// rolled back. If all succeed, the transaction is committed.
func WithTransaction(ctx context.Context, txBeginner TxBeginner, fun func(tx DBTX) error) error {
	tx, err := txBeginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rbCtx)
	}()

	if err := fun(tx); err != nil {
		return fmt.Errorf("in tx: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
