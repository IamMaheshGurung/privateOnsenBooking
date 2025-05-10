package models

import (
	"time"
)

// Guest represents a hotel guest
type Guest struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"not null;unique"`
	Phone     string    `json:"phone" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Room represents a hotel room
type Room struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	RoomNo        string    `json:"room_no" gorm:"not null;unique"`
	Type          string    `json:"type" gorm:"not null"`      // Standard, Deluxe, Suite, etc.
	Capacity      int       `json:"capacity" gorm:"default:2"` // Number of guests
	PricePerNight float64   `json:"price_per_night" gorm:"not null"`
	Status        string    `json:"status" gorm:"default:'active'"` // Available, Booked, Maintenance
	Description   string    `json:"description"`                    // Room description
	Amenities     string    `json:"amenities"`                      // Comma-separated list or JSON
	ImageURL      string    `json:"image_url"`                      // Main room image URL
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RoomBooking represents a hotel room booking
type RoomBooking struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	GuestID            uint      `json:"guest_id"`
	Guest              Guest     `json:"guest" gorm:"foreignKey:GuestID"`
	RoomID             uint      `json:"room_id"`
	Room               Room      `json:"room" gorm:"foreignKey:RoomID"`
	CheckIn            time.Time `json:"check_in" gorm:"not null"`
	CheckOut           time.Time `json:"check_out" gorm:"not null"`
	ActualCheckIn      time.Time `json:"actual_check_in"`
	ActualCheckOut     time.Time `json:"actual_check_out"`
	CancellationFee    float64   `json:"cancellation_fee"`                  // Cancellation fee if applicable
	CancellationReason string    `json:"cancellation_reason"`               // Reason for cancellation if applicable
	CancelledAt        time.Time `json:"cancelled_at"`                      // Timestamp of cancellation
	ReferenceNumber    string    `json:"reference"`                         // Reference number for the booking
	Status             string    `json:"status" gorm:"default:'confirmed'"` // Status as string instead of bool
	SpecialRequests    string    `json:"special_requests"`                  // Any special guest requests
	TotalPrice         float64   `json:"total_price"`                       // Total price for the stay
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// OnsenBooking represents a private onsen booking
type OnsenBooking struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	GuestID     uint        `json:"guest_id"`
	Guest       Guest       `json:"guest" gorm:"foreignKey:GuestID"`
	RoomID      uint        `json:"room_id" gorm:"not null"`
	Room        Room        `json:"room" gorm:"foreignKey:RoomID"`
	BookingID   uint        `json:"booking_id"` // Reference to the related RoomBooking
	RoomBooking RoomBooking `json:"room_booking" gorm:"foreignKey:BookingID"`
	Date        time.Time   `json:"date" gorm:"not null"`              // Date of onsen booking
	TimeSlot    string      `json:"time_slot" gorm:"not null"`         // Time slot (e.g., "18:00-19:00")
	Status      string      `json:"status" gorm:"default:'confirmed'"` // Status of the onsen booking
	Price       float64     `json:"price"`                             // Price for onsen booking
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// Booking status constants
const (
	BookingStatusConfirmed  = "confirmed"
	BookingStatusCancelled  = "cancelled"
	BookingStatusCheckedIn  = "checked_in"
	BookingStatusCheckedOut = "checked_out"
	BookingStatusCompleted  = "completed"
	BookingStatusRejected   = "rejected"
)
