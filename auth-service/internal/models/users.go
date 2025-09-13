package models

import (
	"time"

	"github.com/karan-bishtt/auth-service/internal/utils"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleVendor Role = "vendor"
)

// User table
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FirstName string    `json:"first_name" gorm:"not null;size:100"`
	LastName  string    `json:"last_name" gorm:"not null;size:100"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Password  string    `json:"-" gorm:"not null"`
	Role      Role      `json:"role" gorm:"not null;type:varchar(20);default:'vendor';check:role IN ('admin','vendor')"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Vendor specific details - only populated if role is vendor
	// JSON handling: With omitempty, nil pointers are excluded from JSON output
	VendorDetails *VendorDetails `json:"vendor_details,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	// User permissions
	UserPermissions []UserPermission `json:"permissions,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook - hash password before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword
	}
	return nil
}

// BeforeUpdate hook - hash password before updating user if password is being changed
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Check if password field is being updated
	if tx.Statement.Changed("password") || tx.Statement.Changed("Password") {
		if u.Password != "" {
			hashedPassword, err := utils.HashPassword(u.Password)
			if err != nil {
				return err
			}
			u.Password = hashedPassword
		}
	}
	return nil
}

// Vendor Details table
type VendorDetails struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	Revenue      float64   `json:"revenue" gorm:"type:decimal(15,2)"`
	NoOfEmployee int       `json:"no_of_employee"`
	GSTNo        string    `json:"gst_no" gorm:"size:25"`
	PANNo        string    `json:"pan_no" gorm:"size:10"`
	PhoneNo      string    `json:"phone_no" gorm:"size:15"`
	CategoryID   *uint     `json:"category_id"` // Foreign key to category service
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Add approval fields here
	IsApproved    bool       `json:"is_approved" gorm:"default:false"`
	ApprovedBy    *uint      `json:"approved_by"`
	ApprovedAt    *time.Time `json:"approved_at"`
	ApprovalNotes string     `json:"approval_notes" gorm:"type:text"`
}

type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Description string    `json:"description" gorm:"size:255"`
	Resource    string    `json:"resource" gorm:"not null;size:50"` // e.g., "rfp", "quote", "user"
	Action      string    `json:"action" gorm:"not null;size:50"`   // e.g., "create", "read", "update", "delete"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User Permissions table
type UserPermission struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	UserID       uint        `json:"user_id" gorm:"not null"`
	PermissionID uint        `json:"permission_id" gorm:"not null"`
	User         *User       `json:"-" gorm:"foreignKey:UserID"`
	Permission   *Permission `json:"permission" gorm:"foreignKey:PermissionID"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// Table Names
func (User) TableName() string {
	return "users"
}

func (VendorDetails) TableName() string {
	return "vendor_details"
}

func (Permission) TableName() string {
	return "permissions"
}

func (UserPermission) TableName() string {
	return "user_permissions"
}

// Helper Methods
func (u *User) IsVendor() bool {
	return u.Role == RoleVendor
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) HasPermission(resource, action string) bool {
	for _, up := range u.UserPermissions {
		if up.Permission != nil &&
			up.Permission.Resource == resource &&
			up.Permission.Action == action {
			return true
		}
	}
	return false
}
