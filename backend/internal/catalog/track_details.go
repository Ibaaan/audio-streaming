package catalog

// TrackArtistCredit combines an artist with their credit on a specific track.
type TrackArtistCredit struct {
	Artist
	Role *string `db:"role"`
}

// TrackAlbumAppearance combines an album with the track's position in it.
type TrackAlbumAppearance struct {
	Album
	DiscNo  int16 `db:"disc_no"`
	TrackNo int16 `db:"track_no"`
}

// ArtistTrack is a track credited to the queried artist, including their role.
type ArtistTrack struct {
	TrackOverview
	Role *string `db:"role"`
}

// TrackOverview contains exactly the related data needed for list responses.
type TrackOverview struct {
	Track
	Artists []TrackArtistCredit   `db:"artists"`
	Album   *TrackAlbumAppearance `db:"album"`
}
