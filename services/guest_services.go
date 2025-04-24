package services

import (
	"errors"
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

// CreateGuest creates a new guest record
func (gs *GuestService) CreateGuest(name, email, phone string) (*models.Guest, error) {
	// Validate input
	if name == "" || email == "" || phone == "" {
		return nil, errors.New("name, email, and phone are required")
	}

	// Check if guest with email already exists
	var existingGuest models.Guest
	if err := gs.db.Where("email = ?", email).First(&existingGuest).Error; err == nil {
		gs.logger.Info("guest with this email already exists", zap.String("email", email))
		return &existingGuest, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		gs.logger.Error("error checking for existing guest", zap.Error(err))
		return nil, fmt.Errorf("error checking for existing guest: %w", err)
	}

	// Create new guest
	guest := models.Guest{
		Name:  name,
		Email: email,
		Phone: phone,
	}

	if err := gs.db.Create(&guest).Error; err != nil {
		gs.logger.Error("failed to create guest", zap.Error(err))
		return nil, fmt.Errorf("failed to create guest: %w", err)
	}

	gs.logger.Info("guest created successfully", zap.String("email", email), zap.Uint("id", guest.ID))
	return &guest, nil
}

// GetGuestByID retrieves a guest by their ID
func (gs *GuestService) GetGuestByID(id uint) (*models.Guest, error) {
	var guest models.Guest

	if err := gs.db.First(&guest, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("guest not found with ID %d", id)
		}
		gs.logger.Error("failed to get guest", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get guest: %w", err)
	}

	return &guest, nil
}

// GetGuestByEmail retrieves a guest by their email
func (gs *GuestService) GetGuestByEmail(email string) (*models.Guest, error) {
	var guest models.Guest

	if err := gs.db.Where("email = ?", email).First(&guest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("guest not found with email %s", email)
		}
		gs.logger.Error("failed to get guest by email", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("failed to get guest by email: %w", err)
	}

	return &guest, nil
}

// UpdateGuest updates guest information
func (gs *GuestService) UpdateGuest(id uint, name, email, phone string) (*models.Guest, error) {
	guest, err := gs.GetGuestByID(id)
	if err != nil {
		return nil, err
	}

	// Check if another guest already has this email
	if email != guest.Email {
		var existingGuest models.Guest
		if err := gs.db.Where("email = ? AND id != ?", email, id).First(&existingGuest).Error; err == nil {
			return nil, fmt.Errorf("another guest with email %s already exists", email)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			gs.logger.Error("error checking for existing guest", zap.Error(err))
			return nil, fmt.Errorf("error checking for existing guest: %w", err)
		}
	}

	// Update fields if provided
	if name != "" {
		guest.Name = name
	}
	if email != "" {
		guest.Email = email
	}
	if phone != "" {
		guest.Phone = phone
	}

	if err := gs.db.Save(guest).Error; err != nil {
		gs.logger.Error("failed to update guest", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update guest: %w", err)
	}

	gs.logger.Info("guest updated successfully", zap.Uint("id", id))
	return guest, nil
}

// DeleteGuest deletes a guest by their ID
func (gs *GuestService) DeleteGuest(id uint) error {
	// Check if guest exists
	if _, err := gs.GetGuestByID(id); err != nil {
		return err
	}

	// Check if guest has active bookings
	var bookingCount int64
	if err := gs.db.Model(&models.RoomBooking{}).
		Where("guest_id = ? AND status IN (?)", id, []string{
			models.BookingStatusConfirmed, models.BookingStatusCheckedIn,
		}).Count(&bookingCount).Error; err != nil {
		gs.logger.Error("failed to check for active bookings", zap.Uint("guestID", id), zap.Error(err))
		return fmt.Errorf("failed to check for active bookings: %w", err)
	}

	if bookingCount > 0 {
		return fmt.Errorf("cannot delete guest with active bookings")
	}

	// Delete the guest
	if err := gs.db.Delete(&models.Guest{}, id).Error; err != nil {
		gs.logger.Error("failed to delete guest", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete guest: %w", err)
	}

	gs.logger.Info("guest deleted successfully", zap.Uint("id", id))
	return nil
}

// GetAllGuests retrieves all guests with optional filtering
func (gs *GuestService) GetAllGuests(limit, offset int, searchTerm string) ([]models.Guest, int64, error) {
	var guests []models.Guest
	var totalCount int64

	// Build the query
	query := gs.db.Model(&models.Guest{})

	// Apply search if provided
	if searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		query = query.Where("name LIKE ? OR email LIKE ? OR phone LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		gs.logger.Error("failed to count guests", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count guests: %w", err)
	}

	// Apply pagination and get results
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&guests).Error; err != nil {
		gs.logger.Error("failed to get guests", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get guests: %w", err)
	}

	return guests, totalCount, nil
}

// GetGuestBookingHistory retrieves booking history for a guest
func (gs *GuestService) GetGuestBookingHistory(guestID uint) ([]models.RoomBooking, error) {
	// Check if guest exists
	if _, err := gs.GetGuestByID(guestID); err != nil {
		return nil, err
	}

	var bookings []models.RoomBooking
	if err := gs.db.Where("guest_id = ?", guestID).
		Preload("Room").
		Order("check_in DESC").
		Find(&bookings).Error; err != nil {
		gs.logger.Error("failed to get guest booking history",
			zap.Uint("guestID", guestID), zap.Error(err))
		return nil, fmt.Errorf("failed to get guest booking history: %w", err)
	}

	return bookings, nil
}

// GetGuestOnsenBookingHistory retrieves onsen booking history for a guest
func (gs *GuestService) GetGuestOnsenBookingHistory(guestID uint) ([]models.OnsenBooking, error) {
	// Check if guest exists
	if _, err := gs.GetGuestByID(guestID); err != nil {
		return nil, err
	}

	var bookings []models.OnsenBooking
	if err := gs.db.Where("guest_id = ?", guestID).
		Preload("Room").
		Order("date DESC, time_slot ASC").
		Find(&bookings).Error; err != nil {
		gs.logger.Error("failed to get guest onsen booking history",
			zap.Uint("guestID", guestID), zap.Error(err))
		return nil, fmt.Errorf("failed to get guest onsen booking history: %w", err)
	}

	return bookings, nil
}

// CreateOrGetGuest creates a new guest or returns existing guest if email already exists
func (gs *GuestService) CreateOrGetGuest(name, email, phone string) (*models.Guest, error) {
	// Try to find existing guest
	existingGuest, err := gs.GetGuestByEmail(email)
	if err == nil {
		// Guest exists, update info if necessary
		needsUpdate := false
		if name != "" && existingGuest.Name != name {
			existingGuest.Name = name
			needsUpdate = true
		}
		if phone != "" && existingGuest.Phone != phone {
			existingGuest.Phone = phone
			needsUpdate = true
		}

		if needsUpdate {
			if err := gs.db.Save(existingGuest).Error; err != nil {
				gs.logger.Error("failed to update existing guest",
					zap.String("email", email), zap.Error(err))
				return nil, fmt.Errorf("failed to update existing guest: %w", err)
			}
		}
		return existingGuest, nil
	}

	// Guest doesn't exist, create new one
	if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == fmt.Sprintf("guest not found with email %s", email) {
		return gs.CreateGuest(name, email, phone)
	}

	// Other error occurred
	return nil, err
}
