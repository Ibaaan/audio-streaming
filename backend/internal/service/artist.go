package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"audio-streaming/internal/dto"
)

// GetArtist builds the detail DTO and first pages of requested collections.
func (l *Library) GetArtist(
	ctx context.Context, id uuid.UUID, include []string,
) (dto.Artist, error) {
	seen := map[string]bool{}
	for _, value := range include {
		if (value != "albums" && value != "tracks") || seen[value] {
			return dto.Artist{}, fmt.Errorf("include must contain unique albums or tracks values")
		}
		seen[value] = true
	}
	artist, err := l.repo.GetArtist(ctx, id)
	if err != nil {
		return dto.Artist{}, fmt.Errorf("getting artist: %w", err)
	}
	stats, err := l.repo.GetArtistStats(ctx, id)
	if err != nil {
		return dto.Artist{}, fmt.Errorf("getting artist stats: %w", err)
	}
	result := dto.Artist{
		ArtistRef:  dto.ArtistRef{ID: artist.ID, DisplayName: artist.Name},
		Stats:      dto.ArtistStats{AlbumCount: stats.AlbumCount, TrackCount: stats.TrackCount},
		CreateTime: artist.CreatedAt.UTC(), UpdateTime: artist.UpdatedAt.UTC(),
		// As with track covers, artwork stays null until storage URLs are exposed.
	}
	if seen["tracks"] {
		tracks, err := l.ListArtistTracks(ctx, id, 20, 0)
		if err != nil {
			return dto.Artist{}, err
		}
		result.Tracks = &tracks
	}
	if seen["albums"] {
		albums, total, err := l.repo.ListArtistAlbums(ctx, id, 20, 0)
		if err != nil {
			return dto.Artist{}, fmt.Errorf("listing artist albums: %w", err)
		}
		page := dto.ListAlbumsResponse{
			Albums: make([]dto.AlbumSummary, 0, len(albums)), TotalSize: total,
		}
		for _, album := range albums {
			item := dto.AlbumSummary{
				AlbumRef: dto.AlbumRef{
					ID: album.ID, Title: album.Title, AlbumType: dto.AlbumType(album.AlbumType),
				},
				Artists:    make([]dto.ArtistRef, 0, len(album.Artists)),
				TrackCount: album.TrackCount,
			}
			if album.ReleaseYear != nil {
				year := int(*album.ReleaseYear)
				item.ReleaseYear = &year
			}
			for _, credit := range album.Artists {
				item.Artists = append(item.Artists, dto.ArtistRef{
					ID: credit.ID, DisplayName: credit.Name,
				})
			}
			page.Albums = append(page.Albums, item)
		}
		if len(albums) > 0 && len(albums) < total {
			next := len(albums)
			page.NextOffset = &next
		}
		result.Albums = &page
	}
	return result, nil
}
