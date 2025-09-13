package config

import "os"

type Config struct {
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	Environment            string
	NotificationServiceURL string
}

/**
// Returns pointer - only one copy in memory
func Load() *Config {
    return &Config{...} // Creates once, returns reference
}

// vs returning value - creates multiple copies
func Load() Config {
    return Config{...} // Copy created here, another copy when assigned
}

->
return would be Config type  so return Config with & pointer
*/

func Load() *Config {
	return &Config{
		Port:                   getEnv("PORT", "8081"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://postgres:admin@123@192.168.10.166:5432/rfp"),
		JWTSecret:              getEnv("JWT_SECRET", "your_jwt_secret"),
		Environment:            getEnv("ENVIRONMENT", "development"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8082"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
