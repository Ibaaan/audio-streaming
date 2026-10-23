package s3store

import "os"

type Config struct {
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
	bucket    string
}

func ConfigFromEnv() Config {
	return Config{
		endpoint:  getEnv("MINIO_ENDPOINT", "localhost:8333"),
		accessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		secretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		useSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		bucket:    getEnv("MINIO_BUCKET", "audio"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
