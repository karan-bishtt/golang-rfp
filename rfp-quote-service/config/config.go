package config

import (
	"os"
)

type Config struct {
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	Environment            string
	AuthServiceURL         string
	NotificationServiceURL string
	CategoryServiceURL     string
	UploadDir              string
}

func Load() *Config {
	return &Config{
		Port:                   getEnv("PORT", "8084"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://postgres:admin@123@192.168.10.166:5432/rfp"),
		JWTSecret:              getEnv("JWT_SECRET", "your_jwt_secret"),
		Environment:            getEnv("ENVIRONMENT", "development"),
		AuthServiceURL:         getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8082"),
		CategoryServiceURL:     getEnv("CATEGORY_SERVICE_URL", "http://localhost:8083"),
		UploadDir:              getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
