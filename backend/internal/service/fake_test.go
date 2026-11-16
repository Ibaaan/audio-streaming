package service

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
)

// fakeRepo is an in-memory catalog.Repository. WithinTx rolls back by
// restoring a copy of the slices taken before fn ran.
type fakeRepo struct {
	tracks       []catalog.Track
	artists      []catalog.Artist
	albums       []catalog.Album
	images       []catalog.Image
	trackArtists []catalog.TrackArtist
	albumArtists []catalog.AlbumArtist
	albumTracks  []catalog.AlbumTrack
}

var _ catalog.Repository = (*fakeRepo)(nil)

func (f *fakeRepo) WithinTx(ctx context.Context, fn func(tx catalog.Repository) error) error {
	saved := *f
	if err := fn(f); err != nil {
		*f = saved
		return err
	}
	return nil
}

func first[T any](items []T, match func(T) bool) (T, error) {
	for _, v := range items {
		if match(v) {
			return v, nil
		}
	}
	var zero T
	return zero, catalog.ErrNotFound
}

func (f *fakeRepo) AddTrack(ctx context.Context, t *catalog.Track) error {
	f.tracks = append(f.tracks, *t)
	return nil
}

func (f *fakeRepo) GetTrack(ctx context.Context, id uuid.UUID) (catalog.Track, error) {
	return first(f.tracks, func(t catalog.Track) bool { return t.ID == id })
}

func (f *fakeRepo) ListTracks(ctx context.Context) ([]catalog.TrackSummary, error) {
	tracks := make([]catalog.TrackSummary, 0, len(f.tracks))
	for _, track := range f.tracks {
		tracks = append(tracks, catalog.TrackSummary{
			Track: track,
		})
	}
	return tracks, nil
}

func (f *fakeRepo) FindTrack(
	ctx context.Context,
	title string,
	artistID uuid.UUID,
) (catalog.Track, error) {
	return first(f.tracks, func(t catalog.Track) bool {
		_, err := first(f.trackArtists, func(ta catalog.TrackArtist) bool {
			return ta.TrackID == t.ID && ta.ArtistID == artistID
		})
		return err == nil && strings.EqualFold(t.Title, title)
	})
}

func (f *fakeRepo) AddTrackArtist(ctx context.Context, ta catalog.TrackArtist) error {
	f.trackArtists = append(f.trackArtists, ta)
	return nil
}

func (f *fakeRepo) AddArtist(ctx context.Context, a *catalog.Artist) error {
	f.artists = append(f.artists, *a)
	return nil
}

func (f *fakeRepo) GetArtist(ctx context.Context, id uuid.UUID) (catalog.Artist, error) {
	return first(f.artists, func(a catalog.Artist) bool { return a.ID == id })
}

func (f *fakeRepo) FindArtistByName(ctx context.Context, name string) (catalog.Artist, error) {
	return first(f.artists, func(a catalog.Artist) bool {
		return strings.EqualFold(a.Name, name)
	})
}

func (f *fakeRepo) AddAlbum(ctx context.Context, a *catalog.Album) error {
	f.albums = append(f.albums, *a)
	return nil
}

func (f *fakeRepo) GetAlbum(ctx context.Context, id uuid.UUID) (catalog.Album, error) {
	return first(f.albums, func(a catalog.Album) bool { return a.ID == id })
}

func (f *fakeRepo) FindAlbum(
	ctx context.Context,
	title string,
	artistID uuid.UUID,
) (catalog.Album, error) {
	return first(f.albums, func(a catalog.Album) bool {
		_, err := first(f.albumArtists, func(aa catalog.AlbumArtist) bool {
			return aa.AlbumID == a.ID && aa.ArtistID == artistID
		})
		return err == nil && strings.EqualFold(a.Title, title)
	})
}

func (f *fakeRepo) AddAlbumArtist(ctx context.Context, aa catalog.AlbumArtist) error {
	f.albumArtists = append(f.albumArtists, aa)
	return nil
}

func (f *fakeRepo) AddAlbumTrack(ctx context.Context, at catalog.AlbumTrack) error {
	f.albumTracks = append(f.albumTracks, at)
	return nil
}

func (f *fakeRepo) AddImage(ctx context.Context, img *catalog.Image) error {
	f.images = append(f.images, *img)
	return nil
}

func (f *fakeRepo) GetImage(ctx context.Context, id uuid.UUID) (catalog.Image, error) {
	return first(f.images, func(i catalog.Image) bool { return i.ID == id })
}

func (f *fakeRepo) FindImageByPath(ctx context.Context, storagePath string) (catalog.Image, error) {
	return first(f.images, func(i catalog.Image) bool {
		return i.StoragePath == storagePath
	})
}

func (f *fakeRepo) GetArtistByTrackId(
	ctx context.Context, id uuid.UUID,
) ([]catalog.TrackArtistCredit, error) {
	credits := append([]catalog.TrackArtist{}, f.trackArtists...)
	sort.Slice(credits, func(i, j int) bool { return credits[i].Position < credits[j].Position })
	result := []catalog.TrackArtistCredit{}
	for _, credit := range credits {
		if credit.TrackID != id {
			continue
		}
		artist, err := f.GetArtist(ctx, credit.ArtistID)
		if err != nil {
			return nil, err
		}
		result = append(result, catalog.TrackArtistCredit{Artist: artist, Role: credit.Role})
	}
	return result, nil
}

