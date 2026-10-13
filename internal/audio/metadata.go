package audio

import "time"

type Record struct {
	ID       string
	AudioKey string
	CoverKey string
	Tags     Tags
	Duration time.Duration
}

type SourceInfo struct {
	Codec        string
	Lossless     bool
	SampleRate   int
	BitDepth     int
	Channels     int
	Layout       string
	TotalSamples int64
}

type Tags struct {
	Title                   string
	Artist                  string
	Album                   string
	AlbumArtist             string
	TrackNumber, DiscNumber int
	Date                    string
	Extra                   map[string][]string
}

type Cover struct {
	MIMEType string
	Data     []byte
}

type EncodeOptions struct {
	CompressionLevel int
	SeekIntervalSec  int
}

type Expectations struct {
	TotalFrames int64
	SeekPoints  int
	MD5         [16]byte
}

type Report struct {
	Checks []Check
}

type Check struct {
	Name string
	OK   bool
	Got  string
	Want string
}

func (r Report) OK() bool {
	for _, c := range r.Checks {
		if !c.OK {
			return false
		}
	}
	return true
}
