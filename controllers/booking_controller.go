package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"github.com/IamMaheshGurung/privateOnsenBooking/services"
	"github.com/IamMaheshGurung/privateOnsenBooking/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// BookingController handles booking-related HTTP requests
type BookingController struct {
	RoomService   *services.RoomBookingService
	GuestService  *services.GuestService
	EmailService  *services.EmailService
	Logger        *zap.Logger
	MinStayLength int // Minimum number of nights
	MaxStayLength int // Maximum number of nights
}

// NewBookingController creates a new instance of BookingController
func NewBookingController(
	roomService *services.RoomBookingService,
	guestService *services.GuestService,
	emailService *services.EmailService,
	logger *zap.Logger,
) *BookingController {
	return &BookingController{
		RoomService:   roomService,
		GuestService:  guestService,
		EmailService:  emailService,
		Logger:        logger,
		MinStayLength: 1,  // Default minimum: 1 night
		MaxStayLength: 14, // Default maximum: 14 nights
	}
}

// CheckAvailability checks if a room is available
// GET /api/bookings/check?room_id=1&check_in=2023-09-01&check_out=2023-09-05
func (ctrl *BookingController) CheckAvailability(c *fiber.Ctx) error {
	ctrl.Logger.Info("CheckAvailability request received",
		zap.String("path", c.Path()),
		zap.String("queryParams", string(c.Request().URI().QueryString())))

	roomIDStr := c.Query("room_id")
	checkInStr := c.Query("check_in")
	checkOutStr := c.Query("check_out")

	// Validate input
	if roomIDStr == "" || checkInStr == "" || checkOutStr == "" {
		ctrl.Logger.Warn("Missing required parameters",
			zap.String("roomID", roomIDStr),
			zap.String("checkIn", checkInStr),
			zap.String("checkOut", checkOutStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Room ID, check-in, and check-out dates are required",
		})
	}

	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		ctrl.Logger.Warn("Invalid room ID", zap.String("roomID", roomIDStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid room ID",
		})
	}

	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		ctrl.Logger.Warn("Invalid check-in date format", zap.String("checkIn", checkInStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid check-in date format. Use YYYY-MM-DD",
		})
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		ctrl.Logger.Warn("Invalid check-out date format", zap.String("checkOut", checkOutStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid check-out date format. Use YYYY-MM-DD",
		})
	}

	// Validate booking parameters
	if err := ctrl.validateBookingDates(checkIn, checkOut); err != nil {
		ctrl.Logger.Warn("Date validation failed",
			zap.Time("checkIn", checkIn),
			zap.Time("checkOut", checkOut),
			zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Check if room exists first
	_, err = ctrl.RoomService.GetRoomByID(uint(roomID))
	if err != nil {
		ctrl.Logger.Warn("Room not found", zap.Uint64("roomID", roomID))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Room not found",
		})
	}

	// Check room availability
	available, err := ctrl.RoomService.IsRoomAvailable(uint(roomID), checkIn, checkOut)
	if err != nil {
		ctrl.Logger.Error("Failed to check availability",
			zap.Uint64("roomID", roomID),
			zap.Time("checkIn", checkIn),
			zap.Time("checkOut", checkOut),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check availability: " + err.Error(),
		})
	}

	// Calculate price if available
	var price float64
	var nightCount int
	if available {
		room, err := ctrl.RoomService.GetRoomByID(uint(roomID))
		if err == nil {
			nightCount = int(checkOut.Sub(checkIn).Hours() / 24)
			price = room.PricePerNight * float64(nightCount)
		}
	}

	ctrl.Logger.Info("CheckAvailability request completed",
		zap.Uint64("roomID", roomID),
		zap.Time("checkIn", checkIn),
		zap.Time("checkOut", checkOut),
		zap.Bool("available", available))

	return c.JSON(fiber.Map{
		"success":    true,
		"available":  available,
		"nightCount": nightCount,
		"totalPrice": price,
	})
}

