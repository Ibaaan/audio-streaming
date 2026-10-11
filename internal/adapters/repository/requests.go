package repository

import (
	"context"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
