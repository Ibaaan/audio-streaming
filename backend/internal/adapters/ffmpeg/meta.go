package ffmpeg

import (
	"audio-streaming/internal/audio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// maxCover is the largest embedded picture Extract accepts.
const maxCover = 16 << 20

var coverMIME = map[string]string{
	"mjpeg": "image/jpeg",
	"png":   "image/png",
	"bmp":   "image/bmp",
	"gif":   "image/gif",
	"webp":  "image/webp",
}

// Extract reads the tags and the embedded cover, if any. A cover in a
// format we don't recognise is skipped.
func (f *FFmpeg) Extract(ctx context.Context, path string) (audio.Tags, *audio.Cover, error) {
	p, err := f.probe(ctx, path)
	if err != nil {
		return audio.Tags{}, nil, err
	}
	cover, err := f.cover(ctx, path, p)
	if err != nil {
		return audio.Tags{}, nil, err
	}
	return parseTags(p), cover, nil
}

// parseTags merges container and audio stream tags (Ogg keeps them on
// the stream, most other formats on the container). Keys are matched
// case-insensitively; anything not mapped to a field lands in Extra.
func parseTags(p probeOutput) audio.Tags {
	all := map[string]string{}
	add := func(tags map[string]string) {
		for k, v := range tags {
			k = strings.ToLower(k)
			if _, ok := all[k]; !ok {
				all[k] = v
			}
		}
	}
	add(p.Format.Tags)
	if s, ok := p.audio(); ok {
		add(s.Tags)
	}

	take := func(keys ...string) string {
		var v string
		for _, k := range keys {
			if x, ok := all[k]; ok && v == "" {
				v = x
			}
			delete(all, k)
		}
		return v
	}
	t := audio.Tags{
		Title:       take("title"),
		Artist:      take("artist"),
		Album:       take("album"),
		AlbumArtist: take("album_artist", "albumartist", "album artist"),
		TrackNumber: leadingInt(take("track", "tracknumber")),
		DiscNumber:  leadingInt(take("disc", "discnumber")),
		Date:        take("date", "year"),
	}
	if len(all) > 0 {
		t.Extra = make(map[string][]string, len(all))
		for k, v := range all {
			t.Extra[k] = []string{v}
		}
	}
	return t
}

// leadingInt parses "3/12" as 3. It returns 0 when there is no number.
func leadingInt(s string) int {
	s = strings.TrimSpace(s)
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	n, _ := strconv.Atoi(s[:end])
	return n
}

func coverStream(p probeOutput) (probeStream, bool) {
	for _, s := range p.Streams {
		if s.CodecType == "video" && s.Disposition.AttachedPic == 1 {
			return s, true
		}
	}
	return probeStream{}, false
}

func (f *FFmpeg) cover(ctx context.Context, path string, p probeOutput) (*audio.Cover, error) {
	s, ok := coverStream(p)
	if !ok {
		return nil, nil
	}
	mime, ok := coverMIME[s.CodecName]
	if !ok {
		return nil, nil
	}

	args := append(inputArgs(path),
		"-map", fmt.Sprintf("0:%d", s.Index), "-frames:v", "1",
		"-c", "copy", "-f", "image2pipe", "-")
	proc, err := start(ctx, f.ffmpeg, args)
	if err != nil {
		return nil, err
	}
	defer func() { _ = proc.Close() }()

	data, err := io.ReadAll(io.LimitReader(proc, maxCover+1))
	if err != nil {
		return nil, fmt.Errorf("extracting cover: %w", err)
	}
	if len(data) > maxCover {
		return nil, errors.New("cover is larger than 16 MiB")
	}
	return &audio.Cover{MIMEType: mime, Data: data}, nil
}
