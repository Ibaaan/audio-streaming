package s3store

import "audio-streaming/internal/env"

type Config struct {
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
	bucket    string
}

// ConfigFromEnv reads the storage settings from the environment. It panics
// if any of them is unset, so a misconfigured server never starts.
func ConfigFromEnv() Config {
	return Config{
		endpoint:  env.MustGet("MINIO_ENDPOINT"),
		accessKey: env.MustGet("MINIO_ACCESS_KEY"),
		secretKey: env.MustGet("MINIO_SECRET_KEY"),
		useSSL:    env.MustGet("MINIO_USE_SSL") == "true",
		bucket:    env.MustGet("MINIO_BUCKET"),
	}
}
