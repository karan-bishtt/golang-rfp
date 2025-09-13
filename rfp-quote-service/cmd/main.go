package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/karan-bishtt/rfp-quote-service/config"
	"github.com/karan-bishtt/rfp-quote-service/internal/database"
	"github.com/karan-bishtt/rfp-quote-service/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")

	// Test database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Setup routes
	router := routes.SetupRoutes()
	handler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // OK because we are NOT using credentials
		handlers.AllowedMethods([]string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.OptionStatusCode(204),
	)(router)

	log.Printf("RFP-Quote service starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)

	// Start server
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
