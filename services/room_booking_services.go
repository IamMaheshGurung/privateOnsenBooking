package services

import (
	"fmt"

	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoomBookingService handles all room booking related operations
type RoomBookingService struct {
	db           *gorm.DB
	logger       *zap.Logger
	emailservice *EmailService
}

// NewRoomBookingService creates a new instance of RoomBookingService
func NewRoomBookingService(db *gorm.DB, logger *zap.Logger, emailservice *EmailService) *RoomBookingService {
	return &RoomBookingService{
		db:           db,
		logger:       logger,
		emailservice: emailservice,
	}
}

func (rbs *RoomBookingService) GetAllRooms() ([]*models.Room, error) {
	var rooms []*models.Room
	if err := rbs.db.Find(&rooms).Error; err != nil {
		rbs.logger.Error("failed to fetch rooms", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch rooms: %w", err)
	}

	return rooms, nil
}

func (rbs *RoomBookingService) GetRoomByID(id uint) (*models.Room, error) {
	var room *models.Room
	if err := rbs.db.Find(&room, id).Error; err != nil {
		rbs.logger.Error("failed to fetch room", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch room: %w", err)
	}
	return room, nil
}

func (rbs *RoomBookingService) CreateRoom(room models.Room) error {
	if err := rbs.db.Create(room).Error; err != nil {
		rbs.logger.Error("failed to create room", zap.Error(err))
		return fmt.Errorf("failed to create room: %w", err)
	}

	return nil
}

func (rbs *RoomBookingService) UpdateRoom(room models.Room) (*models.Room, error) {
	// Create an updated room using keyed fields
	updatedRoom := models.Room{
		ID:            room.ID,
		RoomNo:        room.RoomNo,
		Type:          room.Type,
		Capacity:      room.Capacity,
		PricePerNight: room.PricePerNight,
		Description:   room.Description,
		Amenities:     room.Amenities,
		ImageURL:      room.ImageURL,
	}

	// Update the room in the database
	if err := rbs.db.Save(&updatedRoom).Error; err != nil {
		rbs.logger.Error("failed to update room", zap.Error(err))
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	return &updatedRoom, nil
}

func (rbs *RoomBookingService) DeleteRoom(id uint) error {
	var room models.Room
	if err := rbs.db.First(&room, id).Error; err != nil {
		rbs.logger.Error("failed to find room", zap.Error(err))
		return fmt.Errorf("failed to find room: %w", err)
	}

	if err := rbs.db.Delete(&room).Error; err != nil {
		rbs.logger.Error("failed to delete room", zap.Error(err))
		return fmt.Errorf("failed to delete room: %w", err)
	}

	return nil
}

// Check if a room is available between checkIn and checkOut
func (rbs *RoomBookingService) IsRoomAvailable(roomID uint, checkIn, checkOut time.Time) (bool, error) {
	var count int64

	err := rbs.db.Model(&models.RoomBooking{}).
		Where("room_id = ? AND check_in < ? AND check_out > ? AND status != ?",
			roomID, checkOut, checkIn, models.BookingStatusCancelled).
		Count(&count).Error

	if err != nil {
		rbs.logger.Error("failed to check room availability", zap.Error(err))
		return false, fmt.Errorf("failed to check room availability: %w", err)
	}

	return count == 0, nil
}

func (rbs *RoomBookingService) IsRoomAvailableForUpdate(roomID uint, bookingID uint, checkIn, checkOut time.Time) (bool, error) {
	var count int64

	// Count conflicting bookings, EXCLUDING the current booking being updated
	if err := rbs.db.Model(&models.RoomBooking{}).
		Where("room_id = ? AND id != ? AND status != ? AND check_in < ? AND check_out > ?",
			roomID, bookingID, models.BookingStatusCancelled, checkOut, checkIn).
		Count(&count).Error; err != nil {
		rbs.logger.Error("failed to check room availability for update", zap.Error(err))
		return false, fmt.Errorf("failed to check room availability for update: %w", err)
	}

	// Room is available if there are no conflicts with other bookings
	return count == 0, nil
}

// GetAvailableRooms returns all rooms that are available between checkIn and checkOut
func (rbs *RoomBookingService) GetAvailableRooms(checkIn, checkOut time.Time) ([]models.Room, error) {
	var allRooms []models.Room

	if err := rbs.db.Find(&allRooms).Error; err != nil {
		rbs.logger.Error("failed to fetch rooms", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch rooms: %w", err)
	}

	var availableRooms []models.Room
	for _, room := range allRooms {
		available, err := rbs.IsRoomAvailable(room.ID, checkIn, checkOut)
		if err != nil {
			return nil, err
		}
		if available {
			availableRooms = append(availableRooms, room)
		}
	}

	return availableRooms, nil
}

// CreateBooking creates a new room booking
func (rbs *RoomBookingService) CreateBooking(guestID, roomID uint, checkIn, checkOut time.Time) (*models.RoomBooking, error) {
	available, err := rbs.IsRoomAvailable(roomID, checkIn, checkOut)
	if err != nil {
		return nil, err
	}

	if !available {
		rbs.logger.Warn("room is not available", zap.Uint("roomID", roomID))
		return nil, fmt.Errorf("room is not available")
	}

	booking := models.RoomBooking{
		GuestID:  guestID,
		RoomID:   roomID,
		CheckIn:  checkIn,
		CheckOut: checkOut,
		Status:   models.BookingStatusConfirmed,
	}

	var guest models.Guest
	if err := rbs.db.First(&guest, guestID).Error; err != nil {
		rbs.logger.Error("failed to fetch guest", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch guest: %w", err)
	}

	var room models.Room
	if err := rbs.db.First(&room, roomID).Error; err != nil {
		rbs.logger.Error("Failed to fetch the room details", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch the rooms: %w", err)
	}

	if err := rbs.db.Create(&booking).Error; err != nil {
		rbs.logger.Error("failed to create booking", zap.Error(err))
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return &booking, nil
}

// UpdateBookingByID updates an existing booking
func (rbs *RoomBookingService) UpdateBookingByID(bookingID uint, checkIn, checkOut time.Time) error {
	var booking models.RoomBooking
	if err := rbs.db.First(&booking, bookingID).Error; err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	// If dates are changing, check availability
	if !booking.CheckIn.Equal(checkIn) || !booking.CheckOut.Equal(checkOut) {
		// Check if the room is available for the new check-in and check-out dates
		// We need to exclude the current booking from the check
		var count int64
		err := rbs.db.Model(&models.RoomBooking{}).
			Where("room_id = ? AND id != ? AND check_in < ? AND check_out > ? AND status != ?",
				booking.RoomID, bookingID, checkOut, checkIn, models.BookingStatusCancelled).
			Count(&count).Error

		if err != nil {
			rbs.logger.Error("failed to check room availability", zap.Error(err))
			return fmt.Errorf("failed to check room availability: %w", err)
		}

		if count > 0 {
			rbs.logger.Warn("room is not available for update", zap.Uint("roomID", booking.RoomID))
			return fmt.Errorf("room is not available for the new dates")
		}
	}

	// Update the booking details
	booking.CheckIn = checkIn
	booking.CheckOut = checkOut

	if err := rbs.db.Save(&booking).Error; err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	return nil
}

// GetBookingByID retrieves a booking by its ID
func (rbs *RoomBookingService) GetBookingByID(bookingID uint) (*models.RoomBooking, error) {
	var booking models.RoomBooking

	if err := rbs.db.Preload("Guest").Preload("Room").First(&booking, bookingID).Error; err != nil {
		rbs.logger.Error("failed to get booking", zap.Uint("bookingID", bookingID), zap.Error(err))
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

// GetBookingByCheckInDate retrieves all bookings for a specific check-in date
func (rbs *RoomBookingService) GetBookingByCheckInDate(date time.Time) ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := rbs.db.Preload("Guest").Preload("Room").
		Where("check_in >= ? AND check_in < ? AND status != ?",
			startOfDay, endOfDay, models.BookingStatusCancelled).
		Find(&bookings).Error; err != nil {
		rbs.logger.Error("failed to get bookings by check-in date", zap.Time("date", date), zap.Error(err))
		return nil, fmt.Errorf("failed to get bookings by check-in date: %w", err)
	}

	return bookings, nil
}

// GetAllBookings retrieves all bookings with optional filtering
func (rbs *RoomBookingService) GetAllBookings(status string, future bool) ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking

	db := rbs.db.Preload("Guest").Preload("Room")

	if status != "" {
		db = db.Where("status = ?", status)
	}

	if future {
		db = db.Where("check_in > ?", time.Now())
	}

	if err := db.Find(&bookings).Error; err != nil {
		rbs.logger.Error("failed to get all bookings", zap.Error(err))
		return nil, fmt.Errorf("failed to get all bookings: %w", err)
	}

	return bookings, nil
}

// CancelBookingByID cancels a booking by its ID
func (rbs *RoomBookingService) CancelBookingByID(bookingID uint) error {
	var booking models.RoomBooking

	if err := rbs.db.First(&booking, bookingID).Error; err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	booking.Status = models.BookingStatusCancelled

	if err := rbs.db.Save(&booking).Error; err != nil {
		rbs.logger.Error("failed to cancel booking", zap.Uint("bookingID", bookingID), zap.Error(err))
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	return nil
}

// GetBookingsByGuestID retrieves all bookings for a specific guest
func (rbs *RoomBookingService) GetBookingsByGuestID(guestID uint) ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking

	if err := rbs.db.Preload("Room").Where("guest_id = ?", guestID).Find(&bookings).Error; err != nil {
		rbs.logger.Error("failed to get bookings by guest ID", zap.Uint("guestID", guestID), zap.Error(err))
		return nil, fmt.Errorf("failed to get bookings by guest ID: %w", err)
	}

	return bookings, nil
}

// GetUpcomingBookings retrieves all upcoming bookings
func (rbs *RoomBookingService) GetUpcomingBookings() ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking

	if err := rbs.db.Preload("Guest").Preload("Room").
		Where("check_in > ? AND status = ?", time.Now(), models.BookingStatusConfirmed).
		Find(&bookings).Error; err != nil {
		rbs.logger.Error("failed to get upcoming bookings", zap.Error(err))
		return nil, fmt.Errorf("failed to get upcoming bookings: %w", err)
	}

	return bookings, nil
}

// GetCurrentBookings retrieves all current bookings
func (rbs *RoomBookingService) GetCurrentBookings() ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking
	now := time.Now()

	if err := rbs.db.Preload("Guest").Preload("Room").
		Where("check_in <= ? AND check_out > ? AND status = ?",
			now, now, models.BookingStatusConfirmed).
		Find(&bookings).Error; err != nil {
		rbs.logger.Error("failed to get current bookings", zap.Error(err))
		return nil, fmt.Errorf("failed to get current bookings: %w", err)
	}

	return bookings, nil
}

// GetBookingsByDateRange retrieves all bookings within a date range
func (rbs *RoomBookingService) GetBookingsByDateRange(start, end time.Time) ([]models.RoomBooking, error) {
	var bookings []models.RoomBooking

	if err := rbs.db.Preload("Guest").Preload("Room").
		Where("(check_in BETWEEN ? AND ?) OR (check_out BETWEEN ? AND ?) OR (check_in <= ? AND check_out >= ?)",
			start, end, start, end, start, end).
		Find(&bookings).Error; err != nil {
		rbs.logger.Error("failed to get bookings by date range", zap.Error(err))
		return nil, fmt.Errorf("failed to get bookings by date range: %w", err)
	}

	return bookings, nil
}

// CheckGuestIn marks a booking as checked in
func (rbs *RoomBookingService) CheckGuestIn(bookingID uint) error {
	var booking models.RoomBooking

	if err := rbs.db.First(&booking, bookingID).Error; err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	booking.Status = models.BookingStatusCheckedIn
	booking.ActualCheckIn = time.Now()

	if err := rbs.db.Save(&booking).Error; err != nil {
		rbs.logger.Error("failed to check guest in", zap.Uint("bookingID", bookingID), zap.Error(err))
		return fmt.Errorf("failed to check guest in: %w", err)
	}

	return nil
}

// CheckGuestOut marks a booking as checked out
func (rbs *RoomBookingService) CheckGuestOut(bookingID uint) error {
	var booking models.RoomBooking

	if err := rbs.db.First(&booking, bookingID).Error; err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	booking.Status = models.BookingStatusCompleted
	booking.ActualCheckOut = time.Now()

	if err := rbs.db.Save(&booking).Error; err != nil {
		rbs.logger.Error("failed to check guest out", zap.Uint("bookingID", bookingID), zap.Error(err))
		return fmt.Errorf("failed to check guest out: %w", err)
	}

	return nil
}
