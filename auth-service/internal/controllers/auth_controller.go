package controllers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/karan-bishtt/auth-service/internal/database"
	"github.com/karan-bishtt/auth-service/internal/models"
	"github.com/karan-bishtt/auth-service/internal/services"
	"github.com/karan-bishtt/auth-service/internal/utils"
	"gorm.io/gorm"
)

// region validators
type AuthController struct {
	notificationService *services.NotificationService
}

// Add these request structs
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTP         string `json:"otp" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type RegisterVendorRequest struct {
	FirstName     string  `json:"firstname" validate:"required"`
	LastName      string  `json:"lastname" validate:"required"`
	Email         string  `json:"email" validate:"required,email"`
	Password      string  `json:"password" validate:"required,min=8"`
	Revenue       float64 `json:"revenue"`
	EmployeeCount int     `json:"no_of_employees"`
	GSTNo         string  `json:"gst_no"`
	PANNo         string  `json:"pancard_no"`
	PhoneNo       string  `json:"mobile"`
	CategoryID    uint    `json:"category"`
}

type RegisterAdminRequest struct {
	FirstName string `json:"firstname" validate:"required"`
	LastName  string `json:"lastname" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Refresh string      `json:"refresh,omitempty"`
	Access  string      `json:"access,omitempty"`
	Role    string      `json:"role,omitempty"`
	Name    string      `json:"name,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Add this struct with other request structs
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Pagination struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Status     int         `json:"status"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// endregion validators

// region helpers
func NewAuthController() *AuthController {
	return &AuthController{
		notificationService: services.NewNotificationService(),
	}
}

func respondWithJSON(w http.ResponseWriter, status int, message, role, refresh, access, name string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := AuthResponse{
		Status:  status,
		Message: message,
		Role:    role,
		Refresh: refresh,
		Access:  access,
		Name:    name,
		Data:    data,
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func respondWithPagination(w http.ResponseWriter, status int, message string, data interface{}, pagination Pagination) {
	w.Header().Set("Content-Type", "application/json")
	response := PaginatedResponse{
		Status:     status,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func assignDefaultVendorPermissions(tx *gorm.DB, userID uint) error {
	vendorPermissions := []string{
		"read_rfp", "create_quote", "read_quote", "update_quote",
	}

	for _, permName := range vendorPermissions {
		var permission models.Permission
		if err := tx.Where("name = ?", permName).First(&permission).Error; err != nil {
			continue // Skip if permission doesn't exist
		}

		userPermission := models.UserPermission{
			UserID:       userID,
			PermissionID: permission.ID,
		}
		tx.Create(&userPermission)
	}
	return nil
}

func assignDefaultAdminPermissions(tx *gorm.DB, userID uint) error {
	adminPermissions := []string{
		"create_rfp", "read_rfp", "update_rfp", "delete_rfp",
		"read_quote", "manage_users", "manage_categories",
	}

	for _, permName := range adminPermissions {
		var permission models.Permission
		if err := tx.Where("name = ?", permName).First(&permission).Error; err != nil {
			continue // Skip if permission doesn't exist
		}

		userPermission := models.UserPermission{
			UserID:       userID,
			PermissionID: permission.ID,
		}
		tx.Create(&userPermission)
	}
	return nil
}

// Helper function to generate 6-digit OTP
func generateOTP() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%06d", int(b[0])<<16|int(b[1])<<8|int(b[2])%1000000)
}

// endregion helpers

// Register Vendor
func (ac *AuthController) RegisterVendor(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondWithJSON(w, 200, "SUCCESSFULL", "", "", "", "", []interface{}{})
		return

	case http.MethodPost:
		var req RegisterVendorRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithJSON(w, 400, "Data is not in correct format", "", "", "", "", nil)
			return
		}
		// validate request
		if err := utils.ValidateStruct(req); err != nil {
			respondWithJSON(w, 400, err.Error(), "", "", "", "", nil)
			return
		}

		// Check if email already exists
		var existingUser models.User
		if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			respondWithJSON(w, 404, "Email is already present", "vendor", "", "", "", nil)
			return
		}

		// Create User
		user := models.User{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Password:  req.Password,
			Role:      models.RoleVendor,
			IsActive:  true,
		}

		// Start transaction
		tx := database.DB.Begin()
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			respondWithJSON(w, 500, "Failed to create user", "", "", "", "", nil)
			return
		}

		// Create vendor details (adjust field names to your actual model if needed)
		vendorDetails := models.VendorDetails{
			UserID:       user.ID,
			Revenue:      req.Revenue,       // NOTE: keep as-is per your original; rename to match your model if needed
			NoOfEmployee: req.EmployeeCount, // NOTE: keep as-is per your original
			GSTNo:        req.GSTNo,
			PANNo:        req.PANNo,
			PhoneNo:      req.PhoneNo,
			// CategoryID:   &req.CategoryID, // NOTE: keep as-is per your original
			CategoryID: &req.CategoryID, // NOTE: keep as-is per your original
		}

		if err := tx.Create(&vendorDetails).Error; err != nil {
			tx.Rollback()
			respondWithJSON(w, 500, "Failed to create vendor details", "", "", "", "", nil)
			return
		}

		tx.Commit()

		// Generate JWT token
		refresh, access, err := utils.GenerateTokenPair(user.ID, string(user.Role))
		if err != nil {
			respondWithJSON(w, 500, "Failed to generate tokens", "", "", "", "", nil)
			return
		}

		// Send notification emails
		email := req.Email
		fullName := req.FirstName + " " + req.LastName
		subject := "Registered Vendor"
		content := fmt.Sprintf(`
			Hi %s,

			You have been successfully registered. Please wait until your request is 
			approved by admin.
		`, fullName)
		ac.notificationService.SendEmail(email, subject, content)
		respondWithJSON(w, 200, "Registration for vendor is successfull", "vendor", refresh, access, fullName, nil)
	}
}

