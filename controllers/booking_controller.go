package controllers

import (
	
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/IamMaheshGurung/privateOnsenBooking/services"
)

type RoomBookingController struct {
	Service *services.RoomBookingService
}

func NewRoomBookingController(service *services.RoomBookingService) *RoomBookingController {
	return &RoomBookingController{Service: service}
}

// POST /book
func (ctrl *RoomBookingController) BookRoom(c *fiber.Ctx) error {
	guestID, err := strconv.Atoi(c.FormValue("guest_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid guest ID")
	}
	roomID, err := strconv.Atoi(c.FormValue("room_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid room ID")
	}

	checkIn, err := time.Parse("2006-01-02", c.FormValue("check_in"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid check-in date")
	}
	checkOut, err := time.Parse("2006-01-02", c.FormValue("check_out"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid check-out date")
	}

    // Check if the room is available
    available, err := ctrl.Service.IsRoomAvailable(uint(roomID), checkIn, checkOut)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
    }
    if !available {
        return c.Status(fiber.StatusConflict).SendString("Room is not available for the selected dates")
    }
    // Create the booking
	err = ctrl.Service.CreateBooking(uint(guestID), uint(roomID), checkIn, checkOut)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Room booked successfully"})
}

// GET /check-availability?room_id=1&check_in=2025-05-01&check_out=2025-05-03
func (ctrl *RoomBookingController) CheckAvailability(c *fiber.Ctx) error {
	roomID, err := strconv.Atoi(c.Query("room_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid room ID")
	}
	checkIn, err := time.Parse("2006-01-02", c.Query("check_in"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid check-in date")
	}
	checkOut, err := time.Parse("2006-01-02", c.Query("check_out"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid check-out date")
	}

	available, err := ctrl.Service.IsRoomAvailable(uint(roomID), checkIn, checkOut)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(fiber.Map{"available": available})
}

// PUT /update/:id
func (ctrl *RoomBookingController) UpdateBooking(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid booking ID")
	}
	checkIn, err := time.Parse("2006-01-02", c.FormValue("check_in"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid check-in date")
	}
	checkOut, err := time.Parse("2006-01-02", c.FormValue("check_out"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid check-out date")
	}

	err = ctrl.Service.UpdateBookingByID(uint(bookingID), checkIn, checkOut)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(fiber.Map{"message": "Booking updated successfully"})
}

// GET /bookings?date=2025-05-01
func (ctrl *RoomBookingController) GetBookingsByDate(c *fiber.Ctx) error {
	checkIn, err := time.Parse("2006-01-02", c.Query("date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid date format")
	}

	bookings, err := ctrl.Service.GetBookingByCheckInDate(checkIn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(bookings)
}

// DELETE /cancel/:id
func (ctrl *RoomBookingController) CancelBooking(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid booking ID")
	}

	err = ctrl.Service.CancelBookingByID(uint(bookingID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(fiber.Map{"message": "Booking cancelled successfully"})
}

