package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/audio"
)

func TestNewEntryRejectsInvalidTags(t *testing.T) {
	tests := []struct {
		name   string
		change func(rec *audio.Record)
	}{
		{"missing track ID", func(r *audio.Record) {
			r.TrackId = uuid.Nil
		}},
		{"blank title", func(r *audio.Record) { r.Tags.Title = " " }},
		{"blank artist", func(r *audio.Record) { r.Tags.Artist = "" }},
		{"zero duration", func(r *audio.Record) {
			r.SourceInfo.Samples = 0
		}},
		{"unknown sample rate", func(r *audio.Record) {
			r.SourceInfo.SampleRate = 0
		}},
		{"album without track number", func(r *audio.Record) {
			r.Tags.TrackNumber = 0
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := testRecord("Duality")
			tt.change(&rec)

			_, err := newEntry(rec, nil)

			assert.ErrorIs(t, err, ErrInvalidRecord)
		})
	}
}

func TestNewEntryTrimsTags(t *testing.T) {
	rec := testRecord(" Duality ")
	rec.Tags.Artist = " Slipknot\t"
	rec.Tags.Album = " Vol. 3 "

	e, err := newEntry(rec, nil)

	require.NoError(t, err)
	assert.Equal(t, "Duality", e.title)
	assert.Equal(t, "Slipknot", e.artist)
	assert.Equal(t, "Vol. 3", e.album.title)
}

func TestNewEntryDefaultsAlbumArtistToArtist(t *testing.T) {
	rec := testRecord("Duality")

	e, err := newEntry(rec, nil)

	require.NoError(t, err)
	assert.Equal(t, "Slipknot", e.album.artist)
}

func TestNewEntryDefaultsDiscNumberToOne(t *testing.T) {
	rec := testRecord("Duality")

	e, err := newEntry(rec, nil)

	require.NoError(t, err)
	assert.Equal(t, int16(1), e.album.discNo)
}

func TestNewEntryAllowsTrackNumberWithoutAlbum(t *testing.T) {
	rec := testRecord("Duality")
	rec.Tags.Album = ""
	rec.Tags.TrackNumber = 0

	e, err := newEntry(rec, nil)

	require.NoError(t, err)
	assert.Nil(t, e.album)
}

func TestParseYear(t *testing.T) {
	tests := []struct {
		date string
		want *int16
	}{
		{"2004", ptr(int16(2004))},
		{"2004-05-25", ptr(int16(2004))},
		{" 2004 ", ptr(int16(2004))},
		{"", nil},
		{"04", nil},
		{"May 2004", nil},
		{"0999", nil},
	}
	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			got := parseYear(tt.date)

			assert.Equal(t, tt.want, got)
		})
	}
}

func ptr[T any](v T) *T { return &v }
