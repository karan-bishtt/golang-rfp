package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/karan-bishtt/user-service/internal/database"
	"github.com/karan-bishtt/user-service/internal/models"
	"github.com/karan-bishtt/user-service/internal/services"

	"github.com/gorilla/mux"
)

type UserController struct {
	authService *services.AuthService
}

type UserResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ApprovalRequest struct {
	IsApproved bool   `json:"is_approved"`
	Notes      string `json:"notes,omitempty"`
}

func NewUserController() *UserController {
	return &UserController{
		authService: services.NewAuthService(),
	}
}

func respondWithJSON(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := UserResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetVendors retrieves all vendors with approval status
func (uc *UserController) GetVendors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Fetch vendors from auth-service
	vendors, err := uc.authService.GetVendors()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to fetch vendors", nil)
		return
	}

	// Get approval status for each vendor
	var vendorApprovals []models.VendorApproval
	database.DB.Find(&vendorApprovals)

	// Create approval map for quick lookup
	approvalMap := make(map[uint]models.VendorApproval)
	for _, approval := range vendorApprovals {
		approvalMap[approval.UserID] = approval
	}

	// Combine vendor data with approval status
	type VendorWithApproval struct {
		models.UserResponse
		ApprovalStatus *models.VendorApproval `json:"approval_status,omitempty"`
	}

	var result []VendorWithApproval
	for _, vendor := range vendors {
		vendorWithApproval := VendorWithApproval{
			UserResponse: vendor,
		}

		if approval, exists := approvalMap[vendor.ID]; exists {
			vendorWithApproval.ApprovalStatus = &approval
		}

		result = append(result, vendorWithApproval)
	}

	respondWithJSON(w, 200, "Vendors retrieved successfully", result)
}

// GetVendor retrieves a specific vendor with approval status
func (uc *UserController) GetVendor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Fetch user from auth-service
	user, err := uc.authService.GetUserByID(uint(userID))
	if err != nil {
		respondWithJSON(w, 404, "User not found", nil)
		return
	}

	// Get approval status
	var approval models.VendorApproval
	database.DB.Where("user_id = ?", userID).First(&approval)

	result := map[string]interface{}{
		"user":            user,
		"approval_status": approval,
	}

	respondWithJSON(w, 200, "Vendor retrieved successfully", result)
}

// ApproveVendor handles vendor approval/disapproval
func (uc *UserController) ApproveVendor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid request format", nil)
		return
	}

	// Verify user exists in auth-service
	user, err := uc.authService.GetUserByID(uint(userID))
	if err != nil {
		respondWithJSON(w, 404, "User not found", nil)
		return
	}

	if user.Role != "vendor" {
		respondWithJSON(w, http.StatusBadRequest, "User is not a vendor", nil)
		return
	}

	// Get current admin user ID from context (from JWT middleware)
	// adminUserID := middleware.GetUserIDFromContext(r) // Implement this

	// Create or update approval record
	var approval models.VendorApproval
	result := database.DB.Where("user_id = ?", userID).First(&approval)

	now := time.Now()
	if result.Error != nil {
		// Create new approval record
		approval = models.VendorApproval{
			UserID:     uint(userID),
			IsApproved: req.IsApproved,
			// ApprovedBy: &adminUserID,
			ApprovedAt: &now,
			Notes:      req.Notes,
		}
		database.DB.Create(&approval)
	} else {
		// Update existing approval record
		approval.IsApproved = req.IsApproved
		// approval.ApprovedBy = &adminUserID
		approval.ApprovedAt = &now
		approval.Notes = req.Notes
		database.DB.Save(&approval)
	}

	action := "disapproved"
	if req.IsApproved {
		action = "approved"
	}

	respondWithJSON(w, 200, fmt.Sprintf("Vendor %s successfully", action), approval)
}
