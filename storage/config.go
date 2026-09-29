package storage

import "os"

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

func ConfigFromEnv() Config {
	return Config{
		Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:8333"),
		AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		Bucket:    getEnv("MINIO_BUCKET", "audio"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
