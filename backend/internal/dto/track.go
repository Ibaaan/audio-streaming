package dto

import (
	"time"

	"github.com/google/uuid"
)

// TrackRef contains the identifying and display fields shared by track responses.
type TrackRef struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	DurationMs int       `json:"duration_ms"`
	Quality    Quality   `json:"quality"`
}

// Track is the response body for GET /v1/tracks/{track}.
// Initialize Artists and Albums to empty slices when absent
// so required collections encode as [] rather than null.
type Track struct {
	TrackRef
	Artists    []TrackArtist `json:"artists"`
	Albums     []TrackAlbum  `json:"albums"`
	Audio      TrackAudio    `json:"audio"`
	CreateTime time.Time     `json:"create_time"`
	UpdateTime time.Time     `json:"update_time"`
}

// TrackAlbum includes the track's position in an associated album.
type TrackAlbum struct {
	AlbumRef
	DiscNumber  int `json:"disc_number"`
	TrackNumber int `json:"track_number"`
}

// TrackAudio describes playback normalization metadata.
type TrackAudio struct {
	LoudnessLUFS float64 `json:"loudness_lufs"`
	TruePeakDB   float64 `json:"true_peak_db"`
}

type Quality string

const (
	QualityLossless Quality = "lossless"
	QualityOptimal  Quality = "optimal"
	QualityLow      Quality = "low"
)
