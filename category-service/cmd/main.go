package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/karan-bishtt/category-service/config"
	"github.com/karan-bishtt/category-service/internal/database"
	"github.com/karan-bishtt/category-service/internal/routes"
)

func main() {
	fmt.Println("staring category service")

	cfg := config.Load()

	// Initialize database
	_, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	router := routes.SetupCategoryRoutes()
	// CORS options
	handler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // OK because we are NOT using credentials
		handlers.AllowedMethods([]string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.OptionStatusCode(204),
	)(router)

	// Optional: log requests
	// If you use cookies/session across origins, uncomment both lines below:
	// creds := handlers.AllowCredentials()
	// NOTE: When using AllowCredentials, you MUST NOT use "*" for AllowedOrigins.
	handler = handlers.LoggingHandler(os.Stdout, handler)

	log.Printf("Auth Service started on post %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))

}
