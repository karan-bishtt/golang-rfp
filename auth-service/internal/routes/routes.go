package routes

import (
	"github.com/gorilla/mux"
	"github.com/karan-bishtt/auth-service/internal/controllers"
	"github.com/karan-bishtt/auth-service/internal/middleware"
)

func SetupAuthRoutes() *mux.Router {
	router := mux.NewRouter()

	// Controllers
	authController := controllers.NewAuthController()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Public Auth routes (No authentication required)
	authRoutes := api.PathPrefix("/auth").Subrouter()
	authRoutes.HandleFunc("/register-vendor", authController.RegisterVendor).Methods("POST")
	authRoutes.HandleFunc("/register-admin", authController.RegisterAdmin).Methods("POST")
	authRoutes.HandleFunc("/login", authController.Login).Methods("POST")
	authRoutes.HandleFunc("/users/{id:[0-9]+}", authController.GetVendorById).Methods("GET")

	// Apply auth middleware to protected routes (not for /auth)

	// Admin routes (Require 'admin' role)
	adminRoutes := api.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(middleware.AuthMiddleware)
	adminRoutes.Use(middleware.RequireRole("admin"))
	adminRoutes.HandleFunc("/get-vendors", authController.GetVendors).Methods("GET")
	adminRoutes.HandleFunc("/get-vendors/{id:[0-9]+}", authController.GetVendorsByCategory).Methods("GET")
	adminRoutes.HandleFunc("/approve-vendors", authController.ApproveVendor).Methods("POST")

	return router
}
