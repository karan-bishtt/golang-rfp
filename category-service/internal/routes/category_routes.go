package routes

import (
	"github.com/karan-bishtt/category-service/internal/controllers"

	"github.com/gorilla/mux"
	"github.com/karan-bishtt/category-service/internal/middleware"
)

func SetupCategoryRoutes() *mux.Router {
	router := mux.NewRouter()

	categoryController := controllers.NewCategoryController()

	// Public routes (if needed)
	api := router.PathPrefix("/api/v1").Subrouter()

	// Public routes (without authentication)
	api.HandleFunc("/categories", categoryController.GetCategories).Methods("GET")
	api.HandleFunc("/categories/{id:[0-9]+}", categoryController.GetCategory).Methods("GET")

	// Protected routes - require authentication and authorization
	protected := api.PathPrefix("/categories").Subrouter()

	// Apply authentication middleware here
	protected.Use(middleware.AuthMiddleware)
	protected.Use(middleware.RequireRole("admin"))

	// Category CRUD routes (admin protected)
	protected.HandleFunc("", categoryController.CreateCategory).Methods("POST")
	protected.HandleFunc("/{id:[0-9]+}", categoryController.UpdateCategory).Methods("PUT")
	protected.HandleFunc("/{id:[0-9]+}", categoryController.DeleteCategory).Methods("DELETE")

	return router

}
