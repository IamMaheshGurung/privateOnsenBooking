package routes

import (
	"github.com/IamMaheshGurung/privateOnsenBooking/controllers"
	"github.com/IamMaheshGurung/privateOnsenBooking/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App,
	roomController *controllers.RoomController,
	bookingController *controllers.BookingController,
	guestController *controllers.GuestController) {

	// Setup global middleware
	setupMiddleware(app)

	// Static pages
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Kwangdi Hotel - Traditional Japanese Ryokan",
		})
	})

	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{
			"Title": "About - Kwangdi Hotel",
		})
	})

	app.Get("/contact", func(c *fiber.Ctx) error {
		return c.Render("contact", fiber.Map{
			"Title": "Contact Us - Kwangdi Hotel",
		})
	})

	// Partials for HTMX
	partials := app.Group("/partials")
	setupPartialsRoutes(partials)

	// Main feature routes
	setupRoomRoutes(app, roomController)
	setupBookingRoutes(app, bookingController)
	setupGuestRoutes(app, guestController)

	// Setup Onsen routes if you add the controller later
	// setupOnsenRoutes(app, onsenController)

	// Setup static file serving
	app.Static("/static", "./static")

	// API routes for frontend functionality
	api := app.Group("/api")
	setupApiRoutes(api, roomController, bookingController, guestController)

	// Admin routes with authentication
	admin := app.Group("/admin", middleware.AdminAuth())
	setupAdminRoutes(admin, roomController, bookingController, guestController)
}

// setupMiddleware configures global middleware for the application
func setupMiddleware(app *fiber.App) {
	// Recover from panics
	app.Use(recover.New())

	// Request logging
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	// CORS setup
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Compress responses
	app.Use(compress.New())

	// Add request ID
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Request-ID", c.GetRespHeader("X-Request-ID"))
		return c.Next()
	})
}

// setupPartialsRoutes configures routes for HTML partials (used by HTMX)
func setupPartialsRoutes(router fiber.Router) {
	// Navigation and footer
	router.Get("/navigation", func(c *fiber.Ctx) error {
		return c.Render("partials/navigation", fiber.Map{})
	})

	router.Get("/footer", func(c *fiber.Ctx) error {
		return c.Render("partials/footer", fiber.Map{})
	})

	// Home page partials
	router.Get("/hero", func(c *fiber.Ctx) error {
		return c.Render("partials/hero", fiber.Map{})
	})

	// Room partials
	router.Get("/room-card", func(c *fiber.Ctx) error {
		return c.Render("partials/room_card", fiber.Map{})
	})

	router.Get("/room-grid", func(c *fiber.Ctx) error {
		return c.Render("partials/room_grid", fiber.Map{})
	})

	// Booking partials
	router.Get("/booking-form", func(c *fiber.Ctx) error {
		return c.Render("partials/booking_form", fiber.Map{})
	})

	router.Get("/booking-summary", func(c *fiber.Ctx) error {
		return c.Render("partials/booking_summary", fiber.Map{})
	})
}

// setupRoomRoutes configures room-related routes
func setupRoomRoutes(app *fiber.App, controller *controllers.RoomController) {
	// Public room routes
	rooms := app.Group("/rooms")

	// Room listing and details
	rooms.Get("/", controller.GetAllRooms)
	rooms.Get("/preview", controller.PreviewRooms)
	rooms.Get("/:id", controller.GetRoomDetails)
	rooms.Get("/types", controller.GetRoomTypes)

	// Room availability checking
	rooms.Get("/available", controller.GetAvailableRooms)
}

// setupBookingRoutes configures booking-related routes
func setupBookingRoutes(app *fiber.App, controller *controllers.BookingController) {
	// Public booking routes
	booking := app.Group("/booking")

	// Main booking flow
	booking.Get("/", controller.ShowBookingForm)
	booking.Post("/reserve", controller.CreateBooking)
	booking.Get("/confirmation/:id", controller.ShowConfirmation)

	// Availability checking
	booking.Get("/check", controller.CheckAvailability)

	// Guest booking management
	lookup := booking.Group("/lookup")
	lookup.Get("/", controller.ShowLookupForm)
	lookup.Post("/", controller.LookupBooking)
	lookup.Get("/:id/:token", controller.ShowBookingDetails)
	lookup.Post("/:id/:token/cancel", controller.CancelBookingByGuest)
}

