package routes

import (
	"github.com/gorilla/mux"
	"github.com/karan-bishtt/notification-service/internal/controllers"
)

func SetupNotificationRoutes() *mux.Router {
	router := mux.NewRouter()

	notificationController := controllers.NewNotificationController()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	// api.Use(middleware.AuthMiddleware)

	// Notification routes
	api.HandleFunc("/send-email", notificationController.SendEmail).Methods("POST")
	api.HandleFunc("/status", notificationController.GetNotificationStatus).Methods("GET")

	return router
}
