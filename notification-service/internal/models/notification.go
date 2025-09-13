package models

import (
	"time"
)

type NotificationStatus string
type NotificationType string

const (
	StatusPending NotificationStatus = "pending"
	StatusSent    NotificationStatus = "sent"
	StatusFailed  NotificationStatus = "failed"
	StatusRetry   NotificationStatus = "retry"

	TypeEmail NotificationType = "email"
	TypeSMS   NotificationType = "sms"
)

type Notification struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	Type        NotificationType   `json:"type" gorm:"not null;type:varchar(20);default:'email';check:type IN ('email')"`
	To          string             `json:"to" gorm:"not null;size:255"` // email or phone
	Subject     string             `json:"subject" gorm:"size:255"`     // for emails
	Content     string             `json:"content" gorm:"type:text;not null"`
	Status      NotificationStatus `json:"status" gorm:"not null;type:varchar(20);default:'pending';check:status IN ('pending','sent','failed','retry')"`
	RetryCount  int                `json:"retry_count" gorm:"default:0"`
	MaxRetries  int                `json:"max_retries" gorm:"default:3"`
	ScheduledAt *time.Time         `json:"scheduled_at,omitempty"` // for future scheduling
	SentAt      *time.Time         `json:"sent_at,omitempty"`
	ErrorMsg    string             `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// Table name
func (Notification) TableName() string {
	return "notifications"
}

// Helper methods
func (n *Notification) CanRetry() bool {
	return n.RetryCount < n.MaxRetries && n.Status == StatusFailed
}

func (n *Notification) IncrementRetry() {
	n.RetryCount++
	if n.RetryCount >= n.MaxRetries {
		n.Status = StatusFailed
	} else {
		n.Status = StatusRetry
	}
}
