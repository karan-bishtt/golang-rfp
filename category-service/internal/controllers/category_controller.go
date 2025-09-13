package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/karan-bishtt/category-service/internal/database"
	"github.com/karan-bishtt/category-service/internal/models"
	"github.com/karan-bishtt/category-service/internal/utils"

	"github.com/gorilla/mux"
)

// ❌ Without controllers - scattered functions
// ✅ With controllers - organized by domain
// Controllers can hold shared dependencies

/**
type CategoryController struct {
    db     *gorm.DB           // Database connection
    logger *log.Logger        // Logger instance
    cache  *redis.Client      // Cache client
    config *config.Config     // Configuration
}

func NewCategoryController(db *gorm.DB, logger *log.Logger, cache *redis.Client) *CategoryController {
    return &CategoryController{
        db:     db,
        logger: logger,
        cache:  cache,
    }
}
*/

type CategoryController struct{}

type CategoryResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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

type AddCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	IsActive *bool   `json:"is_active" gorm:"default:true"`
}

func NewCategoryController() *CategoryController {
	return &CategoryController{}
}

// region json response

func respondWithJSON(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := CategoryResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	w.WriteHeader(http.StatusOK) // Always return 200, actual status in JSON
	json.NewEncoder(w).Encode(response)
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

// endregion json response

// CreateCategory creates a new category
func (cc *CategoryController) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req AddCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Data is not in correct format", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Trim and check for empty name
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondWithJSON(w, http.StatusBadRequest, "Category name cannot be empty", nil)
		return
	}

	// Check if category already exists
	var existingCategory models.Category
	if err := database.DB.Where("LOWER(name) = LOWER(?)", req.Name).First(&existingCategory).Error; err == nil {
		respondWithJSON(w, 404, "Category already exists", nil)
		return
	}

	// Create new category
	category := models.Category{
		Name:     req.Name,
		IsActive: true,
	}

	// Start transaction
	tx := database.DB.Begin()
	if err := tx.Create(&category).Error; err != nil {
		tx.Rollback()
		respondWithJSON(w, http.StatusInternalServerError, "Failed to create category", nil)
		return
	}
	tx.Commit()

	respondWithJSON(w, 200, "Category created successfully", map[string]interface{}{
		"id":        category.ID,
		"name":      category.Name,
		"is_active": category.IsActive,
	})
}

// GetCategories retrieves categories with pagination
func (cc *CategoryController) GetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status") // active, inactive, all

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

	// Build query
	query := database.DB.Model(&models.Category{})

	// Filter by status
	switch strings.ToLower(status) {
	case "active":
		query = query.Where("is_active = ?", true)
	case "inactive":
		query = query.Where("is_active = ?", false)
	default:
		// Show all by default (don't add where clause)
	}

	// Search filter
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Calculate pagination
	offset := (page - 1) * limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Get categories
	var categories []models.Category
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&categories).Error; err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to retrieve categories", nil)
		return
	}

	pagination := Pagination{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  totalPages,
	}

	respondWithPagination(w, 200, "Categories retrieved successfully", categories, pagination)
}

// GetCategory retrieves a single category by ID
func (cc *CategoryController) GetCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}

	var category models.Category
	if err := database.DB.First(&category, uint(id)).Error; err != nil {
		respondWithJSON(w, 404, "Category not found", nil)
		return
	}

	respondWithJSON(w, 200, "Category retrieved successfully", category)
}

// UpdateCategory updates an existing category
func (cc *CategoryController) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Data is not in correct format", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Find existing category
	var category models.Category
	if err := database.DB.First(&category, uint(id)).Error; err != nil {
		respondWithJSON(w, 404, "Category not found", nil)
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Update fields if provided
	updates := make(map[string]interface{})

	if req.Name != nil {
		newName := strings.TrimSpace(*req.Name)
		if newName == "" {
			tx.Rollback()
			respondWithJSON(w, http.StatusBadRequest, "Category name cannot be empty", nil)
			return
		}

		// Check if another category with this name exists
		var existingCategory models.Category
		if err := tx.Where("LOWER(name) = LOWER(?) AND id != ?", newName, id).First(&existingCategory).Error; err == nil {
			tx.Rollback()
			respondWithJSON(w, 409, "Another category with this name already exists", nil)
			return
		}

		updates["name"] = newName
		updates["is_active"] = *req.IsActive
	}

	// Perform update
	if len(updates) > 0 {
		if err := tx.Model(&category).Updates(updates).Error; err != nil {
			tx.Rollback()
			respondWithJSON(w, http.StatusInternalServerError, "Failed to update category", nil)
			return
		}
	}

	tx.Commit()

	// Fetch updated category
	database.DB.First(&category, uint(id))

	respondWithJSON(w, 200, "Category updated successfully", category)
}

// DeleteCategory soft deletes a category
func (cc *CategoryController) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}

	// Find existing category
	var category models.Category
	if err := database.DB.First(&category, uint(id)).Error; err != nil {
		respondWithJSON(w, 404, "Category not found", nil)
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Soft delete the category
	if err := tx.Delete(&category).Error; err != nil {
		tx.Rollback()
		respondWithJSON(w, http.StatusInternalServerError, "Failed to delete category", nil)
		return
	}

	tx.Commit()

	respondWithJSON(w, 200, "Category deleted successfully", nil)
}

// ToggleCategoryStatus toggles the active status of a category
func (cc *CategoryController) ToggleCategoryStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}

	// Find existing category
	var category models.Category
	if err := database.DB.First(&category, uint(id)).Error; err != nil {
		respondWithJSON(w, 404, "Category not found", nil)
		return
	}

	// Toggle status
	newStatus := !category.IsActive
	if err := database.DB.Model(&category).Update("is_active", newStatus).Error; err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to update category status", nil)
		return
	}

	category.IsActive = newStatus
	respondWithJSON(w, 200, "Category status updated successfully", category)
}
