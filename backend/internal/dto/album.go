package dto

import (
	"time"

	"github.com/google/uuid"
)

// AlbumRef is album metadata embedded in track responses.
// Unknown release years and missing artwork encode as null.
type AlbumRef struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	AlbumType   AlbumType `json:"album_type"`
	ReleaseYear *int      `json:"release_year"`
	Cover       *Image    `json:"cover"`
}

type AlbumType string

const (
	AlbumTypeAlbum       AlbumType = "album"
	AlbumTypeSingle      AlbumType = "single"
	AlbumTypeEP          AlbumType = "ep"
	AlbumTypeCompilation AlbumType = "compilation"
)

type AlbumSummary struct {
	AlbumRef
	Artists    []ArtistRef `json:"artists"`
	TrackCount int         `json:"track_count"`
}

type ListAlbumsResponse struct {
	Albums     []AlbumSummary `json:"albums"`
	NextOffset *int           `json:"next_offset,omitempty"`
	TotalSize  int            `json:"total_size"`
}

// Album is the detail response. Tracks is omitted unless include=tracks;
// a pointer to an empty slice encodes an included but empty tracklist as [].
type Album struct {
	AlbumRef
	Artists    []ArtistRef   `json:"artists"`
	TrackCount int           `json:"track_count"`
	DiscCount  int           `json:"disc_count"`
	DurationMs int64         `json:"duration_ms"`
	Tracks     *[]AlbumTrack `json:"tracks,omitempty"`
	CreateTime time.Time     `json:"create_time"`
	UpdateTime time.Time     `json:"update_time"`
}

// AlbumTrack does not repeat the containing album.
type AlbumTrack struct {
	TrackRef
	Artists     []TrackArtist `json:"artists"`
	DiscNumber  int           `json:"disc_number"`
	TrackNumber int           `json:"track_number"`
}
