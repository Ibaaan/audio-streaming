package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"audio-streaming/internal/dto"
)

// GetAlbum constructs album details with an optional complete ordered tracklist.
func (l *Library) GetAlbum(
	ctx context.Context, id uuid.UUID, include []string,
) (dto.Album, error) {
	if len(include) > 1 || (len(include) == 1 && include[0] != "tracks") {
		return dto.Album{}, fmt.Errorf("include must contain only tracks once")
	}
	album, err := l.repo.GetAlbumDetails(ctx, id)
	if err != nil {
		return dto.Album{}, fmt.Errorf("getting album: %w", err)
	}
	result := dto.Album{
		AlbumRef: dto.AlbumRef{
			ID: album.ID, Title: album.Title, AlbumType: dto.AlbumType(album.AlbumType),
		},
		Artists: make([]dto.ArtistRef, 0, len(album.Artists)), TrackCount: len(album.Tracks),
		CreateTime: album.CreatedAt.UTC(), UpdateTime: album.UpdatedAt.UTC(),
		// Artwork stays null until storage URL resolution is available.
	}
	if album.ReleaseYear != nil {
		year := int(*album.ReleaseYear)
		result.ReleaseYear = &year
	}
	for _, artist := range album.Artists {
		result.Artists = append(result.Artists, dto.ArtistRef{
			ID: artist.ID, DisplayName: artist.Name,
		})
	}
	tracks := make([]dto.AlbumTrack, 0, len(album.Tracks))
	discs := map[int16]bool{}
	for _, track := range album.Tracks {
		discs[track.DiscNo] = true
		result.DurationMs += int64(track.DurationMs)
		if len(include) == 0 {
			continue
		}
		item := dto.AlbumTrack{
			TrackRef: dto.TrackRef{
				ID: track.ID, Title: track.Title, DurationMs: int(track.DurationMs),
			},
			Artists:    make([]dto.TrackArtist, 0, len(track.Artists)),
			DiscNumber: int(track.DiscNo), TrackNumber: int(track.TrackNo),
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
		tracks = append(tracks, item)
	}
	result.DiscCount = len(discs)
	if len(include) > 0 {
		result.Tracks = &tracks
	}
	return result, nil
}
