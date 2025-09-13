package services

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/karan-bishtt/rfp-quote-service/config"
)

type NotificationService struct {
	baseURL string
	client  *http.Client
}

type EmailRequest struct {
	EmailTo string `json:"email_to"`
	Subject string `json:"subject"`
	Content string `json:"content"`
	SendNow bool   `json:"send_now"`
}

func NewNotificationService() *NotificationService {
	cfg := config.Load()
	return &NotificationService{
		baseURL: cfg.NotificationServiceURL,
		client:  &http.Client{},
	}
}

// SendEmail sends an email via notification service
func (ns *NotificationService) SendEmail(to, subject, content string) error {
	emailReq := EmailRequest{
		EmailTo: to,
		Subject: subject,
		Content: content,
		SendNow: true,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return err
	}

	url := ns.baseURL + "/api/v1/send-email"
	resp, err := ns.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil // Ignore response for fire-and-forget
}

// SendBulkEmails sends emails to multiple recipients
func (ns *NotificationService) SendBulkEmails(emails []string, subject, content string) {
	for _, email := range emails {
		go ns.SendEmail(email, subject, content)
	}
}
