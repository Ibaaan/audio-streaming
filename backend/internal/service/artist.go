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
		page, err := l.ListArtistAlbums(ctx, id, 20, 0)
		if err != nil {
			return dto.Artist{}, err
		}
		result.Albums = &page
	}

	return result, nil
}