// GetAvailableRooms returns all rooms available for the specified dates
// GET /api/bookings/available?check_in=2023-09-01&check_out=2023-09-05
func (ctrl *BookingController) GetAvailableRooms(c *fiber.Ctx) error {
	ctrl.Logger.Info("GetAvailableRooms request received",
		zap.String("path", c.Path()),
		zap.String("queryParams", string(c.Request().URI().QueryString())))

	checkInStr := c.Query("check_in")
	checkOutStr := c.Query("check_out")
	guestStr := c.Query("guests")

	// Validate input
	if checkInStr == "" || checkOutStr == "" {
		ctrl.Logger.Warn("Missing required parameters",
			zap.String("checkIn", checkInStr),
			zap.String("checkOut", checkOutStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Check-in and check-out dates are required",
		})
	}

	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		ctrl.Logger.Warn("Invalid check-in date format", zap.String("checkIn", checkInStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid check-in date format. Use YYYY-MM-DD",
		})
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		ctrl.Logger.Warn("Invalid check-out date format", zap.String("checkOut", checkOutStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid check-out date format. Use YYYY-MM-DD",
		})
	}

	// Validate booking parameters
	if err := ctrl.validateBookingDates(checkIn, checkOut); err != nil {
		ctrl.Logger.Warn("Date validation failed",
			zap.Time("checkIn", checkIn),
			zap.Time("checkOut", checkOut),
			zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Get available rooms
	rooms, err := ctrl.RoomService.GetAvailableRooms(checkIn, checkOut, guestStr)
	if err != nil {
		ctrl.Logger.Error("Failed to get available rooms",
			zap.Time("checkIn", checkIn),
			zap.Time("checkOut", checkOut),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get available rooms: " + err.Error(),
		})
	}

	// Calculate stay length in nights
	nightCount := int(checkOut.Sub(checkIn).Hours() / 24)

	// Add total price for the stay to each room
	var roomsWithPrice []fiber.Map
	for _, room := range rooms {
		roomsWithPrice = append(roomsWithPrice, fiber.Map{
			"id":              room.ID,
			"room_no":         room.RoomNo,
			"type":            room.Type,
			"capacity":        room.Capacity,
			"price_per_night": room.PricePerNight,
			"description":     room.Description,
			"amenities":       room.Amenities,
			"image_url":       room.ImageURL,
			"total_price":     room.PricePerNight * float64(nightCount),
			"night_count":     nightCount,
		})
	}

	ctrl.Logger.Info("GetAvailableRooms request completed",
		zap.Time("checkIn", checkIn),
		zap.Time("checkOut", checkOut),
		zap.Int("roomCount", len(rooms)))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    roomsWithPrice,
	})
}

