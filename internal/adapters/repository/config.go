package repository

import "os"

type config struct {
	user   string
	pass   string
	host   string
	port   string
	dbname string
}

func ConfigFromEnv() config {
	return config{
		user:   getEnv("POSTGRES_USER", "audio"),
		pass:   getEnv("POSTGRES_PASSWORD", "audio"),
		host:   getEnv("POSTGRES_HOST", "localhost"),
		port:   getEnv("POSTGRES_PORT", "5432"),
		dbname: getEnv("POSTGRES_DB", "audio"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
