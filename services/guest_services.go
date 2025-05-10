package services

import (
	"fmt"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GuestService handles all guest-related operations
type GuestService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewGuestService creates a new instance of GuestService
func NewGuestService(db *gorm.DB, logger *zap.Logger) *GuestService {
	return &GuestService{
		db:     db,
		logger: logger,
	}
}

// GetGuestByID retrieves a guest by ID
func (gs *GuestService) GetGuestByID(id uint) (*models.Guest, error) {
	var guest models.Guest
	if err := gs.db.First(&guest, id).Error; err != nil {
		gs.logger.Error("Failed to get guest by ID", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get guest by ID: %w", err)
	}
	return &guest, nil
}

// GetGuestByEmail retrieves a guest by email
func (gs *GuestService) GetGuestByEmail(email string) (*models.Guest, error) {
	var guest models.Guest
	if err := gs.db.Where("email = ?", email).First(&guest).Error; err != nil {
		gs.logger.Error("Failed to get guest by email", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("failed to get guest by email: %w", err)
	}
	return &guest, nil
}

// CreateGuest creates a new guest
func (gs *GuestService) CreateGuest(guest *models.Guest) (*models.Guest, error) {
	if err := gs.db.Create(guest).Error; err != nil {
		gs.logger.Error("Failed to create guest", zap.Error(err))
		return nil, fmt.Errorf("failed to create guest: %w", err)
	}
	return guest, nil
}

// UpdateGuest updates an existing guest
func (gs *GuestService) UpdateGuest(id uint, name, email, phone string) (*models.Guest, error) {
	var guest models.Guest
	if err := gs.db.First(&guest, id).Error; err != nil {
		gs.logger.Error("Failed to find guest for update", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to find guest: %w", err)
	}

	// Update fields
	if name != "" {
		guest.Name = name
	}
	if email != "" {
		guest.Email = email
	}
	if phone != "" {
		guest.Phone = phone
	}

	if err := gs.db.Save(&guest).Error; err != nil {
		gs.logger.Error("Failed to update guest", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update guest: %w", err)
	}

	return &guest, nil
}

// CreateOrGetGuest creates a new guest or returns an existing one
func (gs *GuestService) CreateOrGetGuest(name, email, phone string) (*models.Guest, error) {
	var guest models.Guest

	// Check if guest with this email already exists
	err := gs.db.Where("email = ?", email).First(&guest).Error
	if err == nil {
		// Guest exists, update info if needed
		updated := false

		if guest.Name != name && name != "" {
			guest.Name = name
			updated = true
		}

		if guest.Phone != phone && phone != "" {
			guest.Phone = phone
			updated = true
		}

		if updated {
			if err := gs.db.Save(&guest).Error; err != nil {
				gs.logger.Error("Failed to update existing guest", zap.String("email", email), zap.Error(err))
				return nil, fmt.Errorf("failed to update guest: %w", err)
			}
		}

		return &guest, nil
	}

	// Guest doesn't exist, create new
	newGuest := models.Guest{
		Name:  name,
		Email: email,
		Phone: phone,
	}

	if err := gs.db.Create(&newGuest).Error; err != nil {
		gs.logger.Error("Failed to create new guest", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("failed to create guest: %w", err)
	}

	return &newGuest, nil
}

// DeleteGuest deletes a guest
func (gs *GuestService) DeleteGuest(id uint) error {
	// Check if guest exists
	var guest models.Guest
	if err := gs.db.First(&guest, id).Error; err != nil {
		gs.logger.Error("Failed to find guest for deletion", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to find guest: %w", err)
	}

	// Delete the guest
	if err := gs.db.Delete(&guest).Error; err != nil {
		gs.logger.Error("Failed to delete guest", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete guest: %w", err)
	}

	return nil
}

// GetAllGuests returns all guests
func (gs *GuestService) GetAllGuests() ([]models.Guest, error) {
	var guests []models.Guest
	if err := gs.db.Find(&guests).Error; err != nil {
		gs.logger.Error("Failed to get all guests", zap.Error(err))
		return nil, fmt.Errorf("failed to get all guests: %w", err)
	}
	return guests, nil
}

// GetGuestBookingHistory retrieves booking history for a guest
func (gs *GuestService) GetGuestBookingHistory(guestID uint) ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking

	// Check if guest exists
	if err := gs.db.First(&models.Guest{}, guestID).Error; err != nil {
		gs.logger.Error("Failed to find guest for booking history", zap.Uint("guestID", guestID), zap.Error(err))
		return nil, fmt.Errorf("guest not found: %w", err)
	}

	// Get all bookings for the guest
	if err := gs.db.Preload("Room").Where("guest_id = ?", guestID).Find(&bookings).Error; err != nil {
		gs.logger.Error("Failed to get guest booking history", zap.Uint("guestID", guestID), zap.Error(err))
		return nil, fmt.Errorf("failed to get booking history: %w", err)
	}

	return bookings, nil
}