// Register Admin
func (ac *AuthController) RegisterAdmin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	var req RegisterAdminRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Data is not in correct format", "", "", "", "", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, err.Error(), "", "", "", "", nil)
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		respondWithJSON(w, 404, "Email is already present", string(models.RoleAdmin), "", "", "", nil)
		return
	}

	// Create admin user
	user := models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password, // Will be hashed by BeforeCreate hook
		Role:      models.RoleAdmin,
		IsActive:  true,
	}

	// Start transaction
	tx := database.DB.Begin()

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		respondWithJSON(w, http.StatusInternalServerError, "Failed to create user", "", "", "", "", nil)
		return
	}

	// Assign Default admin permission
	if err := assignDefaultAdminPermissions(tx, user.ID); err != nil {
		tx.Rollback()
		respondWithJSON(w, 500, "Failed to assign permission", "", "", "", "", nil)
		return
	}

	tx.Commit()

	// generate token
	refresh, access, err := utils.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		respondWithJSON(w, 500, "Failed to generate tokens", "", "", "", "", nil)
		return
	}

	// Send notification emails
	email := req.Email
	fullName := req.FirstName + " " + req.LastName
	subject := "Registered Admin"
	content := fmt.Sprintf(`
		Hi %s,

		You have been successfully registered.
	`, fullName)
	ac.notificationService.SendEmail(email, subject, content)
	respondWithJSON(w, 200, "Registration for Admin is Successful", string(models.RoleAdmin), refresh, access, fullName, nil)
}

// Login handles user authentication
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, 400, "Invalid request format", "", "", "", "", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, 400, err.Error(), "", "", "", "", nil)
		return
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).Preload("VendorDetails").First(&user).Error; err != nil {
		respondWithJSON(w, 400, "Email does not exist", "", "", "", "", nil)
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		respondWithJSON(w, 400, "Invalid password", "", "", "", "", nil)
		return
	}

	// Check if user is active
	if !user.IsActive {
		respondWithJSON(w, 400, "Account is deactivated", "", "", "", "", nil)
		return
	}

	if user.Role == models.RoleVendor && !user.VendorDetails.IsApproved {
		respondWithJSON(w, 400, "Account is not approved by admin", "", "", "", "", nil)
		return
	}

	// Generate JWT tokens
	refresh, access, err := utils.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		respondWithJSON(w, 500, "Failed to generate tokens", "", "", "", "", nil)
		return
	}

	fullName := user.FirstName + " " + user.LastName
	respondWithJSON(w, 200, "Login successful", string(user.Role), refresh, access, fullName, map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

// RefreshToken handles token refresh
func (ac *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, 400, "Invalid request format", "", "", "", "", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, 400, err.Error(), "", "", "", "", nil)
		return
	}

	// Generate new access token
	newAccessToken, err := utils.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		respondWithJSON(w, 401, "Invalid refresh token", "", "", "", "", nil)
		return
	}

	respondWithJSON(w, 200, "Token refreshed successfully", "", "", newAccessToken, "", nil)
}

