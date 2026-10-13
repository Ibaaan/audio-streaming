package repository

import (
	"context"
	_ "embed"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	user   string
	pass   string
	host   string
	port   string
	dbname string
}

func ConfigFromEnv() config {
	return config{
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

func Open(ctx context.Context, cfg config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, generateConnStringPGSQL(cfg))
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func generateConnStringPGSQL(cfg config) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.user, cfg.pass),
		Host:     cfg.host + ":" + cfg.port,
		Path:     cfg.dbname,
		RawQuery: url.Values{"sslmode": {"disable"}}.Encode(),
	}
	return u.String()
}
