package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"audio-streaming/internal/adapters/ffmpeg"
	"audio-streaming/internal/adapters/postgres"
	s3store "audio-streaming/internal/adapters/s3store"
	"audio-streaming/internal/pipeline"
	"audio-streaming/internal/server"
	"audio-streaming/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	pool, err := postgres.Open(ctx, postgres.ConfigFromEnv())
	if err != nil {
		return err
	}
	defer pool.Close()

	applied, err := postgres.Migrate(ctx, pool, "migrations")
	if err != nil {
		return err
	}
	for _, name := range applied {
		slog.Info("applied migration", "name", name)
	}

	store, err := s3store.New(ctx, s3store.ConfigFromEnv())
	if err != nil {
		return err
	}

	repo := postgres.NewRepository(pool)
	ff := ffmpeg.New(os.Getenv("FFMPEG"), os.Getenv("FFPROBE"))
	tmpDir := os.TempDir()
	p, err := pipeline.New(pipeline.Deps{
		Decoder:  ff,
		Meta:     ff,
		Encoder:  server.NoEncoder{},
		Verifier: server.NoVerifier{},
		Blobs:    store,
		Records: server.RecordSaver{
			Lib:   service.NewLibrary(repo),
			Blobs: store,
		},
	}, tmpDir)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.Dir("../static")))
	mux.HandleFunc("GET /v1/tracks", server.ListTracks(repo))
	mux.HandleFunc("POST /v1/tracks", server.UploadTrack(p, tmpDir))
	mux.HandleFunc("GET /v1/tracks/{track}/stream",
		server.StreamTrack(repo, store))

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	server := &http.Server{Addr: addr, Handler: mux}

	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", addr)
		errc <- server.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(shutdownCtx)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
