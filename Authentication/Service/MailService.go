package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
)

type MailService struct {
	To       string
	From     string
	password string
	smtpHost string
	smtpPort string
}

func NewMailService() *MailService {
	return &MailService{}
}

type OTPTemplateData struct {
	OTPCode         string
	ExpiresAt       string
	GeneratedAt     string
	ValidityMinutes int
	CurrentYear     int
	Username        string
}

func (ms *MailService) SendVerificationEmail(user *models.User, verificationCode string) error {
	to := user.Email
	subject := "Blackwater Server Manager - Email Verification"

	// Prepare template data
	data := OTPTemplateData{
		OTPCode:         verificationCode,
		ExpiresAt:       time.Now().Add(15 * time.Minute).Format("15:04"),
		GeneratedAt:     time.Now().Format("15:04"),
		ValidityMinutes: 15,
		CurrentYear:     time.Now().Year(),
		Username:        user.Username,
	}

	// Render HTML template
	htmlBody, err := ms.renderHTMLTemplate("otp_verification", data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %v", err)
	}

	return ms.SendHTMLEmail(to, subject, htmlBody)
}

func (ms *MailService) SendPasswordResetEmail(user *models.User, resetCode string) error {
	to := user.Email
	subject := "Blackwater Server Manager - Password Reset"

	// Prepare template data
	data := OTPTemplateData{
		OTPCode:         resetCode,
		ExpiresAt:       time.Now().Add(30 * time.Minute).Format("15:04"),
		GeneratedAt:     time.Now().Format("15:04"),
		ValidityMinutes: 30,
		CurrentYear:     time.Now().Year(),
		Username:        user.Username,
	}

	// Render HTML template
	htmlBody, err := ms.renderHTMLTemplate("password_reset", data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %v", err)
	}

	return ms.SendHTMLEmail(to, subject, htmlBody)
}

func (ms *MailService) SendOTPEmail(user *models.User, otpCode string, purpose string) error {
	to := user.Email
	subject := fmt.Sprintf("Blackwater Server Manager - %s Verification", purpose)

	// Prepare template data
	data := OTPTemplateData{
		OTPCode:         otpCode,
		ExpiresAt:       time.Now().Add(5 * time.Minute).Format("15:04"),
		GeneratedAt:     time.Now().Format("15:04"),
		ValidityMinutes: 5,
		CurrentYear:     time.Now().Year(),
		Username:        user.Username,
	}

	// Render HTML template
	htmlBody, err := ms.renderHTMLTemplate("otp_email", data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %v", err)
	}

	return ms.SendHTMLEmail(to, subject, htmlBody)
}

func (ms *MailService) renderHTMLTemplate(templateName string, data OTPTemplateData) (string, error) {
	// Template file mapping
	templateFiles := map[string]string{
		"otp_email":        "templates/otp_email.html",
		"otp_verification": "templates/email_verification.html",
		"password_reset":   "templates/password_reset.html",
	}

	// Get template file path
	templateFile, exists := templateFiles[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	// Parse template file
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse template file %s: %v", templateFile, err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

func (ms *MailService) SendHTMLEmail(to, subject, htmlBody string) error {
	// Create HTML email message
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlBody + "\r\n")

	auth := smtp.PlainAuth(ms.From, ms.To, ms.password, ms.smtpHost)
	if config.GetKey("MAIL_AUTH_ENABLED") == "false" {
		auth = nil
	}

	return smtp.SendMail(ms.smtpHost+":"+ms.smtpPort, auth, ms.From, []string{to}, msg)
}

func (ms *MailService) SendEmail(to, subject, body string) error {
	// Keep the original plain text method for backward compatibility
	auth := smtp.PlainAuth("", ms.To, ms.password, ms.smtpHost)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	if config.GetKey("MAIL_AUTH_ENABLED") == "false" {
		auth = nil
	}
	return smtp.SendMail(ms.smtpHost+":"+ms.smtpPort, auth, ms.From, []string{to}, msg)
}
