package dto

import "github.com/google/uuid"

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
