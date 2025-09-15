package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/karan-bishtt/auth-service/config"
	"github.com/karan-bishtt/auth-service/internal/database"
	"github.com/karan-bishtt/auth-service/internal/routes"
)

func main() {
	fmt.Println("starting auth")
	cfg := config.Load()

	if _, err := database.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	router := routes.SetupAuthRoutes()

	handler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // OK because we are NOT using credentials
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.OptionStatusCode(204),
	)(router)

	// Optional: log requests
	handler = handlers.LoggingHandler(os.Stdout, handler)

	log.Printf("Auth Service started on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
