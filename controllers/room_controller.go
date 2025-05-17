package controllers

import (
	"fmt"
	"strconv"
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

// GetAllRoomsPage returns all rooms, optionally filtered by type
func (rc *RoomController) GetAllRoomsPage(c *fiber.Ctx) error {
	roomType := c.Query("typee")

	rc.Logger.Info("Fetching rooms", zap.String("type", roomType))

	var rooms []*models.Room
	var err error

	if roomType != "" {
		rooms, err = rc.Service.GetRoomByType(roomType)
	} else {
		rooms, err = rc.Service.GetAllRooms()
	}

	if err != nil {
		rc.Logger.Error("Failed to get rooms", zap.Error(err))
		// Send a simple error message for now instead of trying to render a template
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching rooms: " + err.Error())
	}

	// Debug log to see what data we're trying to render
	rc.Logger.Info("Rendering rooms page",
		zap.Int("room_count", len(rooms)),
		zap.String("template", "rooms/index"))

	// Try rendering with error handling
	err = c.Render("rooms/index", fiber.Map{
		"Title":       "Accommodations | Kwangdi Pahuna Ghar",
		"Description": "Explore our comfortable and authentic Nepali accommodations",
		"CurrentYear": time.Now().Year(),
		"Rooms":       rooms,
		"RoomType":    roomType,
	})

	if err != nil {
		rc.Logger.Error("Template rendering error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Template error: " + err.Error())
	}

	return nil
}

// Add this to your RoomController and route it to /api/rooms/:id/quick-view
func (rc *RoomController) GetRoomQuickView(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid room ID"})
	}

	room, err := rc.Service.GetRoomByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	// Just return the HTML for the modal
	return c.Render("partials/room_quick_view", fiber.Map{
		"Room": room,
	}, "")
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

// GetAvailableRooms returns rooms available for a specific date range and guest count
// GET /api/rooms/available or /rooms/availability
func (rc *RoomController) GetAvailableRooms(c *fiber.Ctx) error {
	// Parse request parameters
	checkIn := c.Query("check_in")
	checkOut := c.Query("check_out")
	guestsStr := c.Query("guests")

	rc.Logger.Info("Checking room availability",
		zap.String("check_in", checkIn),
		zap.String("check_out", checkOut),
		zap.String("guests", guestsStr))

	// Validate check-in and check-out dates
	if checkIn == "" || checkOut == "" {
		if c.Get("HX-Request") == "true" {
			// If it's an HTMX request, return just the content with error message
			return c.Render("partials/rooms_grid_error", fiber.Map{
				"Message": "Please select both check-in and check-out dates",
			}, "")
		}

		// If it's a regular request, redirect to the rooms page
		return c.Redirect("/rooms")
	}

	// Parse dates
	checkInDate, err := time.Parse("2006-01-02", checkIn)
	if err != nil {
		rc.Logger.Error("Invalid check-in date", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid check-in date format. Use YYYY-MM-DD",
		})
	}

	checkOutDate, err := time.Parse("2006-01-02", checkOut)
	if err != nil {
		rc.Logger.Error("Invalid check-out date", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid check-out date format. Use YYYY-MM-DD",
		})
	}

	// Make sure check-out is after check-in
	if checkOutDate.Before(checkInDate) || checkOutDate.Equal(checkInDate) {
		rc.Logger.Error("Invalid date range",
			zap.Time("check_in", checkInDate),
			zap.Time("check_out", checkOutDate))

		if c.Get("HX-Request") == "true" {
			return c.Render("partials/rooms_grid_error", fiber.Map{
				"Message": "Check-out date must be after check-in date",
			}, "")
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Check-out date must be after check-in date",
		})
	}

	// Parse guest count (default to 1 if not provided or invalid)
	guests := 1
	if guestsStr != "" {
		guests, err = strconv.Atoi(guestsStr)
		if err != nil || guests < 1 {
			guests = 1
		}
	}

	// Get available rooms from service
	rooms, err := rc.Service.GetAvailableRooms(checkInDate, checkOutDate, guests)
	if err != nil {
		rc.Logger.Error("Failed to get available rooms", zap.Error(err))

		if c.Get("HX-Request") == "true" {
			return c.Render("partials/rooms_grid_error", fiber.Map{
				"Message": "Error finding available rooms",
			}, "")
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get available rooms: " + err.Error(),
		})
	}

	// If it's an HTMX request, return just the room grid
	if c.Get("HX-Request") == "true" {
		return c.Render("partials/rooms_grid", fiber.Map{
			"Rooms":    rooms,
			"CheckIn":  checkIn,
			"CheckOut": checkOut,
			"Guests":   guests,
		}, "")
	}

	// For a full page request, render the complete page
	return c.Render("rooms/index", fiber.Map{
		"Title":       "Available Rooms | Kwangdi Pahuna Ghar",
		"Description": "Available rooms for your selected dates",
		"CurrentYear": time.Now().Year(),
		"Rooms":       rooms,
		"CheckIn":     checkIn,
		"CheckOut":    checkOut,
		"Guests":      guests,
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