// CreateBooking creates a new booking
// POST /api/bookings
func (ctrl *BookingController) CreateBooking(c *fiber.Ctx) error {
	ctrl.Logger.Info("CreateBooking request received")

	var bookingData struct {
		GuestName       string  `json:"guest_name"`
		GuestEmail      string  `json:"guest_email"`
		GuestPhone      string  `json:"guest_phone"`
		RoomID          uint    `json:"room_id"`
		CheckIn         string  `json:"check_in"`
		CheckOut        string  `json:"check_out"`
		SpecialRequests string  `json:"special_requests"`
		TotalPrice      float64 `json:"total_price"`
	}

	// Parse request body
	if err := c.BodyParser(&bookingData); err != nil {
		ctrl.Logger.Warn("Cannot parse JSON", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Cannot parse JSON: " + err.Error(),
		})
	}

	// Validate required fields
	if bookingData.GuestName == "" || bookingData.GuestEmail == "" ||
		bookingData.GuestPhone == "" || bookingData.RoomID == 0 ||
		bookingData.CheckIn == "" || bookingData.CheckOut == "" {
		ctrl.Logger.Warn("Missing required fields", zap.Any("data", bookingData))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "All required fields must be provided",
		})
	}

	// Validate email format
	if !utils.IsValidEmail(bookingData.GuestEmail) {
		ctrl.Logger.Warn("Invalid email format", zap.String("email", bookingData.GuestEmail))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid email format",
		})
	}

	// Parse dates
	checkIn, err := time.Parse("2006-01-02", bookingData.CheckIn)
	if err != nil {
		ctrl.Logger.Warn("Invalid check-in date format", zap.String("checkIn", bookingData.CheckIn))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid check-in date format. Use YYYY-MM-DD",
		})
	}

	checkOut, err := time.Parse("2006-01-02", bookingData.CheckOut)
	if err != nil {
		ctrl.Logger.Warn("Invalid check-out date format", zap.String("checkOut", bookingData.CheckOut))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid check-out date format. Use YYYY-MM-DD",
		})
	}

	// Validate booking dates
	if err := ctrl.validateBookingDates(checkIn, checkOut); err != nil {
		ctrl.Logger.Warn("Date validation failed",
			zap.Time("checkIn", checkIn),
			zap.Time("checkOut", checkOut),
			zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Check if room exists
	room, err := ctrl.RoomService.GetRoomByID(bookingData.RoomID)
	if err != nil {
		ctrl.Logger.Warn("Room not found", zap.Uint("roomID", bookingData.RoomID))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Room not found",
		})
	}

	// Double-check room availability
	available, err := ctrl.RoomService.IsRoomAvailable(bookingData.RoomID, checkIn, checkOut)
	if err != nil {
		ctrl.Logger.Error("Failed to check availability", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check room availability: " + err.Error(),
		})
	}

	if !available {
		ctrl.Logger.Warn("Room not available",
			zap.Uint("roomID", bookingData.RoomID),
			zap.Time("checkIn", checkIn),
			zap.Time("checkOut", checkOut))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "The room is no longer available for the selected dates",
		})
	}

	// Calculate correct total price
	nightCount := int(checkOut.Sub(checkIn).Hours() / 24)
	correctPrice := room.PricePerNight * float64(nightCount)

	// Validate price (optional, prevents frontend manipulation)
	if bookingData.TotalPrice > 0 && !utils.IsCloseEnough(bookingData.TotalPrice, correctPrice, 0.01) {
		ctrl.Logger.Warn("Price mismatch",
			zap.Float64("submittedPrice", bookingData.TotalPrice),
			zap.Float64("calculatedPrice", correctPrice))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":      false,
			"error":        "Price calculation mismatch. Please try again.",
			"correctPrice": correctPrice,
		})
	}

	// Create or get guest
	guest, err := ctrl.GuestService.CreateOrGetGuest(
		bookingData.GuestName,
		bookingData.GuestEmail,
		bookingData.GuestPhone,
	)
	if err != nil {
		ctrl.Logger.Error("Failed to create guest", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to process guest information: " + err.Error(),
		})
	}

	// Create a new RoomBooking object
	booking := models.RoomBooking{
		GuestID:         guest.ID,
		RoomID:          bookingData.RoomID,
		CheckIn:         checkIn,
		CheckOut:        checkOut,
		Status:          models.BookingStatusConfirmed,
		SpecialRequests: bookingData.SpecialRequests,
		TotalPrice:      correctPrice,
	}

	// Create booking
	createdBooking, err := ctrl.RoomService.CreateBooking(booking.RoomID, booking.GuestID, booking.CheckIn, booking.CheckOut)
	if err != nil {
		ctrl.Logger.Error("Failed to create booking", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create booking: " + err.Error(),
		})
	}

	// Send confirmation email asynchronously
	go func() {
		if err := ctrl.EmailService.SendBookingConfirmation(createdBooking, guest, room); err != nil {
			ctrl.Logger.Error("Failed to send confirmation email",
				zap.Uint("bookingID", createdBooking.ID),
				zap.Error(err))
		}
	}()

	ctrl.Logger.Info("Booking created successfully",
		zap.Uint("bookingID", createdBooking.ID),
		zap.Uint("guestID", guest.ID),
		zap.Uint("roomID", bookingData.RoomID))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Booking created successfully",
		"data": fiber.Map{
			"booking_id":  createdBooking.ID,
			"guest_id":    createdBooking.GuestID,
			"room_id":     createdBooking.RoomID,
			"check_in":    createdBooking.CheckIn,
			"check_out":   createdBooking.CheckOut,
			"total_price": createdBooking.TotalPrice,
			"status":      createdBooking.Status,
			"created_at":  createdBooking.CreatedAt,
		},
	})
}

