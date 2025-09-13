package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	Environment string
	JWTSecret   string

	// SMTP configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8082"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:admin@123@192.168.10.166:5432/rfp"),
		JWTSecret:   getEnv("JWT_SECRET", "your_jwt_secret"),
		Environment: getEnv("ENVIRONMENT", "development"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", "bishtprateek470@gmail.com"),
		SMTPPassword: getEnv("SMTP_PASSWORD", "svozrnlgcndatoav"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
