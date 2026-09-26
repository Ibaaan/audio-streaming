package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"audio-streaming/storage"
)

func main() {
	ctx := context.Background()

	testVolume, err := storage.CreateTestVolume(ctx,
		storage.ConfigFromEnv(), "test_data")
	if err != nil {
		slog.Error("creating test volume", "error", err)
		os.Exit(1)
	}
	defer testVolume.Delete(ctx)

	// client, err := storage.NewClient(ctx, storage.ConfigFromEnv())
	// if err != nil {
	// 	slog.Error("connecting to storage", "error", err)
	// 	os.Exit(1)
	// }

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tracks", listTracks(testVolume.Client))

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server", "error", err)
		os.Exit(1)
	}
}

func listTracks(client storage.StorageClient) http.HandlerFunc {
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