// GetBookingByID returns a specific booking
// GET /api/bookings/:id
func (ctrl *BookingController) GetBookingByID(c *fiber.Ctx) error {
	bookingID, err := c.ParamsInt("id")
	if err != nil {
		ctrl.Logger.Warn("Invalid booking ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid booking ID",
		})
	}

	// Get booking details
	booking, err := ctrl.RoomService.GetBookingByID(uint(bookingID))
	if err != nil {
		ctrl.Logger.Error("Failed to get booking",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Booking not found: " + err.Error(),
		})
	}

	ctrl.Logger.Info("Booking details retrieved", zap.Int("bookingID", bookingID))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    booking,
	})
}

// CancelBooking cancels a booking
// PUT /api/bookings/:id/cancel
func (ctrl *BookingController) CancelBooking(c *fiber.Ctx) error {
	bookingID, err := c.ParamsInt("id")
	if err != nil {
		ctrl.Logger.Warn("Invalid booking ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid booking ID",
		})
	}

	// Get booking before cancellation for email notification
	booking, err := ctrl.RoomService.GetBookingByID(uint(bookingID))
	if err != nil {
		ctrl.Logger.Error("Failed to get booking for cancellation",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Booking not found: " + err.Error(),
		})
	}

	// Get guest and room for email notification
	guest, err := ctrl.GuestService.GetGuestByID(booking.GuestID)
	if err != nil {
		ctrl.Logger.Error("Failed to get guest for cancellation notification",
			zap.Uint("guestID", booking.GuestID),
			zap.Error(err))
		// Continue with cancellation even if guest retrieval fails
	}

	room, err := ctrl.RoomService.GetRoomByID(booking.RoomID)
	if err != nil {
		ctrl.Logger.Error("Failed to get room for cancellation notification",
			zap.Uint("roomID", booking.RoomID),
			zap.Error(err))
		// Continue with cancellation even if room retrieval fails
	}

	// Cancel the booking
	if err := ctrl.RoomService.CancelBookingByID(uint(bookingID)); err != nil {
		ctrl.Logger.Error("Failed to cancel booking",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to cancel booking: " + err.Error(),
		})
	}

	// Send cancellation email asynchronously
	if guest != nil && room != nil {
		go func() {
			if err := ctrl.EmailService.SendBookingCancellationNotice(booking, guest, room); err != nil {
				ctrl.Logger.Error("Failed to send cancellation email",
					zap.Int("bookingID", bookingID),
					zap.Error(err))
			}
		}()
	}

	ctrl.Logger.Info("Booking cancelled successfully", zap.Int("bookingID", bookingID))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Booking cancelled successfully",
	})
}

// GetGuestBookings returns all bookings for a guest by email
// GET /api/bookings/guest?email=guest@example.com
func (ctrl *BookingController) GetGuestBookings(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		ctrl.Logger.Warn("Missing email parameter")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email parameter is required",
		})
	}

	// Get guest by email
	guest, err := ctrl.GuestService.GetGuestByEmail(email)
	if err != nil {
		ctrl.Logger.Error("Failed to get guest by email",
			zap.String("email", email),
			zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Guest not found with the provided email",
		})
	}

	// Get bookings for guest
	bookings, err := ctrl.GuestService.GetGuestBookingHistory(guest.ID)
	if err != nil {
		ctrl.Logger.Error("Failed to get guest bookings",
			zap.Uint("guestID", guest.ID),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve booking history: " + err.Error(),
		})
	}

	ctrl.Logger.Info("Guest bookings retrieved",
		zap.String("email", email),
		zap.Int("bookingCount", len(bookings)))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bookings,
	})
}

// Admin Routes

