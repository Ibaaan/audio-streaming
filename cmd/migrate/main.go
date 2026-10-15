package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"audio-streaming/internal/adapters/postgres"

	"github.com/jackc/pgx/v5"
)

// Matches postgres.env used by docker-compose.yml.
const defaultDatabaseURL = "postgres://audio:audio@localhost:5432/audio" +
	"?sslmode=disable"

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

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = conn.Close(context.Background()) }()

	applied, err := postgres.Migrate(ctx, conn, dir)
	for _, name := range applied {
		fmt.Println("applied", name)
	}
	if err != nil {
		return err
	}

	fmt.Printf("%d migration(s) applied, database up to date\n",
		len(applied))

	return nil
}
