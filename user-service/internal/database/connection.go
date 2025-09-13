package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(databaseURL string) (*gorm.DB, error) {
	var err error
	fmt.Println("InitDB file initialize")
	fmt.Printf("Connecting to database: %s\n", maskPassword(databaseURL))

	// Configure GORM with better settings
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// Add some performance optimizations
		PrepareStmt: true,
		// Disable foreign key constraints during migration (useful for development)
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	DB, err = gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	fmt.Println("Testing database connection...")
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	fmt.Println("✅ Database connection successful!")

	return DB, nil
}

// Helper function to mask password in database URL for logging
func maskPassword(databaseURL string) string {
	// Simple masking - you might want to use regex for better parsing
	// This is just for logging purposes
	if len(databaseURL) > 20 {
		return databaseURL[:20] + "***[MASKED]***"
	}
	return "***[MASKED]***"
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
