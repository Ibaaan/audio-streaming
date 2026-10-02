package track

import "time"

type Record struct {
	ID       string
	AudioKey string
	CoverKey string
	Tags     Tags
	Duration time.Duration
}

type SourceInfo struct {
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
	CompressionLevel  int
	TotalFrames       int64
	SeekPointInterval int64
}

type Expectations struct {
	TotalFrames int64
	SeekPoints  int
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
