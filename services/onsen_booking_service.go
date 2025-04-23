package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OnsenBookingService handles all onsen booking related operations
type OnsenBookingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOnsenBookingService creates a new instance of OnsenBookingService
func NewOnsenBookingService(db *gorm.DB, logger *zap.Logger) *OnsenBookingService {
	return &OnsenBookingService{
		db:     db,
		logger: logger,
	}
}

// GetAvailableTimeSlots returns available time slots for a specific date
func (obs *OnsenBookingService) GetAvailableTimeSlots(date time.Time) ([]string, error) {
	// Define all possible time slots
	allTimeSlots := []string{
		"09:00-10:00", "10:30-11:30", "12:00-13:00",
		"13:30-14:30", "15:00-16:00", "16:30-17:30",
		"18:00-19:00", "19:30-20:30", "21:00-22:00",
	}

	// Get booked time slots for the date
	var bookedSlots []string
	if err := obs.db.Model(&models.OnsenBooking{}).
		Where("date = ? AND status != ?", date, models.BookingStatusCancelled).
		Pluck("time_slot", &bookedSlots).Error; err != nil {
		obs.logger.Error("failed to get booked time slots", zap.Error(err))
		return nil, fmt.Errorf("failed to get booked time slots: %w", err)
	}

	// Find available slots
	availableSlots := []string{}
	for _, slot := range allTimeSlots {
		isBooked := false
		for _, bookedSlot := range bookedSlots {
			if slot == bookedSlot {
				isBooked = true
				break
			}
		}
		if !isBooked {
			availableSlots = append(availableSlots, slot)
		}
	}

	return availableSlots, nil
}

// IsTimeSlotAvailable checks if a time slot is available for a specific date
func (obs *OnsenBookingService) IsTimeSlotAvailable(date time.Time, timeSlot string) (bool, error) {
	var count int64

	if err := obs.db.Model(&models.OnsenBooking{}).
		Where("date = ? AND time_slot = ? AND status != ?",
			date, timeSlot, models.BookingStatusCancelled).
		Count(&count).Error; err != nil {
		obs.logger.Error("failed to check time slot availability", zap.Error(err))
		return false, fmt.Errorf("failed to check time slot availability: %w", err)
	}

	return count == 0, nil
}

// CreateOnsenBooking creates a new onsen booking
func (obs *OnsenBookingService) CreateOnsenBooking(guestID, roomID uint, bookingID uint, date time.Time, timeSlot string) (*models.OnsenBooking, error) {
	// Check if the time slot is available
	available, err := obs.IsTimeSlotAvailable(date, timeSlot)
	if err != nil {
		return nil, err
	}

	if !available {
		obs.logger.Warn("time slot is not available", zap.Time("date", date), zap.String("timeSlot", timeSlot))
		return nil, errors.New("time slot is not available")
	}

	// Set default onsen price
	const onsenPrice = 5000.0 // 5000 JPY

	// Create booking
	booking := models.OnsenBooking{
		GuestID:   guestID,
		RoomID:    roomID,
		BookingID: bookingID,
		Date:      date,
		TimeSlot:  timeSlot,
		Status:    models.BookingStatusConfirmed,
		Price:     onsenPrice,
	}

	if err := obs.db.Create(&booking).Error; err != nil {
		obs.logger.Error("failed to create onsen booking", zap.Error(err))
		return nil, fmt.Errorf("failed to create onsen booking: %w", err)
	}

	obs.logger.Info("onsen booking created",
		zap.Uint("id", booking.ID),
		zap.Time("date", date),
		zap.String("timeSlot", timeSlot))

	return &booking, nil
}

// GetOnsenBookingByID retrieves an onsen booking by ID
func (obs *OnsenBookingService) GetOnsenBookingByID(id uint) (*models.OnsenBooking, error) {
	var booking models.OnsenBooking

	if err := obs.db.Preload("Guest").Preload("Room").First(&booking, id).Error; err != nil {
		obs.logger.Error("failed to get onsen booking", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get onsen booking: %w", err)
	}

	return &booking, nil
}

// CancelOnsenBooking cancels an onsen booking
func (obs *OnsenBookingService) CancelOnsenBooking(id uint) error {
	booking, err := obs.GetOnsenBookingByID(id)
	if err != nil {
		return err
	}

	booking.Status = models.BookingStatusCancelled

	if err := obs.db.Save(booking).Error; err != nil {
		obs.logger.Error("failed to cancel onsen booking", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to cancel onsen booking: %w", err)
	}

	obs.logger.Info("onsen booking cancelled", zap.Uint("id", id))
	return nil
}

// GetOnsenBookingsByDate retrieves all onsen bookings for a specific date
func (obs *OnsenBookingService) GetOnsenBookingsByDate(date time.Time) ([]models.OnsenBooking, error) {
	var bookings []models.OnsenBooking

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := obs.db.Preload("Guest").Preload("Room").
		Where("date >= ? AND date < ? AND status != ?",
			startOfDay, endOfDay, models.BookingStatusCancelled).
		Order("time_slot").
		Find(&bookings).Error; err != nil {
		obs.logger.Error("failed to get onsen bookings by date", zap.Time("date", date), zap.Error(err))
		return nil, fmt.Errorf("failed to get onsen bookings by date: %w", err)
	}

	return bookings, nil
}

// GetUpcomingOnsenBookings retrieves upcoming onsen bookings
func (obs *OnsenBookingService) GetUpcomingOnsenBookings() ([]models.OnsenBooking, error) {
	var bookings []models.OnsenBooking

	now := time.Now()

	if err := obs.db.Preload("Guest").Preload("Room").
		Where("date > ? AND status = ?",
			now, models.BookingStatusConfirmed).
		Order("date, time_slot").
		Find(&bookings).Error; err != nil {
		obs.logger.Error("failed to get upcoming onsen bookings", zap.Error(err))
		return nil, fmt.Errorf("failed to get upcoming onsen bookings: %w", err)
	}

	return bookings, nil
}
