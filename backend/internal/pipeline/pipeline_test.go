package pipeline

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"audio-streaming/internal/audio"
)

type fakeDecoder struct {
	info audio.SourceInfo
	err  error
}

func (d *fakeDecoder) Probe(ctx context.Context, path string) (
	audio.SourceInfo, error,
) {
	return d.info, d.err
}

func (d *fakeDecoder) Open(
	ctx context.Context, path string, spec audio.DecodeSpec,
) (audio.BlockReader, error) {
	return nil, errors.New("not implemented")
}

func (d *fakeDecoder) OpenPCM16(ctx context.Context, path string) (
	io.ReadCloser, error,
) {
	return nil, errors.New("not implemented")
}

type fakeMeta struct {
	tags  audio.Tags
	cover *audio.Cover
}

func (m fakeMeta) Extract(ctx context.Context, path string) (
	audio.Tags, *audio.Cover, error,
) {
	return m.tags, m.cover, nil
}

type fakeEncoder struct{}

func (fakeEncoder) Encode(
	ctx context.Context, pcm io.Reader,
	opt audio.EncodeOptions, dstPath string,
) error {
	return errors.New("not implemented")
}

type fakeVerifier struct{}

func (fakeVerifier) Verify(path string, exp audio.Expectations) (
	audio.Report, error,
) {
	return audio.Report{}, errors.New("not implemented")
}

type blob struct {
	data        []byte
	contentType string
}

type fakeBlobs struct{ blobs map[string]blob }

func (b *fakeBlobs) Upload(
	ctx context.Context, key string, r io.Reader,
	size int64, contentType string,
) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	b.blobs[key] = blob{data: data, contentType: contentType}
	return int64(len(data)), nil
}

type fakeRecords struct {
	err   error
	saved []audio.Record
}

func (r *fakeRecords) Save(ctx context.Context, rec audio.Record) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, rec)
	return nil
}

type testEnv struct {
	p       *Pipeline
	decoder *fakeDecoder
	meta    *fakeMeta
	blobs   *fakeBlobs
	records *fakeRecords
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	env := &testEnv{
		decoder: &fakeDecoder{info: audio.SourceInfo{
			Codec:      "flac",
			SampleRate: 44100,
			Samples:    44100 * 252,
		}},
		meta: &fakeMeta{tags: audio.Tags{
			Title:  "Duality",
			Artist: "Slipknot",
		}},
		blobs:   &fakeBlobs{blobs: map[string]blob{}},
		records: &fakeRecords{},
	}
	p, err := New(Deps{
		Decoder:  env.decoder,
		Meta:     env.meta,
		Encoder:  fakeEncoder{},
		Verifier: fakeVerifier{},
		Blobs:    env.blobs,
		Records:  env.records,
	}, t.TempDir())
	require.NoError(t, err, "New")
	env.p = p
	return env
}

func writeFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, data, 0o600), "WriteFile")
	return path
}

var audioData = []byte("fLaC not really")

func TestRunUploadsAudioAndSavesRecord(t *testing.T) {
	env := newTestEnv(t)
	path := writeFile(t, "Duality.FLAC", audioData)

	err := env.p.Run(context.Background(), path)

	require.NoError(t, err)
	require.Len(t, env.records.saved, 1)
	rec := env.records.saved[0]
	key := "audio/" + rec.TrackId.String() + ".flac"
	require.Contains(t, env.blobs.blobs, key)
	assert.Equal(t, audioData, env.blobs.blobs[key].data)
	assert.Equal(t, "audio/flac", env.blobs.blobs[key].contentType)
	assert.Equal(t, "Duality", rec.Tags.Title)
	assert.Equal(t, int64(44100*252), rec.SourceInfo.Samples)
}

func TestRunPassesCoverToRecord(t *testing.T) {
	env := newTestEnv(t)
	cover := &audio.Cover{MIMEType: "image/png", Data: []byte("png")}
	env.meta.cover = cover
	path := writeFile(t, "a.flac", audioData)

	err := env.p.Run(context.Background(), path)

	require.NoError(t, err)
	require.Len(t, env.records.saved, 1)
	assert.Equal(t, cover, env.records.saved[0].Cover)
}

func TestRunReturnsSaveError(t *testing.T) {
	env := newTestEnv(t)
	saveErr := errors.New("track exists")
	env.records.err = saveErr
	path := writeFile(t, "a.flac", audioData)

	err := env.p.Run(context.Background(), path)

	assert.ErrorIs(t, err, saveErr)
}

func TestRunRejectsEmptyFile(t *testing.T) {
	env := newTestEnv(t)
	path := writeFile(t, "a.flac", nil)

	err := env.p.Run(context.Background(), path)

	assert.ErrorIs(t, err, ErrInvalidSource)
	assert.Empty(t, env.blobs.blobs)
	assert.Empty(t, env.records.saved)
}

func TestRunRejectsUnprobeableFile(t *testing.T) {
	env := newTestEnv(t)
	env.decoder.err = errors.New("ffprobe: no audio stream")
	path := writeFile(t, "a.txt", []byte("not audio"))

	err := env.p.Run(context.Background(), path)

	assert.ErrorIs(t, err, ErrInvalidSource)
	assert.Empty(t, env.blobs.blobs)
	assert.Empty(t, env.records.saved)
}

func TestNewRejectsMissingDependency(t *testing.T) {
	_, err := New(Deps{Blobs: &fakeBlobs{}}, t.TempDir())

	assert.Error(t, err)
}