// GetAllBookings returns all bookings with filters
// GET /api/admin/bookings?status=confirmed&future=true
func (ctrl *BookingController) GetAllBookings(c *fiber.Ctx) error {
	status := c.Query("status", "")
	futureStr := c.Query("future", "false")

	future := futureStr == "true"

	bookings, err := ctrl.RoomService.GetAllBookings(status, future)
	if err != nil {
		ctrl.Logger.Error("Failed to get bookings",
			zap.String("status", status),
			zap.Bool("future", future),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get bookings: " + err.Error(),
		})
	}

	ctrl.Logger.Info("All bookings retrieved",
		zap.String("status", status),
		zap.Bool("future", future),
		zap.Int("count", len(bookings)))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bookings,
	})
}

// GetBookingsByDate returns bookings for a specific date
// GET /api/admin/bookings/date/:date
func (ctrl *BookingController) GetBookingsByDate(c *fiber.Ctx) error {
	dateStr := c.Params("date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		ctrl.Logger.Warn("Invalid date format", zap.String("date", dateStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid date format. Use YYYY-MM-DD",
		})
	}

	bookings, err := ctrl.RoomService.GetBookingByCheckInDate(date)
	if err != nil {
		ctrl.Logger.Error("Failed to get bookings by date",
			zap.String("date", dateStr),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get bookings: " + err.Error(),
		})
	}

	ctrl.Logger.Info("Bookings by date retrieved",
		zap.String("date", dateStr),
		zap.Int("count", len(bookings)))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bookings,
	})
}

// GetBookingsByDateRange returns bookings within a date range
// GET /api/admin/bookings/range?start=2023-09-01&end=2023-09-30
func (ctrl *BookingController) GetBookingsByDateRange(c *fiber.Ctx) error {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		ctrl.Logger.Warn("Missing date range parameters")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Start and end dates are required",
		})
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		ctrl.Logger.Warn("Invalid start date format", zap.String("start", startStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid start date format. Use YYYY-MM-DD",
		})
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		ctrl.Logger.Warn("Invalid end date format", zap.String("end", endStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid end date format. Use YYYY-MM-DD",
		})
	}

	if end.Before(start) {
		ctrl.Logger.Warn("End date before start date",
			zap.String("start", startStr),
			zap.String("end", endStr))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "End date must be after start date",
		})
	}

	bookings, err := ctrl.RoomService.GetBookingsByDateRange(start, end)
	if err != nil {
		ctrl.Logger.Error("Failed to get bookings by date range",
			zap.String("start", startStr),
			zap.String("end", endStr),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get bookings: " + err.Error(),
		})
	}

	ctrl.Logger.Info("Bookings by date range retrieved",
		zap.String("start", startStr),
		zap.String("end", endStr),
		zap.Int("count", len(bookings)))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bookings,
	})
}

