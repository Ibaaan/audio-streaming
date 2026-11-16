package catalog

// ArtistStats counts direct album credits and track credits in any role.
type ArtistStats struct {
	AlbumCount int `db:"album_count"`
	TrackCount int `db:"track_count"`
}

// AlbumOverview contains the metadata needed for an album list entry.
type AlbumOverview struct {
	Album
	Artists    []Artist `db:"artists"`
	TrackCount int      `db:"track_count"`
}
