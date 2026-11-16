package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"audio-streaming/internal/dto"
)

// ListArtistTracks constructs a page of track DTOs for an existing artist.
func (l *Library) ListArtistTracks(
	ctx context.Context, artistID uuid.UUID, pageSize, offset int,
) (dto.ListArtistTracksResponse, error) {
	if pageSize < 0 || offset < 0 {
		return dto.ListArtistTracksResponse{}, fmt.Errorf("pagination values must not be negative")
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if _, err := l.repo.GetArtist(ctx, artistID); err != nil {
		return dto.ListArtistTracksResponse{}, fmt.Errorf("getting artist: %w", err)
	}
	tracks, total, err := l.repo.ListArtistTracks(ctx, artistID, pageSize, offset)
	if err != nil {
		return dto.ListArtistTracksResponse{}, fmt.Errorf("listing artist tracks: %w", err)
	}
	result := dto.ListArtistTracksResponse{
		Tracks: make([]dto.ArtistTrackItem, 0, len(tracks)), TotalSize: total,
	}
	for _, row := range tracks {
		item := dto.ArtistTrackItem{
			TrackSummary: trackSummary(row.TrackOverview), ArtistRole: dto.ArtistRoleMain,
		}
		if row.Role != nil {
			item.ArtistRole = dto.ArtistRole(*row.Role)
		}

		result.Tracks = append(result.Tracks, item)
	}
	if len(tracks) > 0 && offset < total && len(tracks) < total-offset {
		next := offset + len(tracks)
		result.NextOffset = &next
	}
	return result, nil
}
