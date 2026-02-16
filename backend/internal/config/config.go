package config

import (
	"os"
)

type Config struct {
	DatabaseURL   string
	MinioEndpoint string
	MinioUser     string
	MinioPassword string
	MinioUseSSL   bool

	KaggleUsername string
	KaggleKey      string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://finetune:finetune_pass@postgres:5432/finetune_db"),
		MinioEndpoint: getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinioUser:     getEnv("MINIO_ROOT_USER", "minioadmin"),
		MinioPassword: getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinioUseSSL:   getEnv("MINIO_USE_SSL", "false") == "true",
		KaggleUsername: getEnv("KAGGLE_USERNAME", ""),
		KaggleKey:      getEnv("KAGGLE_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
