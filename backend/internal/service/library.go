package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"audio-streaming/internal/audio"
	"audio-streaming/internal/catalog"
)

// ErrTrackExists is returned by Save when the catalog already has a
// track with the same title and artist.
var ErrTrackExists = errors.New("service: track already exists")

// Library keeps the catalog's metadata. It never touches file contents:
// audio and cover art are put in blob storage by the caller, which
// passes Library only where they were stored.
type Library struct {
	repo catalog.Repository
}

func NewLibrary(repo catalog.Repository) *Library {
	return &Library{repo: repo}
}

// Save stores rec as a new track with ID rec.TrackId, in one
// transaction. Its artist and album are looked up and created if the
// catalog does not have them yet. rec.Cover is ignored; cover, if not
// nil, describes artwork the caller has already stored, and becomes the
// cover of a newly created album. Only its storage path, format and
// size are used.
func (l *Library) Save(
	ctx context.Context,
	rec audio.Record,
	cover *catalog.Image,
) (catalog.Track, error) {
	e, err := newEntry(rec, cover)
	if err != nil {
		return catalog.Track{}, err
	}

	var track catalog.Track
	err = l.repo.WithinTx(ctx, func(tx catalog.Repository) error {
		var err error
		track, err = l.save(ctx, tx, e)
		return err
	})
	return track, err
}

// save and the methods below it must use tx, never l.repo, so that
// everything they write is part of the transaction.
func (l *Library) save(
	ctx context.Context,
	tx catalog.Repository,
	e entry,
) (catalog.Track, error) {
	artist, err := findOrAddArtist(ctx, tx, e.artist)
	if err != nil {
		return catalog.Track{}, err
	}

	track, err := addNewTrack(ctx, tx, e, artist)
	if err != nil {
		return track, err
	}

	err = tx.AddTrackArtist(ctx, catalog.TrackArtist{
		TrackID:  track.ID,
		ArtistID: artist.ID,
	})
	if err != nil {
		return catalog.Track{}, fmt.Errorf("adding track artist: %w", err)
	}

	if e.album != nil {
		if err := l.addToAlbum(ctx, tx, track, *e.album); err != nil {
			return catalog.Track{}, err
		}
	}
	return track, nil
}

func addNewTrack(
	ctx context.Context,
	tx catalog.Repository,
	e entry,
	artist catalog.Artist,
) (catalog.Track, error) {
	_, err := tx.FindTrack(ctx, e.title, artist.ID)
	if err == nil {
		return catalog.Track{}, fmt.Errorf("%w: %q by %q",
			ErrTrackExists, e.title, e.artist)
	}
	if !errors.Is(err, catalog.ErrNotFound) {
		return catalog.Track{}, fmt.Errorf("finding track: %w", err)
	}

	track := catalog.Track{
		ID:         e.trackID,
		Title:      e.title,
		DurationMs: e.durationMs,
	}
	if err := tx.AddTrack(ctx, &track); err != nil {
		return catalog.Track{}, fmt.Errorf("adding track: %w", err)
	}
	return track, nil
}

func (l *Library) addToAlbum(
	ctx context.Context,
	tx catalog.Repository,
	track catalog.Track,
	a albumEntry,
) error {
	artist, err := findOrAddArtist(ctx, tx, a.artist)
	if err != nil {
		return err
	}

	album, err := tx.FindAlbum(ctx, a.title, artist.ID)
	switch {
	case errors.Is(err, catalog.ErrNotFound):
		album, err = l.addAlbum(ctx, tx, a, artist)
		if err != nil {
			return err
		}
	case err != nil:
		return fmt.Errorf("finding album: %w", err)
	}

	err = tx.AddAlbumTrack(ctx, catalog.AlbumTrack{
		AlbumID: album.ID,
		TrackID: track.ID,
		Quality: a.quality,
		DiscNo:  a.discNo,
		TrackNo: a.trackNo,
	})
	if err != nil {
		return fmt.Errorf("adding album track: %w", err)
	}
	return nil
}

func (l *Library) addAlbum(
	ctx context.Context,
	tx catalog.Repository,
	a albumEntry,
	artist catalog.Artist,
) (catalog.Album, error) {
	coverID, err := findOrAddCover(ctx, tx, a.cover)
	if err != nil {
		return catalog.Album{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return catalog.Album{}, err
	}
	album := catalog.Album{
		ID:           id,
		Title:        a.title,
		AlbumType:    catalog.AlbumTypeAlbum,
		ReleaseYear:  a.releaseYear,
		CoverImageID: coverID,
	}
	if err := tx.AddAlbum(ctx, &album); err != nil {
		return catalog.Album{}, fmt.Errorf("adding album: %w", err)
	}
	err = tx.AddAlbumArtist(ctx, catalog.AlbumArtist{
		AlbumID:  album.ID,
		ArtistID: artist.ID,
	})
	if err != nil {
		return catalog.Album{}, fmt.Errorf("adding album artist: %w", err)
	}
	return album, nil
}

// findOrAddCover returns the ID of the image stored at cover's path,
// adding it if the catalog does not have it yet. A nil cover gives a
// nil ID.
func findOrAddCover(
	ctx context.Context,
	tx catalog.Repository,
	cover *catalog.Image,
) (*uuid.UUID, error) {
	if cover == nil {
		return nil, nil
	}
	img, err := tx.FindImageByPath(ctx, cover.StoragePath)
	if err == nil {
		return &img.ID, nil
	}
	if !errors.Is(err, catalog.ErrNotFound) {
		return nil, fmt.Errorf("finding cover: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	img = catalog.Image{
		ID:             id,
		StoragePath:    cover.StoragePath,
		OriginalFormat: cover.OriginalFormat,
		Width:          cover.Width,
		Height:         cover.Height,
	}
	if err := tx.AddImage(ctx, &img); err != nil {
		return nil, fmt.Errorf("adding cover: %w", err)
	}
	return &img.ID, nil
}

func findOrAddArtist(
	ctx context.Context,
	tx catalog.Repository,
	name string,
) (catalog.Artist, error) {
	artist, err := tx.FindArtistByName(ctx, name)
	if err == nil {
		return artist, nil
	}
	if !errors.Is(err, catalog.ErrNotFound) {
		return catalog.Artist{}, fmt.Errorf("finding artist: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return catalog.Artist{}, err
	}
	artist = catalog.Artist{ID: id, Name: name}
	if err := tx.AddArtist(ctx, &artist); err != nil {
		return catalog.Artist{}, fmt.Errorf("adding artist: %w", err)
	}
	return artist, nil
}
