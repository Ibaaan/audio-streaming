package dto

import "github.com/google/uuid"

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
