package repository

import (
	"time"
	"uuid"
)

type trackRow struct {
	ID              uuid.UUID
	Title           string
	StoredFormat    string
	Quality         string
	DurationMs      int32
	PipelineVersion int16
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
