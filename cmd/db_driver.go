package main

import (
	repo "audio-streaming/internal/adapters/repository"
	"context"
)

func main() {
	repo.Open(context.Background(), repo.ConfigFromEnv())
}
