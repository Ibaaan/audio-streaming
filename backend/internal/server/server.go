package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"audio-streaming/internal/adapters/ffmpeg"
	"audio-streaming/internal/adapters/postgres"
	"audio-streaming/internal/adapters/s3store"
	"audio-streaming/internal/pipeline"
	"audio-streaming/internal/service"
)

func Run(ctx context.Context) error {
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
		Encoder:  NoEncoder{},
		Verifier: NoVerifier{},
		Blobs:    store,
		Records: RecordSaver{
			Lib:   service.NewLibrary(repo),
			Blobs: store,
		},
	}, tmpDir)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.Dir("../static")))
	mux.HandleFunc("GET /v1/tracks", ListTracks(repo))
	mux.HandleFunc("POST /v1/tracks", UploadTrack(p, tmpDir))
	mux.HandleFunc("GET /v1/tracks/{track}/stream",
		StreamTrack(repo, store))

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
	return server.Shutdown(shutdownCtx)
}
