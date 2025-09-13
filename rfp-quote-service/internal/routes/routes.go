package routes

import (
	"github.com/karan-bishtt/rfp-quote-service/internal/controllers"
	"github.com/karan-bishtt/rfp-quote-service/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Controllers
	rfpController := controllers.NewRFPController()
	quoteController := controllers.NewQuoteController()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Apply auth middleware to all routes
	api.Use(middleware.AuthMiddleware)

	// RFP routes (Admin only)
	adminRoutes := api.PathPrefix("/rfp").Subrouter()
	adminRoutes.Use(middleware.RequireRole("admin"))

	adminRoutes.HandleFunc("", rfpController.GetRFPs).Methods("GET")
	adminRoutes.HandleFunc("", rfpController.CreateRFP).Methods("POST")
	adminRoutes.HandleFunc("/{id:[0-9]+}", rfpController.DeleteRFP).Methods("DELETE")
	adminRoutes.HandleFunc("/{id:[0-9]+}", rfpController.UpdateRFPStatus).Methods("PUT")
	adminRoutes.HandleFunc("/quotes/{id:[0-9]+}", rfpController.GetRFPQuotes).Methods("GET")

	// Quote routes (Vendor only)
	vendorRoutes := api.PathPrefix("/quote").Subrouter()
	vendorRoutes.Use(middleware.RequireRole("vendor"))

	vendorRoutes.HandleFunc("", quoteController.SubmitQuote).Methods("POST")
	vendorRoutes.HandleFunc("/my-quotes", quoteController.GetVendorRFPs).Methods("GET")
	vendorRoutes.HandleFunc("/available-rfps", quoteController.GetAvailableRFPs).Methods("GET")

	return router
}
