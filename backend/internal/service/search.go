package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"audio-streaming/internal/dto"
)

var ErrInvalidSearch = errors.New("invalid search parameters")

// Search builds only requested result groups, each with independent pagination.
func (l *Library) Search(ctx context.Context, req dto.SearchRequest) (dto.SearchResponse, error) {
	query := strings.TrimSpace(req.Query)
	if !utf8.ValidString(query) || utf8.RuneCountInString(query) < 1 ||
		utf8.RuneCountInString(query) > 200 || req.PageSize < 0 || req.Offset < 0 {
		return dto.SearchResponse{}, fmt.Errorf(
			"%w: query must have 1–200 characters; pagination must be non-negative", ErrInvalidSearch,
		)
	}
	types := req.Types
	if len(types) == 0 {
		types = []dto.SearchType{dto.SearchTypeArtist, dto.SearchTypeAlbum, dto.SearchTypeTrack}
	}
	seen := map[dto.SearchType]bool{}
	for _, kind := range types {
		valid := kind == dto.SearchTypeArtist || kind == dto.SearchTypeAlbum ||
			kind == dto.SearchTypeTrack
		if !valid || seen[kind] {
			return dto.SearchResponse{}, fmt.Errorf(
				"%w: types must contain unique artist, album or track values", ErrInvalidSearch,
			)
		}
		seen[kind] = true
	}
	size := req.PageSize
	if size == 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	result := dto.SearchResponse{}
	for _, kind := range types {
		switch kind {
		case dto.SearchTypeArtist:
			artists, total, err := l.repo.SearchArtists(ctx, query, size, req.Offset)
			if err != nil {
				return dto.SearchResponse{}, fmt.Errorf("searching artists: %w", err)
			}
			page := dto.ListArtistsResponse{
				Artists: make([]dto.ArtistSummary, 0, len(artists)), TotalSize: total,
				NextOffset: searchNextOffset(req.Offset, len(artists), total),
			}
			for _, artist := range artists {
				page.Artists = append(page.Artists, dto.ArtistSummary{
					ArtistRef: dto.ArtistRef{ID: artist.ID, DisplayName: artist.Name},
				})
			}
			result.Artists = &page
		case dto.SearchTypeAlbum:
			albums, total, err := l.repo.SearchAlbums(ctx, query, size, req.Offset)
			if err != nil {
				return dto.SearchResponse{}, fmt.Errorf("searching albums: %w", err)
			}
			page := dto.ListAlbumsResponse{
				Albums: make([]dto.AlbumSummary, 0, len(albums)), TotalSize: total,
				NextOffset: searchNextOffset(req.Offset, len(albums), total),
			}
			for _, album := range albums {
				page.Albums = append(page.Albums, albumSummary(album))
			}

			result.Albums = &page
		case dto.SearchTypeTrack:
			tracks, total, err := l.repo.SearchTracks(ctx, query, size, req.Offset)
			if err != nil {
				return dto.SearchResponse{}, fmt.Errorf("searching tracks: %w", err)
			}
			page := dto.ListTracksResponse{
				Tracks: make([]dto.TrackSummary, 0, len(tracks)), TotalSize: total,
				NextOffset: searchNextOffset(req.Offset, len(tracks), total),
			}
			for _, track := range tracks {
				page.Tracks = append(page.Tracks, trackSummary(track))
			}

			result.Tracks = &page
		}
	}
	return result, nil
}

func searchNextOffset(offset, count, total int) *int {
	if count == 0 || offset >= total || count >= total-offset {
		return nil
	}
	next := offset + count
	return &next
}
