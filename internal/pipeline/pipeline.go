package pipeline

import "errors"

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