// UpdateBooking updates a booking
// PUT /api/admin/bookings/:id
func (ctrl *BookingController) UpdateBooking(c *fiber.Ctx) error {
	bookingID, err := c.ParamsInt("id")
	if err != nil {
		ctrl.Logger.Warn("Invalid booking ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid booking ID",
		})
	}

	var updateData struct {
		GuestName       string `json:"guest_name"`
		CheckIn         string `json:"check_in"`
		CheckOut        string `json:"check_out"`
		SpecialRequests string `json:"special_requests"`
		Status          string `json:"status"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		ctrl.Logger.Warn("Cannot parse JSON", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Cannot parse JSON: " + err.Error(),
		})
	}

	// Get current booking
	currentBooking, err := ctrl.RoomService.GetBookingByID(uint(bookingID))
	if err != nil {
		ctrl.Logger.Error("Failed to get booking for update",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Booking not found: " + err.Error(),
		})
	}

	// Prepare update data
	var checkIn, checkOut time.Time
	var datesChanged bool

	if updateData.CheckIn != "" {
		checkIn, err = time.Parse("2006-01-02", updateData.CheckIn)
		if err != nil {
			ctrl.Logger.Warn("Invalid check-in date format", zap.String("checkIn", updateData.CheckIn))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid check-in date format. Use YYYY-MM-DD",
			})
		}
		datesChanged = !checkIn.Equal(currentBooking.CheckIn)
	} else {
		checkIn = currentBooking.CheckIn
	}

	if updateData.CheckOut != "" {
		checkOut, err = time.Parse("2006-01-02", updateData.CheckOut)
		if err != nil {
			ctrl.Logger.Warn("Invalid check-out date format", zap.String("checkOut", updateData.CheckOut))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid check-out date format. Use YYYY-MM-DD",
			})
		}
		datesChanged = datesChanged || !checkOut.Equal(currentBooking.CheckOut)
	} else {
		checkOut = currentBooking.CheckOut
	}

	// If dates changed, validate them
	if datesChanged {
		if err := ctrl.validateBookingDates(checkIn, checkOut); err != nil {
			ctrl.Logger.Warn("Date validation failed",
				zap.Time("checkIn", checkIn),
				zap.Time("checkOut", checkOut),
				zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		// Check availability for new dates (excluding current booking)
		available, err := ctrl.RoomService.IsRoomAvailableForUpdate(
			currentBooking.RoomID,
			uint(bookingID),
			checkIn,
			checkOut,
		)
		if err != nil {
			ctrl.Logger.Error("Failed to check availability for update",
				zap.Int("bookingID", bookingID),
				zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to check availability: " + err.Error(),
			})
		}

		if !available {
			ctrl.Logger.Warn("Room not available for new dates",
				zap.Int("bookingID", bookingID),
				zap.Time("checkIn", checkIn),
				zap.Time("checkOut", checkOut))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Room is not available for the new dates",
			})
		}
	}

	// Handle status change if provided
	var status string
	if updateData.Status != "" {
		validStatuses := map[string]bool{
			models.BookingStatusConfirmed: true,
			models.BookingStatusCancelled: true,
			models.BookingStatusCheckedIn: true,
			models.BookingStatusCompleted: true,
		}

		if !validStatuses[updateData.Status] {
			ctrl.Logger.Warn("Invalid booking status", zap.String("status", updateData.Status))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid booking status",
			})
		}

		status = updateData.Status
	} else {
		status = currentBooking.Status
	}

	// Create updated booking object
	updatedBooking := models.RoomBooking{
		ID:              currentBooking.ID,
		GuestID:         currentBooking.GuestID,
		RoomID:          currentBooking.RoomID,
		CheckIn:         checkIn,
		CheckOut:        checkOut,
		Status:          status,
		SpecialRequests: currentBooking.SpecialRequests,
		TotalPrice:      currentBooking.TotalPrice,
	}

	if updateData.SpecialRequests != "" {
		updatedBooking.SpecialRequests = updateData.SpecialRequests
	}

	// If check-in dates changed, update total price
	if datesChanged {
		room, _ := ctrl.RoomService.GetRoomByID(currentBooking.RoomID)
		if room != nil {
			nightCount := int(checkOut.Sub(checkIn).Hours() / 24)
			updatedBooking.TotalPrice = room.PricePerNight * float64(nightCount)
		}
	}

	// Update guest name if provided
	var guest *models.Guest
	if updateData.GuestName != "" {
		guest, err = ctrl.GuestService.GetGuestByID(currentBooking.GuestID)
		if err != nil {
			ctrl.Logger.Error("Failed to get guest for name update",
				zap.Uint("guestID", currentBooking.GuestID),
				zap.Error(err))
		} else {
			_, err = ctrl.GuestService.UpdateGuest(
				guest.ID,
				updateData.GuestName,
				guest.Email,
				guest.Phone,
			)
			if err != nil {
				ctrl.Logger.Error("Failed to update guest name",
					zap.Uint("guestID", guest.ID),
					zap.Error(err))
			}
		}
	}

	// Update booking
	if err := ctrl.RoomService.UpdateBookingByID(updatedBooking.ID, updatedBooking.CheckIn, updatedBooking.CheckOut); err != nil {
		ctrl.Logger.Error("Failed to update booking",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update booking: " + err.Error(),
		})
	}

	// If status changed to checked-in, update actual check-in time
	if status == models.BookingStatusCheckedIn && currentBooking.Status != models.BookingStatusCheckedIn {
		if err := ctrl.RoomService.CheckGuestIn(uint(bookingID)); err != nil {
			ctrl.Logger.Error("Failed to record check-in time",
				zap.Int("bookingID", bookingID),
				zap.Error(err))
			// Continue anyway as the main update succeeded
		}
	}

	// If status changed to completed, update actual check-out time
	if status == models.BookingStatusCompleted && currentBooking.Status != models.BookingStatusCompleted {
		if err := ctrl.RoomService.CheckGuestOut(uint(bookingID)); err != nil {
			ctrl.Logger.Error("Failed to record check-out time",
				zap.Int("bookingID", bookingID),
				zap.Error(err))
			// Continue anyway as the main update succeeded
		}
	}

	ctrl.Logger.Info("Booking updated successfully",
		zap.Int("bookingID", bookingID),
		zap.Bool("datesChanged", datesChanged),
		zap.String("status", status))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Booking updated successfully",
		"data":    updatedBooking,
	})
}

// CheckInGuest marks a guest as checked in
// PUT /api/admin/bookings/:id/check-in
func (ctrl *BookingController) CheckInGuest(c *fiber.Ctx) error {
	bookingID, err := c.ParamsInt("id")
	if err != nil {
		ctrl.Logger.Warn("Invalid booking ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid booking ID",
		})
	}

	// Get current booking to check status
	booking, err := ctrl.RoomService.GetBookingByID(uint(bookingID))
	if err != nil {
		ctrl.Logger.Error("Failed to get booking for check-in",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Booking not found: " + err.Error(),
		})
	}

	// Only confirmed bookings can be checked in
	if booking.Status != models.BookingStatusConfirmed {
		ctrl.Logger.Warn("Cannot check in booking with status",
			zap.Int("bookingID", bookingID),
			zap.String("status", booking.Status))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Only confirmed bookings can be checked in",
		})
	}

	// Check in the guest
	if err := ctrl.RoomService.CheckGuestIn(uint(bookingID)); err != nil {
		ctrl.Logger.Error("Failed to check in guest",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check in guest: " + err.Error(),
		})
	}

	ctrl.Logger.Info("Guest checked in successfully", zap.Int("bookingID", bookingID))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Guest checked in successfully",
	})
}

// CheckOutGuest marks a guest as checked out
// PUT /api/admin/bookings/:id/check-out
func (ctrl *BookingController) CheckOutGuest(c *fiber.Ctx) error {
	bookingID, err := c.ParamsInt("id")
	if err != nil {
		ctrl.Logger.Warn("Invalid booking ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid booking ID",
		})
	}

	// Get current booking to check status
	booking, err := ctrl.RoomService.GetBookingByID(uint(bookingID))
	if err != nil {
		ctrl.Logger.Error("Failed to get booking for check-out",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Booking not found: " + err.Error(),
		})
	}

	// Only checked-in bookings can be checked out
	if booking.Status != models.BookingStatusCheckedIn {
		ctrl.Logger.Warn("Cannot check out booking with status",
			zap.Int("bookingID", bookingID),
			zap.String("status", booking.Status))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Only checked-in bookings can be checked out",
		})
	}

	// Check out the guest
	if err := ctrl.RoomService.CheckGuestOut(uint(bookingID)); err != nil {
		ctrl.Logger.Error("Failed to check out guest",
			zap.Int("bookingID", bookingID),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check out guest: " + err.Error(),
		})
	}

	ctrl.Logger.Info("Guest checked out successfully", zap.Int("bookingID", bookingID))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Guest checked out successfully",
	})
}

// Helper function to validate booking dates
func (ctrl *BookingController) validateBookingDates(checkIn, checkOut time.Time) error {
	// Normalize dates to start of day for comparison
	today := time.Now().Truncate(24 * time.Hour)
	checkInDate := checkIn.Truncate(24 * time.Hour)
	checkOutDate := checkOut.Truncate(24 * time.Hour)

	// Check-in must not be in the past
	if checkInDate.Before(today) {
		return fmt.Errorf("check-in date cannot be in the past")
	}

	// Check-out must be after check-in
	if !checkOutDate.After(checkInDate) {
		return fmt.Errorf("check-out date must be after check-in date")
	}

	// Calculate nights
	nights := int(checkOutDate.Sub(checkInDate).Hours() / 24)

	// Minimum stay length
	if nights < ctrl.MinStayLength {
		return fmt.Errorf("minimum stay is %d night(s)", ctrl.MinStayLength)
	}

	// Maximum stay length
	if nights > ctrl.MaxStayLength {
		return fmt.Errorf("maximum stay is %d nights", ctrl.MaxStayLength)
	}

	return nil
}

// ShowBookingForm displays the booking form
func (bc *BookingController) ShowBookingForm(c *fiber.Ctx) error {
	roomID := c.Query("room_id")
	checkIn := c.Query("check_in")
	checkOut := c.Query("check_out")

	return c.Render("booking/form", fiber.Map{
		"RoomID":   roomID,
		"CheckIn":  checkIn,
		"CheckOut": checkOut,
	})
}

// ShowConfirmation displays the booking confirmation page
func (bc *BookingController) ShowConfirmation(c *fiber.Ctx) error {
	bookingID := c.Query("booking_id")
	if bookingID == "" {
		return c.Redirect("/rooms")
	}

	id, err := strconv.ParseUint(bookingID, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("error", fiber.Map{
			"Message": "Invalid booking ID",
		})
	}

	booking, err := bc.RoomService.GetBookingByID(uint(id))
	if err != nil {
		bc.Logger.Error("Failed to get booking details", zap.Error(err))
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"Message": "Booking not found",
		})
	}

	return c.Render("booking/confirmation", fiber.Map{
		"Booking": booking,
	})
}

// ShowLookupForm displays the form to look up a booking
func (ctrl *BookingController) ShowLookupForm(c *fiber.Ctx) error {
	return c.Render("booking/lookup_form", fiber.Map{})
}

// LookupBooking processes the booking lookup request
func (ctrl *BookingController) LookupBooking(c *fiber.Ctx) error {
	email := c.FormValue("email")
	bookingCode := c.FormValue("booking_code")

	if email == "" || bookingCode == "" {
		return c.Status(fiber.StatusBadRequest).Render("partials/lookup_error", fiber.Map{
			"Message": "Please provide both email and booking code",
		})
	}

	booking, err := ctrl.RoomService.GetBookingByEmailAndCode(email, bookingCode)
	if err != nil {
		ctrl.Logger.Error("Failed to look up booking", zap.Error(err))
		return c.Status(fiber.StatusNotFound).Render("partials/lookup_error", fiber.Map{
			"Message": "Booking not found with the provided details",
		})
	}

	return c.Render("partials/booking_details", fiber.Map{
		"Booking": booking,
	})
}

// ShowBookingDetails displays the details of a booking
func (ctrl *BookingController) ShowBookingDetails(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid booking ID",
		})
	}

	booking, err := ctrl.RoomService.GetBookingByID(uint(id))
	if err != nil {
		ctrl.Logger.Error("Failed to get booking details", zap.Error(err))
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"Message": "Booking not found",
		})
	}

	return c.Render("booking/details", fiber.Map{
		"Booking": booking,
	})
}

// CancelBookingByGuest allows a guest to cancel their booking
func (ctrl *BookingController) CancelBookingByGuest(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid booking ID",
		})
	}

	// Verify the cancellation is authorized (e.g., by checking email and booking code)
	email := c.FormValue("email")
	bookingCode := c.FormValue("booking_code")

	if email == "" || bookingCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing authentication details",
		})
	}

	// Check if the provided details match the booking
	authorized, err := ctrl.RoomService.VerifyBookingOwnership(uint(id), email, bookingCode)
	if err != nil || !authorized {
		ctrl.Logger.Error("Unauthorized cancellation attempt", zap.Error(err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authorized to cancel this booking",
		})
	}

	if err := ctrl.RoomService.CancelBooking(uint(id)); err != nil {
		ctrl.Logger.Error("Failed to cancel booking", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to cancel booking",
		})
	}

	// For HTMX: Show cancellation successful message
	return c.Render("partials/cancellation_success", fiber.Map{})
}
