package main

import (
	"log"
	"net/http"

	"github.com/karan-bishtt/notification-service/config"
	"github.com/karan-bishtt/notification-service/internal/database"
	"github.com/karan-bishtt/notification-service/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	_, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")

	// Setup routes
	router := routes.SetupNotificationRoutes()

	log.Printf("Notification service starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
