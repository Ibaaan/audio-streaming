// Package catalog holds the music catalog: tracks, albums, artists and
// their images, as stored by the repository.
package catalog

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ImageFormat string

const (
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatWEBP ImageFormat = "webp"
	ImageFormatGIF  ImageFormat = "gif"
	ImageFormatAVIF ImageFormat = "avif"
)

type AlbumType string

const (
	AlbumTypeAlbum       AlbumType = "album"
	AlbumTypeSingle      AlbumType = "single"
	AlbumTypeEP          AlbumType = "ep"
	AlbumTypeCompilation AlbumType = "compilation"
	AlbumTypeLive        AlbumType = "live"
	AlbumTypeSoundtrack  AlbumType = "soundtrack"
	AlbumTypeOther       AlbumType = "other"
)

type Quality string

const (
	QualityLow      Quality = "low"
	QualityOptimal  Quality = "optimal"
	QualityLossless Quality = "lossless"
)

type Image struct {
	ID             uuid.UUID   `db:"id"              json:"id"`
	StoragePath    string      `db:"storage_path"    json:"storage_path"`
	OriginalFormat ImageFormat `db:"original_format" json:"original_format"`
	Width          *int32      `db:"width"           json:"width,omitempty"`
	Height         *int32      `db:"height"          json:"height,omitempty"`
	CreatedAt      time.Time   `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at"      json:"updated_at"`
}

type Artist struct {
	ID        uuid.UUID  `db:"id"         json:"id"`
	Name      string     `db:"name"       json:"name"`
	SortName  *string    `db:"sort_name"  json:"sort_name,omitempty"`
	ImageID   *uuid.UUID `db:"image_id"   json:"image_id,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

type Album struct {
	ID           uuid.UUID  `db:"id"             json:"id"`
	Title        string     `db:"title"          json:"title"`
	AlbumType    AlbumType  `db:"album_type"     json:"album_type"`
	ReleaseYear  *int16     `db:"release_year"   json:"release_year,omitempty"`
	CoverImageID *uuid.UUID `db:"cover_image_id" json:"cover_image_id,omitempty"`
	CreatedAt    time.Time  `db:"created_at"     json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"     json:"updated_at"`
}

type AlbumArtist struct {
	AlbumID  uuid.UUID `db:"album_id"  json:"album_id"`
	ArtistID uuid.UUID `db:"artist_id" json:"artist_id"`
	Position int16     `db:"position"  json:"position"`
}

type Track struct {
	ID         uuid.UUID `db:"id"          json:"id"`
	Title      string    `db:"title"       json:"title"`
	DurationMs int32     `db:"duration_ms" json:"duration_ms"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"  json:"updated_at"`
}

type TrackArtist struct {
	TrackID  uuid.UUID `db:"track_id"  json:"track_id"`
	ArtistID uuid.UUID `db:"artist_id" json:"artist_id"`
	Role     *string   `db:"role"      json:"role,omitempty"`
	Position int16     `db:"position"  json:"position"`
}

type AlbumTrack struct {
	AlbumID uuid.UUID `db:"album_id" json:"album_id"`
	TrackID uuid.UUID `db:"track_id" json:"track_id"`
	Quality Quality   `db:"quality"  json:"quality"`
	DiscNo  int16     `db:"disc_no"  json:"disc_no"`
	TrackNo int16     `db:"track_no" json:"track_no"`
}

// ErrNotFound is returned by lookups that match no row.
var ErrNotFound = errors.New("catalog: not found")

// TrackSummary is a track joined with what a track list shows next to
// it. Artist joins the track's artists in position order. Album and
// ReleaseYear come from one album the track is on, and are nil if it is
// on none.
type TrackSummary struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	Title       string    `db:"title"        json:"title"`
	Artist      string    `db:"artist"       json:"artist"`
	Album       *string   `db:"album"        json:"album,omitempty"`
	ReleaseYear *int16    `db:"release_year" json:"release_year,omitempty"`
	DurationMs  int32     `db:"duration_ms"  json:"duration_ms"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}
