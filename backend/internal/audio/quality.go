package audio

import "strings"

// Quality is an application tier, not a claim that a lossy codec is lossless.
type Quality string

const (
	QualityLow      Quality = "low"
	QualityOptimal  Quality = "optimal"
	QualityLossless Quality = "lossless"
)

// ClassifyQuality uses source codec and average bitrate in bits/second.
// Thresholds are product policy for the original stream, without transcoding.
// Missing bitrate on lossy/unknown codecs is conservatively classified as low.
func ClassifyQuality(info SourceInfo) Quality {
	codec := strings.ToLower(strings.TrimSpace(info.Codec))
	if info.Lossless || strings.HasPrefix(codec, "pcm_") {
		return QualityLossless
	}
	switch codec {
	case "flac", "alac", "wavpack", "ape", "tta", "tak", "truehd", "mlp":
		return QualityLossless
	}
	high, middle := int64(320000), int64(192000)
	switch codec {
	case "aac", "vorbis", "wmav2", "wmapro":
		high, middle = 256000, 192000
	case "opus":
		high, middle = 192000, 128000
	}
	if info.BitRate >= high {
		return QualityLossless
	}
	if info.BitRate >= middle {
		return QualityOptimal
	}
	return QualityLow
}
