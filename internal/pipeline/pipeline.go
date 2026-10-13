package pipeline

import (
	"audio-streaming/internal/audio"
	"audio-streaming/internal/dsp"
	"context"
	"errors"
)

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

func (p *Pipeline) Run(ctx context.Context, path string) (
	audio.BlockReader, error,
) {
	sourcInfo, err := p.d.Decoder.Probe(ctx, path)
	if err != nil {
		return &audio.SimpleBlockReader{}, err
	}

	plan, err := MakePlan(sourcInfo)
	if err != nil {
		return &audio.SimpleBlockReader{}, err
	}

	decodeSpec, err := decodeSpec(sourcInfo, plan)
	if err != nil {
		return &audio.SimpleBlockReader{}, err
	}

	blockReader, err := p.d.Decoder.Open(ctx, path, decodeSpec)
	if err != nil {
		return &audio.SimpleBlockReader{}, err
	}

	return blockReader, nil
}

func decodeSpec(info audio.SourceInfo, plan Plan) (audio.DecodeSpec, error) {
	spec := audio.DecodeSpec{SampleRate: targetRate}

	if plan.Mix != MixNone {
		m, err := dsp.MixMatrix(info.Channels, info.Layout)
		if err != nil {
			return audio.DecodeSpec{}, err
		}
		spec.Mix = m
	}
	return spec, nil
}
