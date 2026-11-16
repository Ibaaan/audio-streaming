package catalog

// AlbumDetails contains the ordered credits and tracklist for a detail response.
type AlbumDetails struct {
	Album
	Artists []Artist          `db:"artists"`
	Tracks  []AlbumTrackEntry `db:"tracks"`
}

type AlbumTrackEntry struct {
	Track
	DiscNo  int16               `json:"disc_no"`
	TrackNo int16               `json:"track_no"`
	Artists []TrackArtistCredit `json:"artists"`
}
