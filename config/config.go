package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	Environment    string
	AllowedOrigins []string
	S3BucketName   string
	S3Region       string
	UseS3Storage   bool
	IntrospectURL  string
	Region         string
	SecretKey      string
	AccessKey      string
	BucketName     string
	LogLevel       string
	LogFormat      string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"*"}),
		S3BucketName:   getEnv("S3_BUCKET_NAME", ""),
		S3Region:       getEnv("S3_REGION", "us-east-1"),
		UseS3Storage:   getEnv("USE_S3_STORAGE", "false") == "true",
		IntrospectURL:  getEnv("INTROSPECT_URL", "http://localhost:8080/"),
		Region:         getEnv("SPACES_REGION", "sgp1"), // e.g., "nyc3", "sfo3", "ams3", "sgp1", "fra1"
		AccessKey:      getEnv("SPACE_ACCESS_KEY", ""),  // Your Spaces access key
		SecretKey:      getEnv("SPACE_SECRET_KEY", ""),  // Your Spaces secret key
		BucketName:     getEnv("SPACE_BUCKET", ""),      // Your Space name, e.g., "myapp-images"
		LogLevel:       getEnv("LOG_LEVEL", "INFO"),     // DEBUG, INFO, WARN, ERROR, FATAL
		LogFormat:      getEnv("LOG_FORMAT", "text"),    // text or json
	}

	// Build database URL
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "user_management")

	cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsSlice(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	// Handle single wildcard
	if value == "*" {
		return []string{"*"}
	}

	// Split by comma and clean up
	origins := strings.Split(value, ",")
	var result []string
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return fallback
	}

	return result
}
