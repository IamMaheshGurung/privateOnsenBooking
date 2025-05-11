package controllers

import (
	"fmt"
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

	if err := ctrl.Service.DeactivateRoom(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete room: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Room deleted successfully",
	})
}

// GetRoomDetails returns details for a specific room
func (rc *RoomController) GetRoomDetails(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid room ID",
		})
	}

	room, err := rc.Service.GetRoomByID(uint(id))
	if err != nil {
		rc.Logger.Error("Failed to get room details", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get room details",
		})
	}

	return c.Render("rooms/detail", fiber.Map{
		"Room": room,
	})
}

// GetRoomTypes returns all available room types
func (rc *RoomController) GetRoomTypes(c *fiber.Ctx) error {
	rooms, err := rc.Service.GetAllRooms()
	if err != nil {
		rc.Logger.Error("Failed to get room types", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get room types",
		})
	}

	// Extract unique room types
	typeMap := make(map[string]bool)
	var types []string

	for _, room := range rooms {
		if !typeMap[room.Type] {
			typeMap[room.Type] = true
			types = append(types, room.Type)
		}
	}

	return c.JSON(fiber.Map{
		"types": types,
	})
}

func (rc *RoomController) PreviewRooms(c *fiber.Ctx) error {
	rooms, err := rc.Service.GetAllRooms()
	if err != nil {
		rc.Logger.Error("Failed to get rooms for preview", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get room preview",
		})
	}

	// Limit to 3 rooms for preview
	previewRooms := rooms
	if len(rooms) > 3 {
		previewRooms = rooms[:3]
	}

	return c.Render("partials/room_preview", fiber.Map{
		"Rooms": previewRooms,
	})
}

// AdminListRooms displays all rooms in the admin panel
func (rc *RoomController) AdminListRooms(c *fiber.Ctx) error {
	rooms, err := rc.Service.GetAllRooms()
	if err != nil {
		rc.Logger.Error("Failed to list rooms for admin", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).Render("admin/error", fiber.Map{
			"Message": "Failed to load rooms",
			"Error":   err.Error(),
		})
	}

	return c.Render("admin/rooms/index", fiber.Map{
		"Rooms": rooms,
		"Title": "Room Management",
	})
}

// ShowAddRoomForm displays the form to add a new room
func (rc *RoomController) ShowAddRoomForm(c *fiber.Ctx) error {
	return c.Render("admin/rooms/form", fiber.Map{
		"Title":  "Add New Room",
		"Action": "/admin/rooms/add",
		"Room":   models.Room{},
	})
}

// AddRoom processes the form submission to add a new room
func (rc *RoomController) AddRoom(c *fiber.Ctx) error {
	room := new(models.Room)

	// Parse the form
	if err := c.BodyParser(room); err != nil {
		rc.Logger.Error("Failed to parse room form data", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid room data",
		})
	}

	// Handle file upload for room image
	file, err := c.FormFile("image")
	if err == nil {
		// Save the file
		filename := fmt.Sprintf("room_%d_%s", time.Now().Unix(), file.Filename)
		if err := c.SaveFile(file, fmt.Sprintf("./static/uploads/%s", filename)); err != nil {
			rc.Logger.Error("Failed to save room image", zap.Error(err))
		} else {
			room.ImageURL = fmt.Sprintf("/static/uploads/%s", filename)
		}
	}

	// Create the room in the database
	if err := rc.Service.CreateRoom(*room); err != nil {
		rc.Logger.Error("Failed to create room", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create room",
		})
	}

	// Check if this is an HTMX request
	if c.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", "/admin/rooms")
		return c.SendStatus(fiber.StatusCreated)
	}

	return c.Redirect("/admin/rooms")
}

// ShowEditRoomForm displays the form to edit an existing room
func (rc *RoomController) ShowEditRoomForm(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid room ID",
		})
	}

	room, err := rc.Service.GetRoomByID(uint(id))
	if err != nil {
		rc.Logger.Error("Failed to get room for editing", zap.Error(err))
		return c.Status(fiber.StatusNotFound).Render("admin/error", fiber.Map{
			"Message": "Room not found",
			"Error":   err.Error(),
		})
	}

	return c.Render("admin/rooms/form", fiber.Map{
		"Title":  "Edit Room",
		"Action": fmt.Sprintf("/admin/rooms/%d/edit", id),
		"Room":   room,
	})
}

// EditRoom processes the form submission to update an existing room
func (rc *RoomController) EditRoom(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid room ID",
		})
	}

	// Get the existing room
	existingRoom, err := rc.Service.GetRoomByID(uint(id))
	if err != nil {
		rc.Logger.Error("Failed to get room for updating", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
		})
	}

	// Parse form data into a new room
	updatedRoom := new(models.Room)
	if err := c.BodyParser(updatedRoom); err != nil {
		rc.Logger.Error("Failed to parse room update form", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Set the ID to ensure we update the right record
	updatedRoom.ID = uint(id)

	// Handle file upload if a new image is provided
	file, err := c.FormFile("image")
	if err == nil {
		// Save the new image
		filename := fmt.Sprintf("room_%d_%s", time.Now().Unix(), file.Filename)
		if err := c.SaveFile(file, fmt.Sprintf("./static/uploads/%s", filename)); err != nil {
			rc.Logger.Error("Failed to save updated room image", zap.Error(err))
		} else {
			updatedRoom.ImageURL = fmt.Sprintf("/static/uploads/%s", filename)
		}
	} else {
		// Keep the existing image if no new one is uploaded
		updatedRoom.ImageURL = existingRoom.ImageURL
	}

	// Update the room
	result, err := rc.Service.UpdateRoom(*updatedRoom)
	if err != nil {
		rc.Logger.Error("Failed to update room", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update room",
		})
	}

	// Check if this is an HTMX request
	if c.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", "/admin/rooms")
		return c.SendStatus(fiber.StatusOK)
	}

	return c.Render("admin/rooms_updated", fiber.Map{
		"Room":    result,
		"Title":   "Room Updated",
		"Message": "Room updated successfully",
	})
}