// GetVendors - get all vendors with filtering
func (ac *AuthController) GetVendors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	// Query parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	status := r.URL.Query().Get("status") // pending, approved, rejected

	// Set defaults
	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	query := database.DB.Where("role = ?", models.RoleVendor).
		Preload("VendorDetails").
		Preload("UserPermissions.Permission")

	// Filter by approval status
	switch status {
	case "pending":
		query = query.Joins("LEFT JOIN vendor_details ON users.id = vendor_details.user_id").
			Where("vendor_details.is_approved = ? OR vendor_details.is_approved IS NULL", false)
	case "approved":
		query = query.Joins("JOIN vendor_details ON users.id = vendor_details.user_id").
			Where("vendor_details.is_approved = ?", true)
	}

	// Get total count
	var total int64
	var users []models.User
	query.Model(&users).Count(&total)
	// Calculate pagination
	offset := (page - 1) * limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		respondWithJSON(w, 500, "Failed to fetch vendors", "", "", "", "", nil)
		return
	}

	pagination := Pagination{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  totalPages,
	}

	respondWithPagination(w, 200, "Categories retrieved successfully", users, pagination)
}

// GetVendors - get all vendors with filtering
func (ac *AuthController) GetVendorsByCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	// Query parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	status := r.URL.Query().Get("status") // pending, approved, rejected

	// Set defaults
	page := 1
	limit := 10

	vars := mux.Vars(r)
	categoryId := vars["id"] // Ensure the 'id' is present in the route

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Build the query with correct joins and where clauses
	query := database.DB.Joins("JOIN vendor_details ON users.id = vendor_details.user_id").
		Where("users.role = ? AND vendor_details.category_id = ?", models.RoleVendor, categoryId)

	// Filter by approval status
	switch status {
	case "pending":
		query = query.Where("vendor_details.is_approved = ? OR vendor_details.is_approved IS NULL", false)
	case "approved":
		query = query.Where("vendor_details.is_approved = ?", true)
	}

	// Get total count with error handling
	var total int64
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		log.Printf("Error counting vendors: %v", err)
		respondWithJSON(w, 500, "Failed to fetch vendors", "", "", "", "", nil)
		return
	}

	// Calculate pagination
	offset := (page - 1) * limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Fetch vendors with pagination
	var users []models.User
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		log.Printf("Error fetching vendors: %v", err)
		respondWithJSON(w, 500, "Failed to fetch vendors", "", "", "", "", nil)
		return
	}

	// Prepare pagination response
	pagination := Pagination{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  totalPages,
	}

	// Return vendors with pagination info
	respondWithPagination(w, 200, "Vendors retrieved successfully", users, pagination)
}

// GetVendors - get all vendors with filtering
func (ac *AuthController) GetVendorById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	// Get vendorId from query params (or you can use mux.Vars if path param)
	vars := mux.Vars(r)
	vendorIDStr := vars["id"]
	if vendorIDStr == "" {
		respondWithJSON(w, 400, "vendorId is required", "", "", "", "", nil)
		return
	}

	vendorID, err := strconv.Atoi(vendorIDStr)
	if err != nil || vendorID <= 0 {
		respondWithJSON(w, 400, "Invalid vendorId", "", "", "", "", nil)
		return
	}

	var user models.User
	query := database.DB.Where("id = ? AND role = ?", vendorID, models.RoleVendor).
		Preload("VendorDetails").
		Preload("UserPermissions.Permission")

	// Execute query
	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithJSON(w, 404, "Vendor not found", "", "", "", "", nil)
		} else {
			respondWithJSON(w, 500, "Failed to fetch vendor", "", "", "", "", nil)
		}
		return
	}

	respondWithJSON(w, 200, "Vendor retrieved successfully", "", "", "", "", user)
}

// ApproveVendor - approve/reject a vendor
func (ac *AuthController) ApproveVendor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	var req struct {
		IsApproved bool   `json:"is_approved"`
		VendorID   int    `json:"vendor_id" validate:"required"`
		Notes      string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, 400, "Invalid request format", "", "", "", "", nil)
		return
	}
	vendorID := req.VendorID
	// Find vendor
	var user models.User
	if err := database.DB.Where("id = ? AND role = ?", uint(vendorID), models.RoleVendor).
		Preload("VendorDetails").
		First(&user).Error; err != nil {
		respondWithJSON(w, 404, "Vendor not found", "", "", "", "", nil)
		return
	}

	if user.VendorDetails == nil {
		respondWithJSON(w, 400, "Vendor details not found", "", "", "", "", nil)
		return
	}

	// Update approval status
	now := time.Now()
	updates := map[string]interface{}{
		"is_approved":    req.IsApproved,
		"approved_at":    &now,
		"approval_notes": req.Notes,
	}

	if err := database.DB.Model(user.VendorDetails).Updates(updates).Error; err != nil {
		respondWithJSON(w, 500, "Failed to update approval status", "", "", "", "", nil)
		return
	}

	// Fetch updated user
	database.DB.Where("id = ?", uint(vendorID)).Preload("VendorDetails").First(&user)

	action := "rejected"
	subject := "Rejected Vendor"
	if req.IsApproved {
		action = "approved"
		subject = "Approved Vendor"
	}

	// Send notification emails
	email := user.Email
	fullName := user.FirstName + " " + user.LastName
	content := fmt.Sprintf(`
		Hi %s,

		Your request to join the RFP system has been %s.
	`, fullName, action)
	ac.notificationService.SendEmail(email, subject, content)

	respondWithJSON(w, 200, fmt.Sprintf("Vendor %s successfully", action), "", "", "", "", user)
}

