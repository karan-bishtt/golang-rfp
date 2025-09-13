package database

import (
	"fmt"
	"time"

	"github.com/karan-bishtt/auth-service/internal/models"

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

	// Auto migrate tables
	fmt.Println("Running auto migrate...")
	err = DB.AutoMigrate(
		&models.User{},
		&models.VendorDetails{},
		&models.Permission{},
		&models.UserPermission{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}
	fmt.Println("✅ Auto migration completed!")

	// Seed default permissions if needed
	fmt.Println("Adding default permissions...")
	if err := seedDefaultPermissions(); err != nil {
		return nil, fmt.Errorf("failed to seed permissions: %w", err)
	}
	fmt.Println("✅ Default permissions seeded!")

	return DB, nil
}

func seedDefaultPermissions() error {
	permissions := []models.Permission{
		{Name: "create_rfp", Description: "Create RFP", Resource: "rfp", Action: "create"},
		{Name: "read_rfp", Description: "Read RFP", Resource: "rfp", Action: "read"},
		{Name: "update_rfp", Description: "Update RFP", Resource: "rfp", Action: "update"},
		{Name: "delete_rfp", Description: "Delete RFP", Resource: "rfp", Action: "delete"},
		{Name: "create_quote", Description: "Create Quote", Resource: "quote", Action: "create"},
		{Name: "read_quote", Description: "Read Quote", Resource: "quote", Action: "read"},
		{Name: "update_quote", Description: "Update Quote", Resource: "quote", Action: "update"},
		{Name: "delete_quote", Description: "Delete Quote", Resource: "quote", Action: "delete"},
		{Name: "manage_users", Description: "Manage Users", Resource: "user", Action: "manage"},
		{Name: "manage_categories", Description: "Manage Categories", Resource: "category", Action: "manage"},
	}

	for _, permission := range permissions {
		var existingPermission models.Permission
		result := DB.Where("name = ?", permission.Name).First(&existingPermission)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Permission doesn't exist, create it
				if err := DB.Create(&permission).Error; err != nil {
					return fmt.Errorf("failed to create permission %s: %w", permission.Name, err)
				}
				fmt.Printf("  ✅ Created permission: %s\n", permission.Name)
			} else {
				return fmt.Errorf("failed to query permission %s: %w", permission.Name, result.Error)
			}
		} else {
			fmt.Printf("  ℹ️  Permission already exists: %s\n", permission.Name)
		}
	}

	return nil
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
