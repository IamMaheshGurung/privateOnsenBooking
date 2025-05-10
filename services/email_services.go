package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"go.uber.org/zap"
)

// EmailConfig contains all the SMTP configuration options
type EmailConfig struct {
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	TemplatesDir string
	Environment  string // "development", "production", etc.
}

// EmailService handles sending email notifications
type EmailService struct {
	logger *zap.Logger
	config EmailConfig
}

// NewEmailService creates a new email service instance
func NewEmailService(logger *zap.Logger, config EmailConfig) *EmailService {
	// Set default values if not provided
	if config.FromName == "" {
		config.FromName = "Kwangdi Onsen"
	}

	if config.TemplatesDir == "" {
		config.TemplatesDir = "./templates/emails"
	}

	return &EmailService{
		logger: logger,
		config: config,
	}
}

// SendEmail sends an email with the given parameters
func (es *EmailService) SendEmail(to, subject, body string) error {
	// Skip sending in development mode if configured to do so
	if es.config.Environment == "development" && os.Getenv("SEND_EMAILS") != "true" {
		es.logger.Info("Email sending skipped in development mode",
			zap.String("to", to),
			zap.String("subject", subject))
		return nil
	}

	// Check if SMTP configuration is available
	if es.config.SMTPServer == "" || es.config.SMTPPort == 0 {
		es.logger.Warn("SMTP not configured, email not sent",
			zap.String("to", to),
			zap.String("subject", subject))
		return fmt.Errorf("SMTP not configured")
	}

	// Setup authentication
	auth := smtp.PlainAuth(
		"",
		es.config.SMTPUsername,
		es.config.SMTPPassword,
		es.config.SMTPServer,
	)

	// Prepare email headers
	fromHeader := fmt.Sprintf("From: %s <%s>\r\n", es.config.FromName, es.config.FromEmail)
	toHeader := fmt.Sprintf("To: %s\r\n", to)
	subjectHeader := fmt.Sprintf("Subject: %s\r\n", subject)
	mimeHeader := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	// Combine headers and body
	message := fromHeader + toHeader + subjectHeader + mimeHeader + body

	// Send email
	addr := fmt.Sprintf("%s:%d", es.config.SMTPServer, es.config.SMTPPort)
	err := smtp.SendMail(
		addr,
		auth,
		es.config.FromEmail,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		es.logger.Error("failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	es.logger.Info("email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject))
	return nil
}

// renderTemplate parses and renders an email template with the given data
func (es *EmailService) renderTemplate(templateName string, data interface{}) (string, error) {
	templatePath := filepath.Join(es.config.TemplatesDir, templateName+".html")

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		es.logger.Error("email template not found", zap.String("template", templatePath))
		return "", fmt.Errorf("email template not found: %s", templateName)
	}

	// Parse template file
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		es.logger.Error("failed to parse email template",
			zap.String("template", templatePath),
			zap.Error(err))
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute template with data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		es.logger.Error("failed to render email template",
			zap.String("template", templatePath),
			zap.Error(err))
		return "", fmt.Errorf("failed to render email template: %w", err)
	}

	return buf.String(), nil
}

// SendBookingConfirmation sends a booking confirmation email to the guest
func (es *EmailService) SendBookingConfirmation(booking *models.RoomBooking, guest *models.Guest, room *models.Room) error {
	// Skip if no guest email
	if guest == nil || guest.Email == "" {
		es.logger.Warn("no guest email available for booking confirmation",
			zap.Uint("bookingID", booking.ID),
			zap.Uint("guestID", booking.GuestID))
		return fmt.Errorf("no guest email available")
	}

	// Prepare template data
	data := map[string]interface{}{
		"Booking":      booking,
		"Guest":        guest,
		"Room":         room,
		"HotelName":    es.config.FromName,
		"CheckInDate":  booking.CheckIn.Format("Monday, January 2, 2006"),
		"CheckOutDate": booking.CheckOut.Format("Monday, January 2, 2006"),
		"TotalNights":  int(booking.CheckOut.Sub(booking.CheckIn).Hours() / 24),
		"TotalPrice":   fmt.Sprintf("%.2f", booking.TotalPrice),
		"Year":         time.Now().Year(),
	}

	// Render email template
	body, err := es.renderTemplate("booking_confirmation", data)
	if err != nil {
		return err
	}

	// Send email
	subject := fmt.Sprintf("Your Booking Confirmation #%s - %s", booking.ReferenceNumber, es.config.FromName)
	return es.SendEmail(guest.Email, subject, body)
}

