package audio

import (
	"time"

	"github.com/google/uuid"
)

// represent the audio metadata
type Record struct {
	TrackID    uuid.UUID
	SourceInfo SourceInfo
	Tags       Tags
	Cover      *Cover // nil if the file has no embedded artwork
}

func (a Record) Duration() time.Duration { return a.SourceInfo.Duration() }

// technical properties of the source audio stream
type SourceInfo struct {
	Codec      string
	Lossless   bool
	SampleRate int
	BitDepth   int
	Channels   int
	Layout     string
	Samples    int64 // total samples per channel
}

func (f SourceInfo) Duration() time.Duration {
	if f.SampleRate == 0 {
		return 0
	}
	return time.Duration(f.Samples) * time.Second / time.Duration(f.SampleRate)
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
