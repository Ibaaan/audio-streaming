package service

import (
	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

// Summary conversion performs no database calls.
func trackSummary(track catalog.TrackOverview) dto.TrackSummary {
	item := dto.TrackSummary{
		TrackRef: dto.TrackRef{ID: track.ID, Title: track.Title, DurationMs: int(track.DurationMs)},
		Artists:  make([]dto.TrackArtist, 0, len(track.Artists)),
	}
	for _, artist := range track.Artists {
		role := dto.ArtistRoleMain
		if artist.Role != nil {
			role = dto.ArtistRole(*artist.Role)
		}
		item.Artists = append(item.Artists, dto.TrackArtist{
			ArtistRef: dto.ArtistRef{ID: artist.ID, DisplayName: artist.Name}, Role: role,
		})
	}
	if track.Album != nil {
		album := track.Album
		ref := albumReference(album.Album)
		item.Album = &ref
		disc, number := int(album.DiscNo), int(album.TrackNo)
		item.DiscNumber, item.TrackNumber = &disc, &number
	}
	return item
}

func albumReference(album catalog.Album) dto.AlbumRef {
	ref := dto.AlbumRef{ID: album.ID, Title: album.Title, AlbumType: dto.AlbumType(album.AlbumType)}
	if album.ReleaseYear != nil {
		year := int(*album.ReleaseYear)
		ref.ReleaseYear = &year
	}
	return ref
}

func albumSummary(album catalog.AlbumOverview) dto.AlbumSummary {
	item := dto.AlbumSummary{
		AlbumRef:   albumReference(album.Album),
		Artists:    make([]dto.ArtistRef, 0, len(album.Artists)),
		TrackCount: album.TrackCount,
	}
	for _, artist := range album.Artists {
		item.Artists = append(item.Artists, dto.ArtistRef{ID: artist.ID, DisplayName: artist.Name})
	}
	return item
}
