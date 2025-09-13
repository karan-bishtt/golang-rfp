package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/karan-bishtt/rfp-quote-service/internal/database"
	"github.com/karan-bishtt/rfp-quote-service/internal/middleware"
	"github.com/karan-bishtt/rfp-quote-service/internal/models"
	"github.com/karan-bishtt/rfp-quote-service/internal/services"
	"github.com/karan-bishtt/rfp-quote-service/internal/utils"

	"github.com/gorilla/mux"
)

type RFPController struct {
	notificationService *services.NotificationService
	authService         *services.AuthService
}

type RFPResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON strings in "YYYY-MM-DD" format.
type DateOnly time.Time

// UnmarshalJSON implements the json.Unmarshaler interface.
// It parses the JSON string into a time.Time object using a
// specific layout.
func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := string(b)
	// Remove leading and trailing quotes from the JSON string.
	s = s[1 : len(s)-1]

	// The layout string "2006-01-02" is the standard Go format
	// for YYYY-MM-DD.
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = DateOnly(t)
	return nil
}

type CreateRFPRequest struct {
	Title       string   `json:"title" validate:"required,max=255"`
	Description string   `json:"description"`
	Quantity    int      `json:"quantity" validate:"min=1"`
	LastDate    DateOnly `json:"date" validate:"required"`
	MinAmount   float64  `json:"min_amount" validate:"min=0"`
	MaxAmount   float64  `json:"max_amount" validate:"min=0"`
	CategoryID  uint     `json:"category" validate:"required"`
	VendorIDs   []uint   `json:"vendor,omitempty"` // Specific vendors to notify
}

type DeleteRFPRequest struct {
	ID uint `json:"id" validate:"required"`
}

func NewRFPController() *RFPController {
	return &RFPController{
		notificationService: services.NewNotificationService(),
		authService:         services.NewAuthService(),
	}
}

func respondWithJSON(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := RFPResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateRFP creates a new RFP request
func (rc *RFPController) CreateRFP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Get user info from JWT middleware
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, 401, "You are not login", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "admin" {
		respondWithJSON(w, 400, "You are not Admin", nil)
		return
	}

	var req CreateRFPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error counting vendors: %v", err)
		respondWithJSON(w, 400, "Invalid request format", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, 400, err.Error(), nil)
		return
	}

	// Validate min/max amount
	if req.MaxAmount < req.MinAmount {
		respondWithJSON(w, 400, "Max amount must be greater than min amount", nil)
		return
	}

	// Validate last date (should be future)
	parsedDate := time.Time(req.LastDate)
	if parsedDate.Before(time.Now()) {
		respondWithJSON(w, 400, "Last date must be in the future", nil)
		return
	}

	if len(req.VendorIDs) <= 0 {
		respondWithJSON(w, 400, "Select at least one vendor", nil)
		return
	}

	// Create RFP
	rfp := models.RFP{
		Title:       req.Title,
		Description: req.Description,
		Quantity:    req.Quantity,
		LastDate:    parsedDate,
		MinAmount:   req.MinAmount,
		MaxAmount:   req.MaxAmount,
		Status:      models.RFPStatusOpen,
		CategoryID:  &req.CategoryID,
		UserID:      userID,
		IsActive:    true,
	}

	// Start transaction
	tx := database.DB.Begin()
	if err := tx.Create(&rfp).Error; err != nil {
		tx.Rollback()
		respondWithJSON(w, 500, "Failed to create RFP", nil)
		return
	}

	// Add specific vendors
	for _, vendorID := range req.VendorIDs {
		rfpVendor := models.RFPVendor{
			RFPID:    rfp.ID,
			VendorID: vendorID,
		}
		tx.Create(&rfpVendor)
	}
	tx.Commit()
	// Send notifications to vendors
	go func() {
		var vendorEmails []string

		if len(req.VendorIDs) > 0 {
			// Send to specific vendors
			vendorEmails = rc.authService.GetVendorEmailsByIDs(req.VendorIDs)
		} else {
			// Send to all vendors in the category
			vendorEmails = rc.authService.GetVendorEmailsByCategory(req.CategoryID)
		}

		// Send notification emails
		log.Println("email generated start")
		for _, email := range vendorEmails {
			log.Println("sending email to", email)
			subject := "New RFP Request: " + rfp.Title
			content := fmt.Sprintf(`
				A new RFP request has been created.
				
				Title: %s
				Description: %s
				Quantity: %d
				Budget: $%.2f - $%.2f
				Last Date: %s
				
				Please login to view details and submit your quote.
			`, rfp.Title, rfp.Description, rfp.Quantity, rfp.MinAmount, rfp.MaxAmount, rfp.LastDate.Format("2006-01-02"))

			rc.notificationService.SendEmail(email, subject, content)
		}
	}()

	respondWithJSON(w, 200, "New RFP Request is created", rfp)
}

