package pipeline

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"audio-streaming/internal/audio"

	"github.com/google/uuid"
)

// ErrInvalidSource is returned by Run for a file that is not audio it
// can read, such as an empty file or one without an audio stream.
var ErrInvalidSource = errors.New("pipeline: invalid source")

type Deps struct {
	Decoder  Decoder
	Meta     MetaExtractor
	Encoder  Encoder
	Verifier Verifier
	Blobs    Uploader
	Records  RecordSaver
}

type Pipeline struct {
	d      Deps
	tmpDir string
}

func New(d Deps, tmpDir string) (*Pipeline, error) {
	if d.Decoder == nil ||
		d.Meta == nil ||
		d.Encoder == nil ||
		d.Verifier == nil ||
		d.Blobs == nil ||
		d.Records == nil {
		return nil, errors.New("pipeline: missing dependency")
	}
	return &Pipeline{d: d, tmpDir: tmpDir}, nil
}

// Run imports the audio file at path: it uploads the file unchanged to
// audio/<id><ext> and then saves the track record. The record is saved
// last, so a saved track always has its audio stored. Encoding and
// verification are not done yet. The ID is returned only after the record is saved.
func (p *Pipeline) Run(ctx context.Context, path string) (uuid.UUID, error) {
	info, err := p.d.Decoder.Probe(ctx, path)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: probing: %w", ErrInvalidSource, err)
	}
	tags, cover, err := p.d.Meta.Extract(ctx, path)
	if err != nil {
		return uuid.Nil, fmt.Errorf("reading metadata: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	if err := p.uploadAudio(ctx, id, path); err != nil {
		return uuid.Nil, err
	}

	rec := audio.Record{
		TrackID:    id,
		SourceInfo: info,
		Quality:    audio.ClassifyQuality(info),
		Tags:       tags,
		Cover:      cover,
	}
	if err := p.d.Records.Save(ctx, rec); err != nil {
		return uuid.Nil, fmt.Errorf("saving record: %w", err)
	}
	return id, nil
}

func (p *Pipeline) uploadAudio(ctx context.Context, id uuid.UUID, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	st, err := f.Stat()
	if err != nil {
		return err
	}
	if st.Size() == 0 {
		return fmt.Errorf("%w: %s is empty", ErrInvalidSource, path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	key := "audio/" + id.String() + ext
	_, err = p.d.Blobs.Upload(ctx, key, f, st.Size(), contentType(ext))
	return err
}

func contentType(ext string) string {
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}