// setupGuestRoutes configures guest-related routes
func setupGuestRoutes(app *fiber.App, controller *controllers.GuestController) {
	// Guest account routes
	guest := app.Group("/guest")

	// Guest registration and profile
	guest.Post("/register", controller.RegisterGuest)
	guest.Get("/profile", middleware.GuestAuth(), controller.GetGuestProfile)
	guest.Put("/profile", middleware.GuestAuth(), controller.UpdateGuestProfile)

	// Guest bookings
	guest.Get("/bookings", middleware.GuestAuth(), controller.GetGuestBookings)
}

// setupApiRoutes configures API routes for frontend functionality
func setupApiRoutes(router fiber.Router,
	roomController *controllers.RoomController,
	bookingController *controllers.BookingController,
	guestController *controllers.GuestController) {

	// Room API endpoints
	roomsApi := router.Group("/rooms")
	roomsApi.Get("/", roomController.GetAllRooms)
	roomsApi.Get("/:id", roomController.GetRoomByID)
	roomsApi.Get("/available", roomController.GetAvailableRooms)

	// Booking API endpoints
	bookingsApi := router.Group("/bookings")
	bookingsApi.Get("/check", bookingController.CheckAvailability)
	bookingsApi.Get("/available", bookingController.GetAvailableRooms)
	bookingsApi.Post("/", bookingController.CreateBooking)
	bookingsApi.Get("/:id", bookingController.GetBookingByID)
	bookingsApi.Put("/:id/cancel", bookingController.CancelBooking)
	bookingsApi.Get("/guest", bookingController.GetGuestBookings)

	// Guest API endpoints
	guestsApi := router.Group("/guests")
	guestsApi.Post("/", guestController.CreateGuest)
	guestsApi.Get("/:id", guestController.GetGuestByID)
	guestsApi.Get("/email/:email", guestController.GetGuestByEmail)
}

// setupAdminRoutes configures admin routes
func setupAdminRoutes(router fiber.Router,
	roomController *controllers.RoomController,
	bookingController *controllers.BookingController,
	guestController *controllers.GuestController) {

	// Admin dashboard
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Render("admin/dashboard", fiber.Map{
			"Title": "Admin Dashboard - Kwangdi Hotel",
		})
	})

	// Room management
	adminRooms := router.Group("/rooms")
	adminRooms.Get("/", roomController.AdminListRooms)
	adminRooms.Get("/new", roomController.ShowAddRoomForm)
	adminRooms.Post("/new", roomController.AddRoom)
	adminRooms.Get("/:id/edit", roomController.ShowEditRoomForm)
	adminRooms.Post("/:id/edit", roomController.UpdateRoom)
	adminRooms.Delete("/:id", roomController.DeleteRoom)

	// Booking management
	adminBookings := router.Group("/bookings")
	adminBookings.Get("/", bookingController.GetAllBookings)
	adminBookings.Get("/date/:date", bookingController.GetBookingsByDate)
	adminBookings.Get("/range", bookingController.GetBookingsByDateRange)
	adminBookings.Put("/:id", bookingController.UpdateBooking)
	adminBookings.Put("/:id/check-in", bookingController.CheckInGuest)
	adminBookings.Put("/:id/check-out", bookingController.CheckOutGuest)

	// Guest management
	adminGuests := router.Group("/guests")
	adminGuests.Get("/", guestController.GetAllGuests)
	adminGuests.Put("/:id", guestController.UpdateGuest)
	adminGuests.Delete("/:id", guestController.DeleteGuest)
	adminGuests.Get("/:id/bookings", guestController.GetGuestBookingHistory)
}

// SetupErrorHandlers configures custom error handlers
func SetupErrorHandlers(app *fiber.App) {
	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		// Check if request is API or HTML
		if c.Accepts("json") == "json" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "Endpoint not found",
				"code":    404,
			})
		}

		return c.Status(fiber.StatusNotFound).Render("error/404", fiber.Map{
			"Title": "Page Not Found - Kwangdi Hotel",
		})
	})

	// Error handler
	app.Use(func(c *fiber.Ctx, err error) error {
		// Status code defaults to 500
		code := fiber.StatusInternalServerError

		// Check if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Render error page or return JSON based on request
		if c.Accepts("html") == "html" {
			return c.Status(code).Render("error/error", fiber.Map{
				"Title":     "Error - Kwangdi Hotel",
				"ErrorCode": code,
				"Message":   err.Error(),
			})
		}

		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"code":    code,
		})
	})
}
