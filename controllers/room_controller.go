package controllers

import (
	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"github.com/IamMaheshGurung/privateOnsenBooking/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RoomController handles room-related HTTP requests
type RoomController struct {
	Service *services.RoomBookingService
	Logger  *zap.Logger
}

// NewRoomController creates a new instance of RoomController
func NewRoomController(service *services.RoomBookingService, logger *zap.Logger) *RoomController {
	return &RoomController{
		Service: service,
		Logger:  logger,
	}
}

// GetAllRooms returns all rooms
// GET /api/rooms
func (ctrl *RoomController) GetAllRooms(c *fiber.Ctx) error {

	rooms, err := ctrl.Service.GetAllRooms()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get rooms: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    rooms,
	})
}

// GetRoomByID returns a specific room
// GET /api/rooms/:id
func (ctrl *RoomController) GetRoomByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid room ID",
		})
	}

	room, err := ctrl.Service.GetRoomByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    room,
	})
}

// GetAvailableRooms returns available rooms for a date range
// GET /api/rooms/available?check_in=YYYY-MM-DD&check_out=YYYY-MM-DD
func (ctrl *RoomController) GetAvailableRooms(c *fiber.Ctx) error {
	// Parse check-in and check-out dates
	checkInStr := c.Query("check_in")
	checkOutStr := c.Query("check_out")

	if checkInStr == "" || checkOutStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Check-in and check-out dates are required",
		})
	}

	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid check-in date format. Use YYYY-MM-DD",
		})
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid check-out date format. Use YYYY-MM-DD",
		})
	}

	// Validate dates
	if checkIn.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Check-in date cannot be in the past",
		})
	}

	if checkOut.Before(checkIn) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Check-out date must be after check-in date",
		})
	}

	rooms, err := ctrl.Service.GetAvailableRooms(checkIn, checkOut)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get available rooms: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    rooms,
	})
}

// Admin Routes

// CreateRoom creates a new room
// POST /api/admin/rooms
func (ctrl *RoomController) CreateRoom(c *fiber.Ctx) error {
	var room models.Room

	if err := c.BodyParser(&room); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON: " + err.Error(),
		})
	}

	// Validate room data
	if room.RoomNo == "" || room.Type == "" || room.PricePerNight <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room number, type, and price are required",
		})
	}

	err := ctrl.Service.CreateRoom(room)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create room: " + err.Error(),
		})
	}

	message := "Room created successfully"

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    message,
	})
}

// UpdateRoom updates a room
// PUT /api/admin/rooms/:id
func (ctrl *RoomController) UpdateRoom(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid room ID",
		})
	}

	var room models.Room
	if err := c.BodyParser(&room); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON: " + err.Error(),
		})
	}

	room.ID = uint(id)
	updatedRoom, err := ctrl.Service.UpdateRoom(room)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update room: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    updatedRoom,
	})
}

// DeleteRoom deletes a room
// DELETE /api/admin/rooms/:id
func (ctrl *RoomController) DeleteRoom(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid room ID",
		})
	}

	if err := ctrl.Service.DeleteRoom(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete room: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Room deleted successfully",
	})
}
