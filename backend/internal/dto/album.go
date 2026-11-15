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