// ForgotPassword - sends OTP to user's email
func (ac *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, 400, "Invalid request format", "", "", "", "", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, 400, err.Error(), "", "", "", "", nil)
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		respondWithJSON(w, 200, "If the email exists, an OTP has been sent", "", "", "", "", nil)
		return
	}

	// Delete any existing OTP for this email
	database.DB.Where("email = ?", req.Email).Delete(&models.PasswordResetOTP{})

	// Generate new OTP
	otp := generateOTP()

	// Save OTP to database
	resetOTP := models.PasswordResetOTP{
		Email:     req.Email,
		OTP:       otp,
		Attempts:  0,
		ExpiresAt: time.Now().Add(15 * time.Minute), // OTP expires in 15 minutes
	}

	if err := database.DB.Create(&resetOTP).Error; err != nil {
		respondWithJSON(w, 500, "Failed to process request", "", "", "", "", nil)
		return
	}

	// Send OTP email
	subject := "Password Reset OTP"
	content := fmt.Sprintf(`
		Hi %s,

		You have requested to reset your password. Your OTP is:

		%s

		This OTP will expire in 15 minutes. Do not share this OTP with anyone.

		If you did not request this, please ignore this email.
	`, user.FirstName, otp)

	go ac.notificationService.SendEmail(req.Email, subject, content)

	respondWithJSON(w, 200, "If the email exists, an OTP has been sent", "", "", "", "", nil)
}

// ResetPassword - verifies OTP and updates password
func (ac *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, 405, "Method not allowed", "", "", "", "", nil)
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, 400, "Invalid request format", "", "", "", "", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, 400, err.Error(), "", "", "", "", nil)
		return
	}

	// Find OTP record
	var resetOTP models.PasswordResetOTP
	if err := database.DB.Where("email = ?", req.Email).First(&resetOTP).Error; err != nil {
		respondWithJSON(w, 400, "Invalid or expired OTP", "", "", "", "", nil)
		return
	}

	// Check if OTP has expired
	if time.Now().After(resetOTP.ExpiresAt) {
		database.DB.Delete(&resetOTP)
		respondWithJSON(w, 400, "OTP has expired", "", "", "", "", nil)
		return
	}

	// Check if attempts exceeded
	if resetOTP.Attempts >= 3 {
		database.DB.Delete(&resetOTP)
		respondWithJSON(w, 400, "Too many failed attempts. Please request a new OTP", "", "", "", "", nil)
		return
	}

	// Verify OTP
	if resetOTP.OTP != req.OTP {
		// Increment attempts
		resetOTP.Attempts++
		if resetOTP.Attempts >= 3 {
			database.DB.Delete(&resetOTP)
			respondWithJSON(w, 400, "Too many failed attempts. Please request a new OTP", "", "", "", "", nil)
		} else {
			database.DB.Save(&resetOTP)
			remainingAttempts := 3 - resetOTP.Attempts
			respondWithJSON(w, 400, fmt.Sprintf("Invalid OTP. %d attempts remaining", remainingAttempts), "", "", "", "", nil)
		}
		return
	}

	// OTP is valid, update password
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		respondWithJSON(w, 400, "User not found", "", "", "", "", nil)
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		respondWithJSON(w, 500, "Failed to process password", "", "", "", "", nil)
		return
	}

	// Update password
	if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		respondWithJSON(w, 500, "Failed to update password", "", "", "", "", nil)
		return
	}

	// Delete OTP record
	database.DB.Delete(&resetOTP)

	// Send confirmation email
	subject := "Password Reset Successful"
	content := fmt.Sprintf(`
		Hi %s,

		Your password has been successfully reset. You can now login with your new password.

		If you did not make this change, please contact support immediately.
	`, user.FirstName)

	go ac.notificationService.SendEmail(req.Email, subject, content)

	respondWithJSON(w, 200, "Password reset successful", "", "", "", "", nil)
}
