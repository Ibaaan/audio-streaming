package dto

type SearchType string

const (
	SearchTypeArtist SearchType = "artist"
	SearchTypeAlbum  SearchType = "album"
	SearchTypeTrack  SearchType = "track"
)

// SearchRequest contains parsed query parameters.
type SearchRequest struct {
	Query    string
	Types    []SearchType
	PageSize int
	Offset   int
}

type ArtistSummary struct {
	ArtistRef
	Image *Image `json:"image"`
}

type ListArtistsResponse struct {
	Artists    []ArtistSummary `json:"artists"`
	NextOffset *int            `json:"next_offset,omitempty"`
	TotalSize  int             `json:"total_size"`
}

type ListTracksResponse struct {
	Tracks     []TrackSummary `json:"tracks"`
	NextOffset *int           `json:"next_offset,omitempty"`
	TotalSize  int            `json:"total_size"`
}

// SearchResponse omits groups that were not requested.
type SearchResponse struct {
	Artists *ListArtistsResponse `json:"artists,omitempty"`
	Albums  *ListAlbumsResponse  `json:"albums,omitempty"`
	Tracks  *ListTracksResponse  `json:"tracks,omitempty"`
}
