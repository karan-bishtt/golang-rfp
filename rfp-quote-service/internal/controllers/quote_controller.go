package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/karan-bishtt/rfp-quote-service/internal/database"
	"github.com/karan-bishtt/rfp-quote-service/internal/middleware"
	"github.com/karan-bishtt/rfp-quote-service/internal/models"
	"github.com/karan-bishtt/rfp-quote-service/internal/utils"
)

type QuoteController struct{}

type SubmitQuoteRequest struct {
	RFPID           uint    `json:"rfp_id" validate:"required"`
	VendorPrice     float64 `json:"item_price" validate:"min=0"`
	ItemDescription string  `json:"item_description" validate:"required"`
	Quantity        int     `json:"quantity" validate:"min=1"`
	TotalCost       float64 `json:"total_cost" validate:"min=0"`
}

func NewQuoteController() *QuoteController {
	return &QuoteController{}
}

// SubmitQuote allows vendors to submit quotes for RFPs
func (qc *QuoteController) SubmitQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Check authentication and vendor role
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, 401, "You are not login", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "vendor" {
		respondWithJSON(w, 400, "You are not a Vendor", nil)
		return
	}

	var req SubmitQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, 400, "Invalid request format", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, 400, err.Error(), nil)
		return
	}

	// Find RFP and validate it's still open
	var rfp models.RFP
	if err := database.DB.First(&rfp, req.RFPID).Error; err != nil {
		respondWithJSON(w, 404, "RFP not found", nil)
		return
	}

	if !rfp.IsOpen() {
		respondWithJSON(w, 400, "RFP is closed or expired", nil)
		return
	}

	// Check if vendor already submitted a quote
	var existingQuote models.RFPQuote
	if err := database.DB.Where("rfp_id = ? AND vendor_id = ?", req.RFPID, userID).First(&existingQuote).Error; err == nil {
		respondWithJSON(w, 400, "Quote already submitted for this RFP", nil)
		return
	}

	// Validate quote is within budget range
	if req.TotalCost < rfp.MinAmount || req.TotalCost > rfp.MaxAmount {
		respondWithJSON(w, 400, "Quote amount is outside the specified budget range", nil)
		return
	}

	// Create quote
	quote := models.RFPQuote{
		RFPID:           req.RFPID,
		VendorID:        userID,
		VendorPrice:     req.VendorPrice,
		ItemDescription: req.ItemDescription,
		Quantity:        req.Quantity,
		TotalCost:       req.TotalCost,
		Status:          "pending",
		SubmittedAt:     time.Now(),
	}

	// Start transaction
	tx := database.DB.Begin()
	if err := tx.Create(&quote).Error; err != nil {
		tx.Rollback()
		respondWithJSON(w, 500, "Failed to submit quote", nil)
		return
	}
	tx.Commit()

	respondWithJSON(w, 200, "Quote submitted successfully", quote)
}

// GetAvailableRFPs gets all RFPs that a vendor can submit quotes for
func (rc *QuoteController) GetAvailableRFPs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Check authentication and vendor role
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, 401, "You are not login", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "vendor" {
		respondWithJSON(w, 400, "You are not a Vendor", nil)
		return
	}

	var rfps []models.RFP

	// Get RFPs where:
	// 1. Vendor is eligible (in rfp_vendors table)
	// 2. RFP is open and not expired
	// 3. Vendor hasn't already submitted a quote
	err := database.DB.
		Joins("INNER JOIN rfp_vendors ON rfps.id = rfp_vendors.rfp_id").
		Where("rfp_vendors.vendor_id = ?", userID).
		Where("rfps.status = ? AND rfps.last_date > ? AND rfps.is_active = ?",
			models.RFPStatusOpen, time.Now(), true).
		Where("rfps.id NOT IN (?)",
			database.DB.Table("rfp_quotes").
				Select("rfp_id").
				Where("vendor_id = ?", userID)).
		Find(&rfps).Error

	if err != nil {
		respondWithJSON(w, 500, "Failed to fetch available RFPs", nil)
		return
	}

	respondWithJSON(w, 200, "Available RFPs", rfps)
}

// GetVendorRFPs gets all RFPs associated with a vendor (both available and quoted)
func (rc *QuoteController) GetVendorRFPs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Check authentication and vendor role
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		respondWithJSON(w, 401, "You are not login", nil)
		return
	}

	userRole, ok := middleware.GetUserRoleFromContext(r)
	if !ok || userRole != "vendor" {
		respondWithJSON(w, 400, "You are not a Vendor", nil)
		return
	}

	// Get query parameter to filter status
	status := r.URL.Query().Get("status") // available, quoted, all

	var rfps []models.RFP
	var err error

	switch status {
	case "available":
		// Only RFPs available for quoting
		err = database.DB.
			Joins("INNER JOIN rfp_vendors ON rfps.id = rfp_vendors.rfp_id").
			Where("rfp_vendors.vendor_id = ?", userID).
			Where("rfps.status = ? AND rfps.last_date > ? AND rfps.is_active = ?",
				models.RFPStatusOpen, time.Now(), true).
			Where("rfps.id NOT IN (?)",
				database.DB.Table("rfp_quotes").
					Select("rfp_id").
					Where("vendor_id = ?", userID)).
			Find(&rfps).Error

	case "quoted":
		// Only RFPs where vendor has submitted quotes
		err = database.DB.
			Joins("INNER JOIN rfp_quotes ON rfps.id = rfp_quotes.rfp_id").
			Where("rfp_quotes.vendor_id = ?", userID).
			Preload("Quotes", "vendor_id = ?", userID).
			Find(&rfps).Error

	default: // "all" or no parameter
		// All RFPs associated with vendor
		err = database.DB.
			Joins("INNER JOIN rfp_vendors ON rfps.id = rfp_vendors.rfp_id").
			Where("rfp_vendors.vendor_id = ?", userID).
			Preload("Quotes", "vendor_id = ?", userID).
			Find(&rfps).Error
	}

	if err != nil {
		respondWithJSON(w, 500, "Failed to fetch RFPs", nil)
		return
	}

	// For default case, add quote status information
	if status == "" || status == "all" {
		response := make([]map[string]interface{}, len(rfps))
		for i, rfp := range rfps {
			hasQuoted := len(rfp.Quotes) > 0
			canQuote := rfp.IsOpen() && !hasQuoted

			response[i] = map[string]interface{}{
				"rfp":        rfp,
				"has_quoted": hasQuoted,
				"can_quote":  canQuote,
				"is_expired": rfp.IsExpired(),
				"quote_status": func() string {
					if hasQuoted {
						return "applied"
					} else if canQuote {
						return "open"
					} else {
						return "closed"
					}
				}(),
			}
		}
		respondWithJSON(w, 200, "Vendor RFPs", response)
		return
	}

	respondWithJSON(w, 200, "Vendor RFPs", rfps)
}