// GetRFPs lists all RFPs (admin only)
func (rc *RFPController) GetRFPs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Check authentication and admin role
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, 401, "You are not login", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "admin" {
		respondWithJSON(w, 400, "You are not Admin", nil)
		return
	}

	// Query parameters for filtering
	status := r.URL.Query().Get("status")
	categoryID := r.URL.Query().Get("category_id")

	query := database.DB.Where("user_id = ?", userID).Preload("Quotes")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if categoryID != "" {
		if catID, err := strconv.ParseUint(categoryID, 10, 32); err == nil {
			query = query.Where("category_id = ?", catID)
		}
	}

	var rfps []models.RFP
	if err := query.Order("created_at DESC").Find(&rfps).Error; err != nil {
		respondWithJSON(w, 500, "Failed to fetch RFPs", nil)
		return
	}

	respondWithJSON(w, 200, "success", rfps)
}

// DeleteRFP removes an RFP (admin only)
func (rc *RFPController) DeleteRFP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Check authentication and admin role
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, 401, "You are not login", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "admin" {
		respondWithJSON(w, 400, "You are not Admin", nil)
		return
	}

	vars := mux.Vars(r)
	rfpIDStr := vars["id"]

	rfpID, err := strconv.ParseUint(rfpIDStr, 10, 32)
	if err != nil {
		respondWithJSON(w, 400, "Invalid RFP ID", nil)
		return
	}

	// Find RFP (ensure it belongs to the admin)
	var rfp models.RFP
	if err := database.DB.Where("id = ? AND user_id = ?", rfpID, userID).First(&rfp).Error; err != nil {
		respondWithJSON(w, 402, "Rfp request Not found", nil)
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	if err := tx.Delete(&rfp).Error; err != nil {
		tx.Rollback()
		respondWithJSON(w, 500, "Failed to delete RFP", nil)
		return
	}
	tx.Commit()

	// Return updated list
	var remainingRFPs []models.RFP
	database.DB.Where("user_id = ?", userID).Find(&remainingRFPs)

	respondWithJSON(w, 200, "Rfp request successfully deleted", remainingRFPs)
}

func (rc *RFPController) UpdateRFPStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Check authentication and admin role
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, 401, "You are not logged in", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "admin" {
		respondWithJSON(w, 400, "You are not Admin", nil)
		return
	}

	// Get status from the request body (could also be from query parameters)
	var request struct {
		Status string `json:"status" validate:"required,oneof=open closed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithJSON(w, 400, "Invalid request format", nil)
		return
	}

	// Ensure the status is either "open" or "closed"
	var newStatus models.RFPStatus
	var IsActive bool = true
	switch request.Status {
	case "open":
		newStatus = models.RFPStatusOpen
	case "closed":
		newStatus = models.RFPStatusClosed
		IsActive = false
	default:
		respondWithJSON(w, 400, "Invalid status value. It must be 'open' or 'closed'", nil)
		return
	}

	// Extract RFP ID from the URL
	vars := mux.Vars(r)
	rfpIDStr := vars["id"]

	// Parse the RFP ID
	rfpID, err := strconv.ParseUint(rfpIDStr, 10, 32)
	if err != nil {
		respondWithJSON(w, 400, "Invalid RFP ID", nil)
		return
	}

	// Find the RFP
	var rfp models.RFP
	if err := database.DB.Where("id = ? AND user_id = ?", rfpID, userID).First(&rfp).Error; err != nil {
		respondWithJSON(w, 402, "RFP request not found", nil)
		return
	}

	// Start a transaction for updating status
	tx := database.DB.Begin()

	// Update the RFP status
	rfp.Status = newStatus
	rfp.IsActive = IsActive
	if err := tx.Save(&rfp).Error; err != nil {
		tx.Rollback()
		respondWithJSON(w, 500, "Failed to update RFP status", nil)
		return
	}

	// Commit the transaction
	tx.Commit()

	// Return the updated list of RFPs
	var remainingRFPs []models.RFP
	database.DB.Where("user_id = ?", userID).Find(&remainingRFPs)

	respondWithJSON(w, 200, "RFP status updated successfully", remainingRFPs)
}

// GetRFPQuotes gets all quotes for RFPs created by the admin
func (rc *RFPController) GetRFPQuotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	log.Println("hello get rfp quotes clled")

	// Check authentication and admin role
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, http.StatusUnauthorized, "You are not logged in", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "admin" {
		respondWithJSON(w, http.StatusForbidden, "You are not an admin", nil)
		return
	}

	vars := mux.Vars(r)
	rfpIDStr := vars["id"]

	// Parse RFP ID
	rfpID, err := strconv.ParseUint(rfpIDStr, 10, 64)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid RFP ID", nil)
		return
	}

	// Fetch all quotes for this RFP
	var quotes []models.RFPQuote
	if err := database.DB.
		Where("rfp_id = ?", rfpID).
		Preload("RFP").
		Preload("Vendor"). // Add this to load vendor info
		Find(&quotes).Error; err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to fetch quotes", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, "success", quotes)
}
