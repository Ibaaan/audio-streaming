package postgres

import (
	"context"

	"github.com/google/uuid"
)

var INSERT_TRACK_QUERY string = `
INSERT INTO tracks (id, title, duration_ms)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at`

func AddTrack(
	ctx context.Context,
	dbtx DBTX,
	title string,
	durationMs int32,
) (Track, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Track{}, err
	}

	track := Track{
		ID:         id,
		Title:      title,
		DurationMs: durationMs,
	}

	err = dbtx.QueryRow(
		ctx,
		INSERT_TRACK_QUERY,
		track.ID,
		track.Title,
		track.DurationMs,
	).Scan(&track.CreatedAt, &track.UpdatedAt)
	if err != nil {
		return Track{}, err
	}
	return track, nil
}

var INSERT_IMAGE_QUERY string = `
INSERT INTO images (id, storage_path, original_format, width, height)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at, updated_at`

func AddImage(
	ctx context.Context,
	dbtx DBTX,
	storagePath string,
	format ImageFormat,
	width *int32,
	height *int32,
) (Image, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Image{}, err
	}

	image := Image{
		ID:             id,
		StoragePath:    storagePath,
		OriginalFormat: format,
		Width:          width,
		Height:         height,
	}

	err = dbtx.QueryRow(
		ctx,
		INSERT_IMAGE_QUERY,
		image.ID,
		image.StoragePath,
		image.OriginalFormat,
		image.Width,
		image.Height,
	).Scan(&image.CreatedAt, &image.UpdatedAt)
	if err != nil {
		return Image{}, err
	}
	return image, nil
}

var INSERT_ARTIST_QUERY string = `
INSERT INTO artists (id, name, sort_name, image_id)
VALUES ($1, $2, $3, $4)
RETURNING created_at, updated_at`

func AddArtist(
	ctx context.Context,
	dbtx DBTX,
	name string,
	sortName *string,
	imageID *uuid.UUID,
) (Artist, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Artist{}, err
	}

	artist := Artist{
		ID:       id,
		Name:     name,
		SortName: sortName,
		ImageID:  imageID,
	}

	err = dbtx.QueryRow(
		ctx,
		INSERT_ARTIST_QUERY,
		artist.ID,
		artist.Name,
		artist.SortName,
		artist.ImageID,
	).Scan(&artist.CreatedAt, &artist.UpdatedAt)
	if err != nil {
		return Artist{}, err
	}
	return artist, nil
}

var INSERT_ALBUM_QUERY string = `
INSERT INTO albums (id, title, album_type, release_year, cover_image_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at, updated_at`

func AddAlbum(
	ctx context.Context,
	dbtx DBTX,
	title string,
	albumType AlbumType,
	releaseYear *int16,
	coverImageID *uuid.UUID,
) (Album, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Album{}, err
	}

	album := Album{
		ID:           id,
		Title:        title,
		AlbumType:    albumType,
		ReleaseYear:  releaseYear,
		CoverImageID: coverImageID,
	}

	err = dbtx.QueryRow(
		ctx,
		INSERT_ALBUM_QUERY,
		album.ID,
		album.Title,
		album.AlbumType,
		album.ReleaseYear,
		album.CoverImageID,
	).Scan(&album.CreatedAt, &album.UpdatedAt)
	if err != nil {
		return Album{}, err
	}
	return album, nil
}

var INSERT_ALBUM_ARTIST_QUERY string = `
INSERT INTO album_artists (album_id, artist_id, position)
VALUES ($1, $2, $3)`

func AddAlbumArtist(
	ctx context.Context,
	dbtx DBTX,
	albumID uuid.UUID,
	artistID uuid.UUID,
	position int16,
) (AlbumArtist, error) {
	albumArtist := AlbumArtist{
		AlbumID:  albumID,
		ArtistID: artistID,
		Position: position,
	}

	_, err := dbtx.Exec(
		ctx,
		INSERT_ALBUM_ARTIST_QUERY,
		albumArtist.AlbumID,
		albumArtist.ArtistID,
		albumArtist.Position,
	)
	if err != nil {
		return AlbumArtist{}, err
	}
	return albumArtist, nil
}

var INSERT_TRACK_ARTIST_QUERY string = `
INSERT INTO track_artists (track_id, artist_id, role, position)
VALUES ($1, $2, $3, $4)`

func AddTrackArtist(
	ctx context.Context,
	dbtx DBTX,
	trackID uuid.UUID,
	artistID uuid.UUID,
	role *string,
	position int16,
) (TrackArtist, error) {
	trackArtist := TrackArtist{
		TrackID:  trackID,
		ArtistID: artistID,
		Role:     role,
		Position: position,
	}

	_, err := dbtx.Exec(
		ctx,
		INSERT_TRACK_ARTIST_QUERY,
		trackArtist.TrackID,
		trackArtist.ArtistID,
		trackArtist.Role,
		trackArtist.Position,
	)
	if err != nil {
		return TrackArtist{}, err
	}
	return trackArtist, nil
}

var INSERT_ALBUM_TRACK_QUERY string = `
INSERT INTO album_tracks (album_id, track_id, quality, disc_no, track_no)
VALUES ($1, $2, $3, $4, $5)`

func AddAlbumTrack(
	ctx context.Context,
	dbtx DBTX,
	albumID uuid.UUID,
	trackID uuid.UUID,
	quality Quality,
	discNo int16,
	trackNo int16,
) (AlbumTrack, error) {
	albumTrack := AlbumTrack{
		AlbumID: albumID,
		TrackID: trackID,
		Quality: quality,
		DiscNo:  discNo,
		TrackNo: trackNo,
	}

	_, err := dbtx.Exec(
		ctx,
		INSERT_ALBUM_TRACK_QUERY,
		albumTrack.AlbumID,
		albumTrack.TrackID,
		albumTrack.Quality,
		albumTrack.DiscNo,
		albumTrack.TrackNo,
	)
	if err != nil {
		return AlbumTrack{}, err
	}
	return albumTrack, nil
}
