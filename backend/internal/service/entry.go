package service

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"audio-streaming/internal/audio"
	"audio-streaming/internal/catalog"
)

// ErrInvalidRecord is returned for a record whose tags cannot be stored,
// such as one without a title.
var ErrInvalidRecord = errors.New("service: invalid record")

// entry is an audio.Record checked and normalized into the values the
// catalog stores.
type entry struct {
	trackID    uuid.UUID
	title      string
	artist     string
	durationMs int32
	album      *albumEntry // nil if the record has no album tag
}

type albumEntry struct {
	title       string
	artist      string // the album artist, or the track artist if unset
	releaseYear *int16
	discNo      int16
	trackNo     int16
	quality     catalog.Quality
	cover       *catalog.Image
}

func newEntry(rec audio.Record, cover *catalog.Image) (entry, error) {
	tags := rec.Tags
	e := entry{
		trackID: rec.TrackID,
		title:   strings.TrimSpace(tags.Title),
		artist:  strings.TrimSpace(tags.Artist),
	}
	if e.trackID == uuid.Nil {
		return entry{}, fmt.Errorf("%w: missing track ID", ErrInvalidRecord)
	}
	if e.title == "" {
		return entry{}, fmt.Errorf("%w: missing title", ErrInvalidRecord)
	}
	if e.artist == "" {
		return entry{}, fmt.Errorf("%w: missing artist", ErrInvalidRecord)
	}

	ms := rec.Duration().Milliseconds()
	if ms <= 0 || ms > math.MaxInt32 {
		return entry{}, fmt.Errorf("%w: duration %dms out of range",
			ErrInvalidRecord, ms)
	}
	e.durationMs = int32(ms)

	albumTitle := strings.TrimSpace(tags.Album)
	if albumTitle == "" {
		return e, nil
	}
	if tags.TrackNumber <= 0 || tags.TrackNumber > math.MaxInt16 {
		return entry{}, fmt.Errorf("%w: track number %d out of range",
			ErrInvalidRecord, tags.TrackNumber)
	}
	album := albumEntry{
		title:       albumTitle,
		artist:      strings.TrimSpace(tags.AlbumArtist),
		releaseYear: parseYear(tags.Date),
		discNo:      1,
		trackNo:     int16(tags.TrackNumber),
		quality:     catalog.QualityOptimal, // TODO make this not static
		cover:       cover,
	}
	if cover != nil && cover.StoragePath == "" {
		return entry{}, fmt.Errorf("%w: cover without storage path",
			ErrInvalidRecord)
	}
	if album.artist == "" {
		album.artist = e.artist
	}
	if tags.DiscNumber > 0 && tags.DiscNumber <= math.MaxInt16 {
		album.discNo = int16(tags.DiscNumber)
	}
	if rec.SourceInfo.Lossless {
		album.quality = catalog.QualityLossless
	}
	e.album = &album
	return e, nil
}

// parseYear reads the year from the start of a date tag such as "2019"
// or "2019-04-12". It returns nil if there is no plausible year.
func parseYear(date string) *int16 {
	date = strings.TrimSpace(date)
	if len(date) < 4 {
		return nil
	}
	year, err := strconv.Atoi(date[:4])
	if err != nil || year < 1000 || year > 9999 {
		return nil
	}
	y := int16(year)
	return &y
}
