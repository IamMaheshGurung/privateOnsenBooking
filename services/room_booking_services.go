package services

import (
	"fmt"
	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RoomBookingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewRoomBookingService(db *gorm.DB, logger *zap.Logger) *RoomBookingService {
	return &RoomBookingService{
		db:     db,
		logger: logger,
	}
}

// Check if a room is available between checkIn and checkOut
func (rbs *RoomBookingService) IsRoomAvailable(roomID uint, checkIn, checkOut time.Time) (bool, error) {
	var count int64

	err := rbs.db.Model(&models.RoomBooking{}).
		Where("room_id = ? AND check_in < ? AND check_out > ?", roomID, checkOut, checkIn).
		Count(&count).Error

	if err != nil {
		rbs.logger.Error("failed to check room availability", zap.Error(err))
		return false, fmt.Errorf("failed to check room availability: %w", err)
	}

	return count == 0, nil
}

// Create a new room booking
func (rbs *RoomBookingService) CreateBooking(guestID, roomID uint, checkIn, checkOut time.Time) error {
	available, err := rbs.IsRoomAvailable(roomID, checkIn, checkOut)
	if err != nil {
		return err
	}

	if !available {
		rbs.logger.Warn("room is not available", zap.Uint("roomID", roomID))
		return fmt.Errorf("room is not available")
	}

	booking := &models.RoomBooking{
		GuestID:   guestID,
		RoomID:    roomID,
		CheckIn:   checkIn,
		CheckOut:  checkOut,
		CreatedAt: time.Now(),
	}

	if err := rbs.db.Create(booking).Error; err != nil {
		rbs.logger.Error("failed to create booking", zap.Error(err))
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

// Update an existing booking by ID
func (rbs *RoomBookingService) UpdateBookingByID(bookingID uint, name string, checkIn, checkOut time.Time) error {
	var booking models.RoomBooking
	if err := rbs.db.First(&booking, bookingID).Error; err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}
	booking.Name = name
	booking.CheckIn = checkIn
	booking.CheckOut = checkOut

	if err := rbs.db.Save(&booking).Error; err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	return nil
}

// Fetch bookings by check-in date
func (rbs *RoomBookingService) GetBookingByCheckInDate(checkIn time.Time) ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking
	if err := rbs.db.Where("check_in = ?", checkIn).Find(&bookings).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch bookings: %w", err)
	}
	return bookings, nil
}

// Cancel/Delete booking by ID
func (rbs *RoomBookingService) CancelBookingByID(bookingID uint) error {
	if err := rbs.db.Delete(&models.RoomBooking{}, bookingID).Error; err != nil {
		rbs.logger.Error("failed to cancel booking", zap.Error(err))
		return fmt.Errorf("failed to cancel booking: %w", err)
	}
	return nil
}

// Get all bookings of a specific guest
func (rbs *RoomBookingService) GetBookingsByGuestID(guestID uint) ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking
	if err := rbs.db.Where("guest_id = ?", guestID).Find(&bookings).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch bookings for guest: %w", err)
	}
	return bookings, nil
}
