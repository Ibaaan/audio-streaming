package main

import (
	"audio-streaming/internal/adapters/postgres"
	"context"
)

func main() {
	postgres.Open(context.Background(), postgres.ConfigFromEnv())
}
