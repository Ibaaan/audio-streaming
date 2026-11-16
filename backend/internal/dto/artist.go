package dto

import (
	"time"

	"github.com/google/uuid"
)

type ArtistRef struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
}

// TrackArtist adds the artist's credit to the resource reference.
// The order of artists in a track response is the credited order.
type TrackArtist struct {
	ArtistRef
	Role ArtistRole `json:"role"`
}

type ArtistRole string

const (
	ArtistRoleMain     ArtistRole = "main"
	ArtistRoleFeatured ArtistRole = "featured"
	ArtistRoleRemixer  ArtistRole = "remixer"
	ArtistRoleProducer ArtistRole = "producer"
	ArtistRoleComposer ArtistRole = "composer"
)

// Artist is the detail response, with collections included only on request.
type Artist struct {
	ArtistRef
	Image      *Image                    `json:"image"`
	Stats      ArtistStats               `json:"stats"`
	Albums     *ListAlbumsResponse       `json:"albums,omitempty"`
	Tracks     *ListArtistTracksResponse `json:"tracks,omitempty"`
	CreateTime time.Time                 `json:"create_time"`
	UpdateTime time.Time                 `json:"update_time"`
}

type ArtistStats struct {
	AlbumCount int `json:"album_count"`
	TrackCount int `json:"track_count"`
}
