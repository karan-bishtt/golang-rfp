package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/karan-bishtt/notification-service/internal/database"
	"github.com/karan-bishtt/notification-service/internal/models"
	"github.com/karan-bishtt/notification-service/internal/utils"
)

type NotificationController struct{}

type NotificationResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SendEmailRequest struct {
	To      string `json:"email_to" validate:"required,email"`
	Subject string `json:"subject,omitempty"`
	Content string `json:"content" validate:"required"`
	SendNow bool   `json:"send_now,omitempty"` // true = send immediately, false = queue
}

type SendSMSRequest struct {
	To      string `json:"phone_to" validate:"required"`
	Content string `json:"content" validate:"required"`
	SendNow bool   `json:"send_now,omitempty"`
}

// ROUTING - HELPER
func NewNotificationController() *NotificationController {
	return &NotificationController{}
}

func respondWithJSON(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := NotificationResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendEmail handles email sending requests
func (nc *NotificationController) SendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid request format", nil)
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Set default subject if empty
	if req.Subject == "" {
		req.Subject = "Notification"
	}

	// Create notification record
	notification := models.Notification{
		Type:       models.TypeEmail,
		To:         req.To,
		Subject:    req.Subject,
		Content:    req.Content,
		Status:     models.StatusPending,
		MaxRetries: 3,
	}

	// Save to database
	if err := database.DB.Create(&notification).Error; err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to create notification", nil)
		return
	}

	// Always send in background (remove SendNow option)
	go nc.processEmailNotificationAsync(notification.ID)

	respondWithJSON(w, 200, "Email queued successfully", map[string]interface{}{
		"notification_id": notification.ID,
		"status":          "queued",
		"message":         "Email will be sent shortly",
	})

	// Send immediately if requested
	// if req.SendNow {
	// 	success := nc.processEmailNotification(&notification)
	// 	if success {
	// 		respondWithJSON(w, 200, "Email sent successfully", map[string]interface{}{
	// 			"notification_id": notification.ID,
	// 			"status":          notification.Status,
	// 			"sent_at":         notification.SentAt,
	// 		})
	// 	} else {
	// 		respondWithJSON(w, 500, "Failed to send email", map[string]interface{}{
	// 			"notification_id": notification.ID,
	// 			"status":          notification.Status,
	// 			"error":           notification.ErrorMsg,
	// 		})
	// 	}
	// 	return
	// }

	// Queued for later processing
	// respondWithJSON(w, 200, "Email queued successfully", map[string]interface{}{
	// 	"notification_id": notification.ID,
	// 	"status":          notification.Status,
	// })
}

// GetNotificationStatus retrieves notification status
func (nc *NotificationController) GetNotificationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithJSON(w, http.StatusBadRequest, "Notification ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Invalid notification ID", nil)
		return
	}

	var notification models.Notification
	if err := database.DB.First(&notification, uint(id)).Error; err != nil {
		respondWithJSON(w, 404, "Notification not found", nil)
		return
	}

	respondWithJSON(w, 200, "Notification status retrieved", notification)
}

// ProcessPendingNotifications processes queued notifications
func (nc *NotificationController) ProcessPendingNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Get pending notifications
	var notifications []models.Notification
	if err := database.DB.Where("status IN ?", []models.NotificationStatus{
		models.StatusPending, models.StatusRetry,
	}).Find(&notifications).Error; err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Failed to fetch notifications", nil)
		return
	}

	processed := 0
	successful := 0
	failed := 0

	for _, notification := range notifications {
		processed++
		var success bool

		switch notification.Type {
		case models.TypeEmail:
			success = nc.processEmailNotification(&notification)
			// case models.TypeSMS:

		}

		if success {
			successful++
		} else {
			failed++
		}
	}

	respondWithJSON(w, 200, "Batch processing completed", map[string]interface{}{
		"processed":  processed,
		"successful": successful,
		"failed":     failed,
	})
}

// processEmailNotification handles actual email sending
func (nc *NotificationController) processEmailNotification(notification *models.Notification) bool {
	// Send email using email utility
	err := utils.SendEmail(notification.To, notification.Subject, notification.Content)

	now := time.Now()
	if err != nil {
		notification.Status = models.StatusFailed
		notification.ErrorMsg = err.Error()
		notification.IncrementRetry()
	} else {
		notification.Status = models.StatusSent
		notification.SentAt = &now
		notification.ErrorMsg = ""
	}

	// Update notification in database
	database.DB.Save(notification)

	return err == nil
}

// Async processing with database safety
func (nc *NotificationController) processEmailNotificationAsync(notificationID uint) {
	// Fetch fresh notification from database to avoid stale data
	var notification models.Notification
	if err := database.DB.First(&notification, notificationID).Error; err != nil {
		log.Printf("Failed to fetch notification %d: %v", notificationID, err)
		return
	}

	// Process the email
	nc.processEmailNotification(&notification)
}
