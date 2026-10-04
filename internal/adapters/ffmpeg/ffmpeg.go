package ffmpeg

type FFmpeg struct {
	ffmpeg  string
	ffprobe string
}

func New(ffmpegPath, ffprobePath string) *FFmpeg {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}
	return &FFmpeg{ffmpeg: ffmpegPath, ffprobe: ffprobePath}
}

func inputArgs(path string) []string {
	return []string{
		"-nostdin", "-hide_banner", "-loglevel", "error",
		"-protocol_whitelist", "file",
		"-i", "file:" + path,
	}
}
