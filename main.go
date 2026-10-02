package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	s3store "audio-streaming/internal/adapters/s3store"
)

func main() {
	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	testVolume, err := s3store.CreateTestVolume(ctx,
		s3store.ConfigFromEnv(), "test_data")
	if err != nil {
		slog.Error("creating test volume", "error", err)
		os.Exit(1)
	}
	defer stop()

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.Dir("static")))
	mux.HandleFunc("GET /tracks", listTracks(testVolume.Client))
	mux.HandleFunc("GET /song", getUrl(testVolume.Client,
		"Slipknot - Duality.flac"))

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		slog.Info("listening", "addr", addr)
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	err = testVolume.Delete(context.Background())
	if err != nil {
		slog.Error("deleting test volume", "error", err)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "error", err)
	}
}

type Lister interface {
	List(
		ctx context.Context, prefix string,
	) ([]string, error)
}

func listTracks(client Lister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := client.List(r.Context(), r.URL.Query().Get("prefix"))
		if err != nil {
			slog.Error("listing tracks", "error", err)
			http.Error(w, "failed to list tracks",
				http.StatusInternalServerError)
			return
		}
		if keys == nil {
			keys = []string{}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(keys); err != nil {
			slog.Error("encoding tracks", "error", err)
		}
	}
}

type URLer interface {
	PresignedGetURL(ctx context.Context, key string,
		expiry time.Duration) (string, error)
}

func getUrl(client URLer, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		url, err := client.PresignedGetURL(r.Context(), key, time.Minute)
		if err != nil {
			slog.Error("getting presigned URL", "error", err)
			http.Error(w, "failed to get presigned URL",
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(url))
	}
}
