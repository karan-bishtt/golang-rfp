package models

import (
	"time"
)

type RFPStatus string

// Add to your models package
type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func (User) TableName() string {
	return "users" // Make sure this matches your auth-service table
}

const (
	RFPStatusOpen   RFPStatus = "open"
	RFPStatusClosed RFPStatus = "closed"
	RFPStatusDraft  RFPStatus = "draft"
)

type RFP struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	Quantity    int       `json:"quantity" gorm:"default:1"`
	LastDate    time.Time `json:"last_date" gorm:"not null"`
	MinAmount   float64   `json:"min_amount" gorm:"type:decimal(15,2)"`
	MaxAmount   float64   `json:"max_amount" gorm:"type:decimal(15,2)"`
	Status      RFPStatus `json:"status" gorm:"type:varchar(20);default:'open'"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CategoryID  *uint     `json:"category_id"`
	UserID      uint      `json:"user_id" gorm:"not null"` // Admin who created RFP
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Quotes []RFPQuote `json:"quotes,omitempty" gorm:"foreignKey:RFPID;constraint:OnDelete:CASCADE"`
	// EligibleVendors []uint     `json:"eligible_vendors,omitempty" gorm:"many2many:rfp_vendors;"`
}

type RFPVendor struct {
	RFPID     uint      `gorm:"primaryKey"`
	VendorID  uint      `gorm:"primaryKey"`
	InvitedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type RFPQuote struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	RFPID           uint      `json:"rfp_id" gorm:"not null"`
	VendorID        uint      `json:"vendor_id" gorm:"not null"`
	VendorPrice     float64   `json:"vendor_price" gorm:"type:decimal(15,2)"`
	ItemDescription string    `json:"item_description" gorm:"type:text"`
	Quantity        int       `json:"quantity"`
	TotalCost       float64   `json:"total_cost" gorm:"type:decimal(15,2)"`
	Status          string    `json:"status" gorm:"default:'pending'"` // pending, accepted, rejected
	SubmittedAt     time.Time `json:"submitted_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	RFP    *RFP  `json:"rfp,omitempty" gorm:"foreignKey:RFPID"`
	Vendor *User `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
}

// Table names
func (RFP) TableName() string {
	return "rfps"
}

func (RFPQuote) TableName() string {
	return "rfp_quotes"
}

// Helper methods
func (r *RFP) IsOpen() bool {
	return r.Status == RFPStatusOpen && r.LastDate.After(time.Now())
}

func (r *RFP) IsExpired() bool {
	return r.LastDate.Before(time.Now())
}
