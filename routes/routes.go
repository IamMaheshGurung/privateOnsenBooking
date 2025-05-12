package routes

import (
	"fmt"
	"time"

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

	// 1. Setup global middleware first
	setupMiddleware(app)

	// 2. CRITICAL: Setup partials routes before any page routes
	// This ensures partials are registered and available for HTMX requests
	partials := app.Group("/partials")
	setupPartialsRoutes(partials)

	// 3. Add debugging routes to help troubleshoot template issues
	setupDebugRoutes(app)

	// 4. Then setup static file serving (early to ensure assets are available)
	app.Static("/static", "./static")

	// 5. Only after partials and static files are setup, register page routes
	setupPageRoutes(app)

	// 6. Register feature routes
	setupRoomRoutes(app, roomController)
	setupBookingRoutes(app, bookingController)
	setupGuestRoutes(app, guestController)

	// 7. API routes for frontend functionality
	api := app.Group("/api")
	setupApiRoutes(api, roomController, bookingController, guestController)

	// 8. Admin routes with authentication
	admin := app.Group("/admin", middleware.AdminAuth())
	setupAdminRoutes(admin, roomController, bookingController, guestController)
}

// setupPartialsRoutes configures routes for HTML partials (used by HTMX)
// This MUST be called before any page routes to ensure partials are available
func setupPartialsRoutes(router fiber.Router) {
	// Log that we're setting up partials
	fmt.Println("Setting up partials routes")

	// Navigation and footer
	router.Get("/navigation", func(c *fiber.Ctx) error {
		fmt.Println("Loading navigation partial")
		return c.Render("partials/navigation", fiber.Map{
			"CurrentPath": c.Path(),
			"CurrentYear": time.Now().Year(),
		})
	})

	router.Get("/footer", func(c *fiber.Ctx) error {
		fmt.Println("Loading footer partial")
		return c.Render("partials/footer", fiber.Map{
			"CurrentYear": time.Now().Year(),
		})
	})

	// Home page partials
	router.Get("/hero", func(c *fiber.Ctx) error {
		fmt.Println("Loading hero partial")
		return c.Render("partials/hero", fiber.Map{})
	})

	// Room partials
	router.Get("/room-card", func(c *fiber.Ctx) error {
		fmt.Println("Loading room-card partial")
		return c.Render("partials/room_card", fiber.Map{})
	})

	router.Get("/room-grid", func(c *fiber.Ctx) error {
		fmt.Println("Loading room-grid partial")
		return c.Render("partials/room_grid", fiber.Map{})
	})

	// Booking partials
	router.Get("/booking-form", func(c *fiber.Ctx) error {
		return c.Render("partials/booking_form", fiber.Map{})
	})

	router.Get("/booking-summary", func(c *fiber.Ctx) error {
		return c.Render("partials/booking_summary", fiber.Map{})
	})

	// HTMX test endpoint (for debugging)
	router.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("<div class='bg-green-100 p-4 rounded text-green-800'>HTMX is working! Loaded at: " +
			time.Now().Format("15:04:05") + "</div>")
	})
}

// setupDebugRoutes adds debugging routes to test template rendering
func setupDebugRoutes(app *fiber.App) {
	// Direct HTML response (no templates)
	app.Get("/debug/direct", func(c *fiber.Ctx) error {
		return c.SendString(`
            <!DOCTYPE html>
            <html>
            <head>
                <title>Direct HTML Response</title>
                <script src="https://cdn.tailwindcss.com"></script>
            </head>
            <body class="p-8 bg-gray-100">
                <h1 class="text-2xl font-bold mb-4">Direct HTML Response</h1>
                <p>If you can see this, basic HTTP response is working.</p>
                <p>Generated at: ` + time.Now().Format(time.RFC3339) + `</p>
            </body>
            </html>
        `)
	})

	// HTMX test page
	app.Get("/debug/htmx", func(c *fiber.Ctx) error {
		return c.Render("debug/htmx-test", fiber.Map{
			"Title": "HTMX Test Page",
		})
	})
}

// setupPageRoutes configures basic page routes (after partials are setup)
func setupPageRoutes(app *fiber.App) {
	// Home page

	// Add this to your setupPageRoutes function in routes.go
	app.Get("/simple", func(c *fiber.Ctx) error {
		// Send direct HTML without using templates
		return c.SendString(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Simple Test</title>
        </head>
        <body>
            <h1>Simple Test Page</h1>
            <p>If you can see this, your server is responding correctly.</p>
            <p>Time: ` + time.Now().Format(time.RFC3339) + `</p>
        </body>
        </html>
    `)
	})
	app.Get("/home", func(c *fiber.Ctx) error {
		// Log that we're handling the request
		fmt.Println("Rendering index template")

		// Try rendering with additional data
		err := c.Render("index", fiber.Map{
			"Title":       "Kwangdi Hotel - Traditional Japanese Ryokan",
			"Debug":       "If you can see this, template is rendering",
			"CurrentYear": time.Now().Year(),
		})

		// Check for errors
		if err != nil {
			fmt.Println("Error rendering template:", err)
			return c.Status(500).SendString("Error: " + err.Error())
		}

		fmt.Println("Successfully rendered template")
		return nil
	})

	// About page
	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{
			"Title":       "About - Kwangdi Hotel",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Contact page
	app.Get("/contact", func(c *fiber.Ctx) error {
		return c.Render("contact", fiber.Map{
			"Title":       "Contact Us - Kwangdi Hotel",
			"CurrentYear": time.Now().Year(),
		})
	})
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
