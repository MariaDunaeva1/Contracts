package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// App
	AppEnv     string
	AppVersion string
	Port       string

	// Database
	DatabaseURL              string
	DBMaxConnections         int
	DBMaxIdleConnections     int
	DBConnectionMaxLifetime  time.Duration

	// MinIO
	MinioEndpoint string
	MinioUser     string
	MinioPassword string
	MinioUseSSL   bool
	MinioBucket   string

	// Kaggle
	KaggleUsername string
	KaggleKey      string

	// Security
	AllowedOrigins                string
	RateLimitRequestsPerMinute    int
	RateLimitExpensiveEndpoints   int
	MaxRequestSizeMB              int

	// Workers
	WorkerPoolSize int
	WorkerTimeout  time.Duration

	// Logging
	LogLevel  string
	LogFormat string

	// Monitoring
	MetricsEnabled bool
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		AppEnv:                       getEnv("APP_ENV", "development"),
		AppVersion:                   getEnv("APP_VERSION", "1.0.0"),
		Port:                         getEnv("PORT", "8080"),
		DatabaseURL:                  getEnv("DATABASE_URL", "postgres://finetune:finetune_pass@postgres:5432/finetune_db"),
		DBMaxConnections:             getEnvInt("DB_MAX_CONNECTIONS", 25),
		DBMaxIdleConnections:         getEnvInt("DB_MAX_IDLE_CONNECTIONS", 5),
		DBConnectionMaxLifetime:      getEnvDuration("DB_CONNECTION_MAX_LIFETIME", 5*time.Minute),
		MinioEndpoint:                getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinioUser:                    getEnv("MINIO_USER", "minioadmin"),
		MinioPassword:                getEnv("MINIO_PASSWORD", "minioadmin"),
		MinioUseSSL:                  getEnv("MINIO_USE_SSL", "false") == "true",
		MinioBucket:                  getEnv("MINIO_BUCKET", "finetune-models"),
		KaggleUsername:               getEnv("KAGGLE_USERNAME", ""),
		KaggleKey:                    getEnv("KAGGLE_KEY", ""),
		AllowedOrigins:               getEnv("ALLOWED_ORIGINS", "*"),
		RateLimitRequestsPerMinute:   getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
		RateLimitExpensiveEndpoints:  getEnvInt("RATE_LIMIT_EXPENSIVE_ENDPOINTS", 10),
		MaxRequestSizeMB:             getEnvInt("MAX_REQUEST_SIZE_MB", 10),
		WorkerPoolSize:               getEnvInt("WORKER_POOL_SIZE", 5),
		WorkerTimeout:                getEnvDuration("WORKER_TIMEOUT", 24*time.Hour),
		LogLevel:                     getEnv("LOG_LEVEL", "info"),
		LogFormat:                    getEnv("LOG_FORMAT", "console"),
		MetricsEnabled:               getEnv("METRICS_ENABLED", "true") == "true",
	}

	// Validate required fields in production
	if cfg.AppEnv == "production" {
		required := map[string]string{
			"DATABASE_URL":    cfg.DatabaseURL,
			"KAGGLE_USERNAME": cfg.KaggleUsername,
			"KAGGLE_KEY":      cfg.KaggleKey,
		}
		for key, value := range required {
			if value == "" {
				return nil, fmt.Errorf("required environment variable %s is not set", key)
			}
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
