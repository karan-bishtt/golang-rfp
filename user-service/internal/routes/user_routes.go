package routes

import (
	"github.com/gorilla/mux"
	"github.com/karan-bishtt/user-service/internal/controllers"
)

func SetupUsersRoutes() *mux.Router {
	router := mux.NewRouter()

	userController := controllers.NewUserController()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Notification routes
	api.HandleFunc("/vendors", userController.GetVendors).Methods("GET")
	api.HandleFunc("/vendors/{id}", userController.GetVendor).Methods("GET")
	api.HandleFunc("/vendors/{id}/approve", userController.ApproveVendor).Methods("POST")

	return router
}