func (f *fakeRepo) GetAlbumByTrackId(
	ctx context.Context, id uuid.UUID,
) ([]catalog.TrackAlbumAppearance, error) {
	result := []catalog.TrackAlbumAppearance{}
	for _, link := range f.albumTracks {
		if link.TrackID != id {
			continue
		}
		album, err := f.GetAlbum(ctx, link.AlbumID)
		if err != nil {
			return nil, err
		}
		result = append(result, catalog.TrackAlbumAppearance{
			Album: album, DiscNo: link.DiscNo, TrackNo: link.TrackNo,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].ID.String() < result[j].ID.String()
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (f *fakeRepo) ListArtistTracks(
	ctx context.Context, id uuid.UUID, limit, offset int,
) ([]catalog.ArtistTrack, int, error) {
	result := []catalog.ArtistTrack{}
	for _, credit := range f.trackArtists {
		if credit.ArtistID != id {
			continue
		}
		track, err := f.GetTrack(ctx, credit.TrackID)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, catalog.ArtistTrack{Track: track, Role: credit.Role})
	}
	sort.Slice(result, func(i, j int) bool {
		a, _ := f.GetAlbumByTrackId(ctx, result[i].ID)
		b, _ := f.GetAlbumByTrackId(ctx, result[j].ID)
		var ay, by, ad, bd, at, bt int16
		if len(a) > 0 {
			ad, at = a[0].DiscNo, a[0].TrackNo
			if a[0].ReleaseYear != nil {
				ay = *a[0].ReleaseYear
			}
		}
		if len(b) > 0 {
			bd, bt = b[0].DiscNo, b[0].TrackNo
			if b[0].ReleaseYear != nil {
				by = *b[0].ReleaseYear
			}
		}
		if ay != by {
			return ay > by
		}
		if (ad == 0) != (bd == 0) {
			return ad != 0
		}
		if ad != bd {
			return ad < bd
		}
		if at != bt {
			return at < bt
		}
		return result[i].ID.String() < result[j].ID.String()
	})
	total := len(result)
	start := min(offset, total)
	end := start + min(limit, total-start)
	return result[start:end], total, nil
}

func (f *fakeRepo) GetArtistStats(ctx context.Context, id uuid.UUID) (catalog.ArtistStats, error) {
	if _, err := f.GetArtist(ctx, id); err != nil {
		return catalog.ArtistStats{}, err
	}
	stats := catalog.ArtistStats{}
	for _, credit := range f.albumArtists {
		if credit.ArtistID == id {
			stats.AlbumCount++
		}
	}
	for _, credit := range f.trackArtists {
		if credit.ArtistID == id {
			stats.TrackCount++
		}
	}
	return stats, nil
}

func (f *fakeRepo) ListArtistAlbums(
	ctx context.Context, id uuid.UUID, limit, offset int,
) ([]catalog.AlbumOverview, int, error) {
	result := []catalog.AlbumOverview{}
	for _, credit := range f.albumArtists {
		if credit.ArtistID != id {
			continue
		}
		album, err := f.GetAlbum(ctx, credit.AlbumID)
		if err != nil {
			return nil, 0, err
		}
		item := catalog.AlbumOverview{Album: album, Artists: []catalog.Artist{}}
		credits := append([]catalog.AlbumArtist{}, f.albumArtists...)
		sort.Slice(credits, func(i, j int) bool { return credits[i].Position < credits[j].Position })
		for _, link := range credits {
			if link.AlbumID != album.ID {
				continue
			}
			artist, err := f.GetArtist(ctx, link.ArtistID)
			if err != nil {
				return nil, 0, err
			}
			item.Artists = append(item.Artists, artist)
		}
		for _, link := range f.albumTracks {
			if link.AlbumID == album.ID {
				item.TrackCount++
			}
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.ReleaseYear == nil && b.ReleaseYear != nil {
			return false
		}
		if a.ReleaseYear != nil && b.ReleaseYear == nil {
			return true
		}
		if a.ReleaseYear != nil && b.ReleaseYear != nil && *a.ReleaseYear != *b.ReleaseYear {
			return *a.ReleaseYear > *b.ReleaseYear
		}
		return a.ID.String() < b.ID.String()
	})
	total := len(result)
	start := min(offset, total)
	end := start + min(limit, total-start)
	return result[start:end], total, nil
}

func (f *fakeRepo) GetAlbumDetails(ctx context.Context, id uuid.UUID) (catalog.AlbumDetails, error) {
	album, err := f.GetAlbum(ctx, id)
	if err != nil {
		return catalog.AlbumDetails{}, err
	}
	result := catalog.AlbumDetails{Album: album, Artists: []catalog.Artist{}, Tracks: []catalog.AlbumTrackEntry{}}
	credits := append([]catalog.AlbumArtist{}, f.albumArtists...)
	sort.Slice(credits, func(i, j int) bool { return credits[i].Position < credits[j].Position })
	for _, credit := range credits {
		if credit.AlbumID != id {
			continue
		}
		artist, err := f.GetArtist(ctx, credit.ArtistID)
		if err != nil {
			return catalog.AlbumDetails{}, err
		}
		result.Artists = append(result.Artists, artist)
	}
	for _, link := range f.albumTracks {
		if link.AlbumID != id {
			continue
		}
		track, err := f.GetTrack(ctx, link.TrackID)
		if err != nil {
			return catalog.AlbumDetails{}, err
		}
		artists, err := f.GetArtistByTrackId(ctx, track.ID)
		if err != nil {
			return catalog.AlbumDetails{}, err
		}
		result.Tracks = append(result.Tracks, catalog.AlbumTrackEntry{
			Track: track, DiscNo: link.DiscNo, TrackNo: link.TrackNo, Artists: artists,
		})
	}
	sort.Slice(result.Tracks, func(i, j int) bool {
		a, b := result.Tracks[i], result.Tracks[j]
		if a.DiscNo != b.DiscNo {
			return a.DiscNo < b.DiscNo
		}
		return a.TrackNo < b.TrackNo
	})
	return result, nil
}