// SendBookingCancellationNotice sends a booking cancellation confirmation email
func (es *EmailService) SendBookingCancellationNotice(booking *models.RoomBooking, guest *models.Guest, room *models.Room) error {
	// Skip if no guest email
	if guest == nil || guest.Email == "" {
		es.logger.Warn("no guest email available for cancellation notice",
			zap.Uint("bookingID", booking.ID))
		return fmt.Errorf("no guest email available")
	}

	cancellationFeeText := "No cancellation fee has been applied."
	if booking.CancellationFee > 0 {
		cancellationFeeText = fmt.Sprintf("A cancellation fee of %.2f has been applied.", booking.CancellationFee)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Booking":             booking,
		"Guest":               guest,
		"Room":                room,
		"HotelName":           es.config.FromName,
		"CancellationDate":    booking.CancelledAt.Format("Monday, January 2, 2006"),
		"CheckInDate":         booking.CheckIn.Format("Monday, January 2, 2006"),
		"CancellationFeeText": cancellationFeeText,
		"Year":                time.Now().Year(),
	}

	// Render email template
	body, err := es.renderTemplate("booking_cancellation", data)
	if err != nil {
		return err
	}

	// Send email
	subject := fmt.Sprintf("Booking Cancellation #%s - %s", booking.ReferenceNumber, es.config.FromName)
	return es.SendEmail(guest.Email, subject, body)
}

// SendCheckInReminder sends a reminder email before check-in date
func (es *EmailService) SendCheckInReminder(booking *models.RoomBooking, guest *models.Guest, room *models.Room) error {
	// Skip if no guest email
	if guest == nil || guest.Email == "" {
		es.logger.Warn("no guest email available for check-in reminder",
			zap.Uint("bookingID", booking.ID))
		return fmt.Errorf("no guest email available")
	}

	// Calculate days until check-in
	daysUntilCheckIn := int(booking.CheckIn.Sub(time.Now()).Hours() / 24)

	// Prepare template data
	data := map[string]interface{}{
		"Booking":          booking,
		"Guest":            guest,
		"Room":             room,
		"HotelName":        es.config.FromName,
		"CheckInDate":      booking.CheckIn.Format("Monday, January 2, 2006"),
		"CheckInTime":      "3:00 PM", // Modify as needed
		"DaysUntilCheckIn": daysUntilCheckIn,
		"Year":             time.Now().Year(),
	}

	// Render email template
	body, err := es.renderTemplate("checkin_reminder", data)
	if err != nil {
		return err
	}

	// Send email
	subject := fmt.Sprintf("Your Stay at %s Begins Soon", es.config.FromName)
	return es.SendEmail(guest.Email, subject, body)
}

// SendContactFormNotification sends an email to hotel staff when contact form is submitted
func (es *EmailService) SendContactFormNotification(name, email, message string) error {
	// Prepare template data
	data := map[string]interface{}{
		"Name":          name,
		"Email":         email,
		"Message":       message,
		"SubmittedDate": time.Now().Format("Monday, January 2, 2006 at 3:04 PM"),
		"Year":          time.Now().Year(),
	}

	// Render email template
	body, err := es.renderTemplate("contact_form_notification", data)
	if err != nil {
		return err
	}

	// Send email to the configured address (or default to the FromEmail)
	notificationEmail := os.Getenv("CONTACT_NOTIFICATION_EMAIL")
	if notificationEmail == "" {
		notificationEmail = es.config.FromEmail
	}

	subject := fmt.Sprintf("New Contact Form Submission - %s", name)
	return es.SendEmail(notificationEmail, subject, body)
}

// SendSpecialOfferEmail sends a promotional email to a guest
func (es *EmailService) SendSpecialOfferEmail(guest *models.Guest, offerTitle, offerDescription string, validUntil time.Time) error {
	// Skip if no guest email
	if guest == nil || guest.Email == "" {
		es.logger.Warn("no guest email available for special offer",
			zap.String("guestName", guest.Name))
		return fmt.Errorf("no guest email available")
	}

	// Prepare template data
	data := map[string]interface{}{
		"Guest":            guest,
		"HotelName":        es.config.FromName,
		"OfferTitle":       offerTitle,
		"OfferDescription": offerDescription,
		"ValidUntil":       validUntil.Format("Monday, January 2, 2006"),
		"Year":             time.Now().Year(),
	}

	// Render email template
	body, err := es.renderTemplate("special_offer", data)
	if err != nil {
		return err
	}

	// Send email
	subject := fmt.Sprintf("Special Offer: %s - %s", offerTitle, es.config.FromName)
	return es.SendEmail(guest.Email, subject, body)
}

// SendAdminNotification sends a notification email to admin
func (es *EmailService) SendAdminNotification(subject, message string, data map[string]interface{}) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		es.logger.Warn("admin email not configured, notification not sent")
		return fmt.Errorf("admin email not configured")
	}

	// Add default data
	if data == nil {
		data = make(map[string]interface{})
	}

	data["Subject"] = subject
	data["Message"] = message
	data["Timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	data["HotelName"] = es.config.FromName
	data["Year"] = time.Now().Year()

	// Render email template
	body, err := es.renderTemplate("admin_notification", data)
	if err != nil {
		return err
	}

	// Send email
	return es.SendEmail(adminEmail, subject, body)
}
