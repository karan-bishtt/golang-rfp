package models

import "time"

// DTOs for receiving data from auth-service
type UserResponse struct {
	ID            uint                   `json:"id"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Email         string                 `json:"email"`
	Role          string                 `json:"role"`
	IsActive      bool                   `json:"is_active"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	VendorDetails *VendorDetailsResponse `json:"vendor_details,omitempty"`
}

type VendorDetailsResponse struct {
	ID           uint    `json:"id"`
	UserID       uint    `json:"user_id"`
	Revenue      float64 `json:"revenue"`
	NoOfEmployee int     `json:"no_of_employee"`
	GSTNo        string  `json:"gst_no"`
	PANNo        string  `json:"pan_no"`
	PhoneNo      string  `json:"phone_no"`
	CategoryID   *uint   `json:"category_id"`
	IsApproved   bool    `json:"is_approved"` // This could be managed in user-service
}

// Local model for vendor approval status
type VendorApproval struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"uniqueIndex;not null"`
	IsApproved bool       `json:"is_approved" gorm:"default:false"`
	ApprovedBy *uint      `json:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at"`
	Notes      string     `json:"notes" gorm:"type:text"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (VendorApproval) TableName() string {
	return "vendor_approvals"
}
