package dto

// TrackSummary is a track in a list with its primary album and position.
type TrackSummary struct {
	TrackRef
	Artists     []TrackArtist `json:"artists"`
	Album       *AlbumRef     `json:"album"`
	DiscNumber  *int          `json:"disc_number"`
	TrackNumber *int          `json:"track_number"`
}

type ArtistTrackItem struct {
	TrackSummary
	ArtistRole ArtistRole `json:"artist_role"`
}

type ListArtistTracksResponse struct {
	Tracks     []ArtistTrackItem `json:"tracks"`
	NextOffset *int              `json:"next_offset,omitempty"`
	TotalSize  int               `json:"total_size"`
}
