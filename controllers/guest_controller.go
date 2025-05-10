package controllers

import (
	"strconv"

	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"github.com/IamMaheshGurung/privateOnsenBooking/services"
	"github.com/IamMaheshGurung/privateOnsenBooking/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// GuestController handles guest-related HTTP requests
type GuestController struct {
	Service *services.GuestService
	Logger  *zap.Logger
}

// NewGuestController creates a new instance of GuestController
func NewGuestController(service *services.GuestService, logger *zap.Logger) *GuestController {
	return &GuestController{
		Service: service,
		Logger:  logger,
	}
}

// RegisterGuest registers a new guest
func (ctrl *GuestController) RegisterGuest(c *fiber.Ctx) error {
	var guest models.Guest

	if err := c.BodyParser(&guest); err != nil {
		ctrl.Logger.Warn("Failed to parse guest registration", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
	}

	// Validate required fields
	if guest.Name == "" || guest.Email == "" || guest.Phone == "" {
		ctrl.Logger.Warn("Missing required fields for guest registration")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name, email, and phone are required",
		})
	}

	// Validate email format
	if !utils.IsValidEmail(guest.Email) {
		ctrl.Logger.Warn("Invalid email format", zap.String("email", guest.Email))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid email format",
		})
	}

	// Create guest
	createdGuest, err := ctrl.Service.CreateGuest(&guest)
	if err != nil {
		ctrl.Logger.Error("Failed to create guest", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to register guest: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Guest registered successfully",
		"data":    createdGuest,
	})
}

// GetGuestProfile returns the guest profile
func (ctrl *GuestController) GetGuestProfile(c *fiber.Ctx) error {
	// Get guest ID from authenticated user
	guestID, ok := c.Locals("guestID").(uint)
	if !ok {
		ctrl.Logger.Warn("Guest ID not found in request context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Get guest details
	guest, err := ctrl.Service.GetGuestByID(guestID)
	if err != nil {
		ctrl.Logger.Error("Failed to get guest profile", zap.Uint("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Guest not found: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    guest,
	})
}

// UpdateGuestProfile updates the guest profile
func (ctrl *GuestController) UpdateGuestProfile(c *fiber.Ctx) error {
	// Get guest ID from authenticated user
	guestID, ok := c.Locals("guestID").(uint)
	if !ok {
		ctrl.Logger.Warn("Guest ID not found in request context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Parse request body
	var updateData struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		ctrl.Logger.Warn("Failed to parse profile update data", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
	}

	// Get current guest for email
	currentGuest, err := ctrl.Service.GetGuestByID(guestID)
	if err != nil {
		ctrl.Logger.Error("Failed to get current guest data", zap.Uint("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Guest not found: " + err.Error(),
		})
	}

	// Update guest profile
	updatedGuest, err := ctrl.Service.UpdateGuest(guestID, updateData.FullName, currentGuest.Email, updateData.Phone)
	if err != nil {
		ctrl.Logger.Error("Failed to update guest profile", zap.Uint("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update profile: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data":    updatedGuest,
	})
}

// GetGuestBookings returns the guest's bookings
func (ctrl *GuestController) GetGuestBookings(c *fiber.Ctx) error {
	// Get guest ID from authenticated user
	guestID, ok := c.Locals("guestID").(uint)
	if !ok {
		ctrl.Logger.Warn("Guest ID not found in request context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Get bookings
	bookings, err := ctrl.Service.GetGuestBookingHistory(guestID)
	if err != nil {
		ctrl.Logger.Error("Failed to get guest bookings", zap.Uint("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve bookings: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bookings,
	})
}

// Admin endpoints
// GetAllGuests returns all guests (admin only)
func (ctrl *GuestController) GetAllGuests(c *fiber.Ctx) error {
	guests, err := ctrl.Service.GetAllGuests()
	if err != nil {
		ctrl.Logger.Error("Failed to get all guests", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve guests: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    guests,
	})
}

// GetGuestByID returns a guest by ID (admin only)
func (ctrl *GuestController) GetGuestByID(c *fiber.Ctx) error {
	guestID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		ctrl.Logger.Warn("Invalid guest ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid guest ID",
		})
	}

	guest, err := ctrl.Service.GetGuestByID(uint(guestID))
	if err != nil {
		ctrl.Logger.Error("Failed to get guest", zap.Int("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Guest not found: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    guest,
	})
}

// GetGuestByEmail returns a guest by email (admin only)
func (ctrl *GuestController) GetGuestByEmail(c *fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		ctrl.Logger.Warn("Missing email parameter")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email parameter is required",
		})
	}

	guest, err := ctrl.Service.GetGuestByEmail(email)
	if err != nil {
		ctrl.Logger.Error("Failed to get guest by email", zap.String("email", email), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Guest not found: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    guest,
	})
}

// UpdateGuest updates a guest (admin only)
func (ctrl *GuestController) UpdateGuest(c *fiber.Ctx) error {
	guestID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		ctrl.Logger.Warn("Invalid guest ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid guest ID",
		})
	}

	var updateData struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		ctrl.Logger.Warn("Failed to parse guest update data", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
	}

	updatedGuest, err := ctrl.Service.UpdateGuest(uint(guestID), updateData.FullName, updateData.Email, updateData.Phone)
	if err != nil {
		ctrl.Logger.Error("Failed to update guest", zap.Int("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update guest: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Guest updated successfully",
		"data":    updatedGuest,
	})
}

// DeleteGuest deletes a guest (admin only)
func (ctrl *GuestController) DeleteGuest(c *fiber.Ctx) error {
	guestID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		ctrl.Logger.Warn("Invalid guest ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid guest ID",
		})
	}

	if err := ctrl.Service.DeleteGuest(uint(guestID)); err != nil {
		ctrl.Logger.Error("Failed to delete guest", zap.Int("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete guest: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Guest deleted successfully",
	})
}

// GetGuestBookingHistory returns the booking history for a guest (admin only)
func (ctrl *GuestController) GetGuestBookingHistory(c *fiber.Ctx) error {
	guestID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		ctrl.Logger.Warn("Invalid guest ID", zap.String("id", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid guest ID",
		})
	}

	bookings, err := ctrl.Service.GetGuestBookingHistory(uint(guestID))
	if err != nil {
		ctrl.Logger.Error("Failed to get guest booking history", zap.Int("guestID", guestID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve booking history: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bookings,
	})
}

// CreateGuest creates a new guest (admin only)
func (ctrl *GuestController) CreateGuest(c *fiber.Ctx) error {
	var guest models.Guest

	if err := c.BodyParser(&guest); err != nil {
		ctrl.Logger.Warn("Failed to parse guest creation request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
	}

	createdGuest, err := ctrl.Service.CreateGuest(&guest)
	if err != nil {
		ctrl.Logger.Error("Failed to create guest", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create guest: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Guest created successfully",
		"data":    createdGuest,
	})
}
