package controllers

import (
	"strconv"

	"github.com/IamMaheshGurung/privateOnsenBooking/services"
	"github.com/gofiber/fiber/v2"
)

// GuestController handles guest-related HTTP requests
type GuestController struct {
	Service *services.GuestService
}

// NewGuestController creates a new instance of GuestController
func NewGuestController(service *services.GuestService) *GuestController {
	return &GuestController{Service: service}
}

// CreateGuest creates a new guest
// POST /api/guests
func (ctrl *GuestController) CreateGuest(c *fiber.Ctx) error {
	var guest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	if err := c.BodyParser(&guest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON: " + err.Error(),
		})
	}

	// Validate guest data
	if guest.Name == "" || guest.Email == "" || guest.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, email, and phone are required",
		})
	}

	createdGuest, err := ctrl.Service.CreateGuest(guest.Name, guest.Email, guest.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create guest: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    createdGuest,
	})
}

// GetGuestByID returns a specific guest
// GET /api/guests/:id
func (ctrl *GuestController) GetGuestByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid guest ID",
		})
	}

	guest, err := ctrl.Service.GetGuestByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Guest not found: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    guest,
	})
}

// GetGuestByEmail returns a specific guest by email
// GET /api/guests/email/:email
func (ctrl *GuestController) GetGuestByEmail(c *fiber.Ctx) error {
	email := c.Params("email")

	guest, err := ctrl.Service.GetGuestByEmail(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Guest not found: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    guest,
	})
}

// Admin Routes

// GetAllGuests returns all guests with pagination
// GET /api/admin/guests?limit=10&offset=0&search=name
func (ctrl *GuestController) GetAllGuests(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")
	search := c.Query("search", "")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	guests, count, err := ctrl.Service.GetAllGuests(limit, offset, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get guests: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"total":   count,
		"data":    guests,
	})
}

// UpdateGuest updates a guest
// PUT /api/admin/guests/:id
func (ctrl *GuestController) UpdateGuest(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid guest ID",
		})
	}

	var guest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	if err := c.BodyParser(&guest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON: " + err.Error(),
		})
	}

	updatedGuest, err := ctrl.Service.UpdateGuest(uint(id), guest.Name, guest.Email, guest.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update guest: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    updatedGuest,
	})
}

// DeleteGuest deletes a guest
// DELETE /api/admin/guests/:id
func (ctrl *GuestController) DeleteGuest(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid guest ID",
		})
	}

	if err := ctrl.Service.DeleteGuest(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete guest: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Guest deleted successfully",
	})
}

// GetGuestBookingHistory returns booking history for a guest
// GET /api/admin/guests/:id/bookings
func (ctrl *GuestController) GetGuestBookingHistory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid guest ID",
		})
	}

	bookings, err := ctrl.Service.GetGuestBookingHistory(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get booking history: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    bookings,
	})
}
