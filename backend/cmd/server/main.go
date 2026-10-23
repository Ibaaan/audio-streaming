package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"audio-streaming/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := server.Run(ctx); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}
