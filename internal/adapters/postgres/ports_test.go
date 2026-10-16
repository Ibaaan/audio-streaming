package postgres

import "audio-streaming/internal/service"

var _ service.Repository = (*Repository)(nil)
