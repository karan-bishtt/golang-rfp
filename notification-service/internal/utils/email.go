package utils

import (
	"fmt"
	"net/smtp"

	"github.com/karan-bishtt/notification-service/config"
)

// SendEmail sends an email using SMTP
func SendEmail(to, subject, content string) error {
	cfg := config.Load()

	// SMTP configuration
	smtpHost := cfg.SMTPHost
	smtpPort := cfg.SMTPPort
	smtpUser := cfg.SMTPUser
	smtpPass := cfg.SMTPPassword

	// Create message
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, content)

	// Authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Send email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		smtpUser,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
