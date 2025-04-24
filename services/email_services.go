package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"

	"go.uber.org/zap"
)

// EmailConfig holds the configuration for email sending
type EmailConfig struct {
	SMTPServer   string
	SMTPPort     int
	Username     string
	Password     string
	FromEmail    string
	FromName     string
	TemplatesDir string
}

// EmailService handles all email sending operations
type EmailService struct {
	logger *zap.Logger
	config EmailConfig
}

// NewEmailService creates a new instance of EmailService
func NewEmailService(logger *zap.Logger, config EmailConfig) *EmailService {
	return &EmailService{
		logger: logger,
		config: config,
	}
}

// EmailData contains data used in email templates
type EmailData struct {
	GuestName       string
	BookingID       uint
	CheckIn         time.Time
	CheckOut        time.Time
	RoomType        string
	RoomNumber      string
	TotalPrice      float64
	OnsenDate       time.Time
	OnsenTimeSlot   string
	OnsenPrice      float64
	HotelName       string
	HotelAddress    string
	HotelPhone      string
	BookingDate     time.Time
	CancellationURL string
	CustomMessage   string
	Subject         string
}

// sendEmail sends an email using the SMTP configuration
func (es *EmailService) sendEmail(to, subject, body string) error {
	// Set up authentication information
	auth := smtp.PlainAuth(
		"",
		es.config.Username,
		es.config.Password,
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

// renderTemplate renders an email template with the provided data
func (es *EmailService) renderTemplate(templateName string, data EmailData) (string, error) {
	// Construct template path
	templatePath := fmt.Sprintf("%s/%s.html", es.config.TemplatesDir, templateName)

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		es.logger.Error("failed to parse email template",
			zap.String("template", templateName),
			zap.Error(err))
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute the template with data
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		es.logger.Error("failed to execute email template",
			zap.String("template", templateName),
			zap.Error(err))
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return tpl.String(), nil
}

// SendRoomBookingConfirmation sends a booking confirmation email
func (es *EmailService) SendRoomBookingConfirmation(booking *models.RoomBooking, guest *models.Guest, room *models.Room) error {
	// Prepare email data
	data := EmailData{
		GuestName:       guest.Name,
		BookingID:       booking.ID,
		CheckIn:         booking.CheckIn,
		CheckOut:        booking.CheckOut,
		RoomType:        room.Type,
		RoomNumber:      room.RoomNo,
		TotalPrice:      booking.TotalPrice,
		HotelName:       "Traditional Japanese Ryokan", // Configure these values in a real app
		HotelAddress:    "123 Sakura Street, Kyoto, Japan",
		HotelPhone:      "+81 123-456-7890",
		BookingDate:     booking.CreatedAt,
		CancellationURL: fmt.Sprintf("https://yourdomain.com/bookings/cancel/%d", booking.ID),
	}

	// Render template
	body, err := es.renderTemplate("room_booking_confirmation", data)
	if err != nil {
		return err
	}

	// Set subject
	subject := fmt.Sprintf("Booking Confirmation #%d - Traditional Japanese Ryokan", booking.ID)

	// Send email
	return es.sendEmail(guest.Email, subject, body)
}

// SendOnsenBookingConfirmation sends an onsen booking confirmation email
func (es *EmailService) SendOnsenBookingConfirmation(booking *models.OnsenBooking, guest *models.Guest) error {
	// Prepare email data
	data := EmailData{
		GuestName:     guest.Name,
		BookingID:     booking.ID,
		OnsenDate:     booking.Date,
		OnsenTimeSlot: booking.TimeSlot,
		OnsenPrice:    booking.Price,
		HotelName:     "Traditional Japanese Ryokan",
		HotelAddress:  "123 Sakura Street, Kyoto, Japan",
		HotelPhone:    "+81 123-456-7890",
		BookingDate:   booking.CreatedAt,
	}

	// Render template
	body, err := es.renderTemplate("onsen_booking_confirmation", data)
	if err != nil {
		return err
	}

	// Set subject
	subject := fmt.Sprintf("Private Onsen Reservation #%d - Traditional Japanese Ryokan", booking.ID)

	// Send email
	return es.sendEmail(guest.Email, subject, body)
}

// SendBookingCancellationNotice sends a cancellation notice email
func (es *EmailService) SendBookingCancellationNotice(booking *models.RoomBooking, guest *models.Guest, room *models.Room) error {
	// Prepare email data
	data := EmailData{
		GuestName:   guest.Name,
		BookingID:   booking.ID,
		CheckIn:     booking.CheckIn,
		CheckOut:    booking.CheckOut,
		RoomType:    room.Type,
		RoomNumber:  room.RoomNo,
		HotelName:   "Traditional Japanese Ryokan",
		HotelPhone:  "+81 123-456-7890",
		BookingDate: booking.CreatedAt,
	}

	// Render template
	body, err := es.renderTemplate("booking_cancellation", data)
	if err != nil {
		return err
	}

	// Set subject
	subject := fmt.Sprintf("Booking Cancellation #%d - Traditional Japanese Ryokan", booking.ID)

	// Send email
	return es.sendEmail(guest.Email, subject, body)
}

// SendOnsenCancellationNotice sends an onsen cancellation notice email
func (es *EmailService) SendOnsenCancellationNotice(booking *models.OnsenBooking, guest *models.Guest) error {
	// Prepare email data
	data := EmailData{
		GuestName:     guest.Name,
		BookingID:     booking.ID,
		OnsenDate:     booking.Date,
		OnsenTimeSlot: booking.TimeSlot,
		HotelName:     "Traditional Japanese Ryokan",
		HotelPhone:    "+81 123-456-7890",
	}

	// Render template
	body, err := es.renderTemplate("onsen_cancellation", data)
	if err != nil {
		return err
	}

	// Set subject
	subject := fmt.Sprintf("Onsen Reservation Cancellation #%d - Traditional Japanese Ryokan", booking.ID)

	// Send email
	return es.sendEmail(guest.Email, subject, body)
}

// SendCustomEmail sends a custom email to a guest
func (es *EmailService) SendCustomEmail(email string, subject string, message string) error {
	// Prepare email data
	data := EmailData{
		CustomMessage: message,
		HotelName:     "Traditional Japanese Ryokan",
		HotelAddress:  "123 Sakura Street, Kyoto, Japan",
		HotelPhone:    "+81 123-456-7890",
		Subject:       subject,
	}

	// Render template
	body, err := es.renderTemplate("custom_message", data)
	if err != nil {
		return err
	}

	// Send email
	return es.sendEmail(email, subject, body)
}

// SendBookingReminder sends a booking reminder email
func (es *EmailService) SendBookingReminder(booking *models.RoomBooking, guest *models.Guest, room *models.Room) error {
	// Prepare email data
	data := EmailData{
		GuestName:    guest.Name,
		BookingID:    booking.ID,
		CheckIn:      booking.CheckIn,
		CheckOut:     booking.CheckOut,
		RoomType:     room.Type,
		RoomNumber:   room.RoomNo,
		HotelName:    "Traditional Japanese Ryokan",
		HotelAddress: "123 Sakura Street, Kyoto, Japan",
		HotelPhone:   "+81 123-456-7890",
	}

	// Render template
	body, err := es.renderTemplate("booking_reminder", data)
	if err != nil {
		return err
	}

	// Set subject
	subject := fmt.Sprintf("Your Stay Reminder - Traditional Japanese Ryokan - %s",
		booking.CheckIn.Format("Jan 2"))

	// Send email
	return es.sendEmail(guest.Email, subject, body)
}
