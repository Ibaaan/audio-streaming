package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"audio-streaming/internal/dto"
)

// ListArtistAlbums constructs a page of albums for an existing artist.
func (l *Library) ListArtistAlbums(
	ctx context.Context, id uuid.UUID, pageSize, offset int,
) (dto.ListAlbumsResponse, error) {
	if pageSize < 0 || offset < 0 {
		return dto.ListAlbumsResponse{}, fmt.Errorf("pagination values must not be negative")
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if _, err := l.repo.GetArtist(ctx, id); err != nil {
		return dto.ListAlbumsResponse{}, fmt.Errorf("getting artist: %w", err)
	}
	albums, total, err := l.repo.ListArtistAlbums(ctx, id, pageSize, offset)
	if err != nil {
		return dto.ListAlbumsResponse{}, fmt.Errorf("listing artist albums: %w", err)
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
	if len(albums) > 0 && offset < total && len(albums) < total-offset {
		next := offset + len(albums)
		page.NextOffset = &next
	}
	return page, nil
}
