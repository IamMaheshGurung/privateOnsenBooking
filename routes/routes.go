package routes

import (
	"fmt"
	"time"

	"github.com/IamMaheshGurung/privateOnsenBooking/controllers"
	"github.com/gofiber/fiber/v2"
	// Import your controllers if you have them
	// "github.com/yourusername/privateOnsen/controllers"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	app *fiber.App,
	roomController *controllers.RoomController,
	bookingController *controllers.BookingController,
	guestController *controllers.GuestController,
) {
	// Setup routes by category
	setupBasicRoutes(app)
	setupPageRoutes(app)
	setupBookingRoutes(app)
	setupExperienceRoutes(app)
	setupGalleryRoutes(app)
	setupBlogRoutes(app)
	setupRoomRoutes(app)
	setupDiningRoutes(app)
	setupAdminRoutes(app)
	setupAPIRoutes(app)
}

// setupBasicRoutes configures test and debug routes
func setupBasicRoutes(app *fiber.App) {
	// Simple test endpoint
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(`
            <!DOCTYPE html>
            <html>
            <head>
                <title>Test Page</title>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
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

// setupPageRoutes configures basic page routes
func setupPageRoutes(app *fiber.App) {
	// Home page
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":       "Kwangdi Pahuna Ghar - Traditional Nepali Guesthouse",
			"Description": "Experience authentic Nepali hospitality in the beautiful Shantipur valley of Gulmi district",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Alternative home route
	app.Get("/home", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":       "Kwangdi Pahuna Ghar - Traditional Nepali Guesthouse",
			"Description": "Experience authentic Nepali hospitality in the beautiful Shantipur valley of Gulmi district",
			"CurrentYear": time.Now().Year(),
		})
	})

	// About page
	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{
			"Title":       "About Us | Kwangdi Pahuna Ghar",
			"Description": "Learn about our story, values and the team behind Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Contact page
	app.Get("/contact", func(c *fiber.Ctx) error {
		return c.Render("contact", fiber.Map{
			"Title":       "Contact Us | Kwangdi Pahuna Ghar",
			"Description": "Get in touch with us for bookings, inquiries or feedback",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Location page
	app.Get("/location", func(c *fiber.Ctx) error {
		return c.Render("location", fiber.Map{
			"Title":       "Location | Kwangdi Pahuna Ghar",
			"Description": "Find us in the beautiful Shantipur valley of Gulmi district, Nepal",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Virtual Tour page
	app.Get("/virtual-tour", func(c *fiber.Ctx) error {
		return c.Render("virtual-tour", fiber.Map{
			"Title":       "Virtual Tour | Kwangdi Pahuna Ghar",
			"Description": "Take a virtual tour of our guesthouse and surroundings",
			"CurrentYear": time.Now().Year(),
		})
	})

	// FAQ page
	app.Get("/faq", func(c *fiber.Ctx) error {
		return c.Render("faq", fiber.Map{
			"Title":       "FAQs | Kwangdi Pahuna Ghar",
			"Description": "Frequently asked questions about your stay at Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Legal pages
	app.Get("/terms", func(c *fiber.Ctx) error {
		return c.Render("terms", fiber.Map{
			"Title":       "Terms & Conditions | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/privacy", func(c *fiber.Ctx) error {
		return c.Render("privacy", fiber.Map{
			"Title":       "Privacy Policy | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/cookies", func(c *fiber.Ctx) error {
		return c.Render("cookies", fiber.Map{
			"Title":       "Cookie Policy | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/sitemap", func(c *fiber.Ctx) error {
		return c.Render("sitemap", fiber.Map{
			"Title":       "Sitemap | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupRoomRoutes configures room-related routes
func setupRoomRoutes(app *fiber.App) {
	// Rooms main page
	app.Get("/rooms", func(c *fiber.Ctx) error {
		return c.Render("rooms/index", fiber.Map{
			"Title":       "Accommodations | Kwangdi Pahuna Ghar",
			"Description": "Explore our comfortable and authentic Nepali accommodations",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Room categories
	app.Get("/rooms/standard", func(c *fiber.Ctx) error {
		return c.Render("rooms/standard", fiber.Map{
			"Title":       "Standard Rooms | Kwangdi Pahuna Ghar",
			"Description": "Our comfortable standard rooms with traditional Nepali touches",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/rooms/deluxe", func(c *fiber.Ctx) error {
		return c.Render("rooms/deluxe", fiber.Map{
			"Title":       "Deluxe Rooms | Kwangdi Pahuna Ghar",
			"Description": "Spacious deluxe rooms with premium amenities and valley views",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/rooms/family", func(c *fiber.Ctx) error {
		return c.Render("rooms/family", fiber.Map{
			"Title":       "Family Suites | Kwangdi Pahuna Ghar",
			"Description": "Our finest family accommodations with panoramic mountain views",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Individual room details - dynamic routes
	app.Get("/rooms/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		// In a real app, you would fetch the room data from a database
		return c.Render("rooms/detail", fiber.Map{
			"Title":       "Room Details | Kwangdi Pahuna Ghar",
			"Description": "Detailed information about our accommodations",
			"RoomID":      id,
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupExperienceRoutes configures experience-related routes
func setupExperienceRoutes(app *fiber.App) {
	// Main experiences page
	app.Get("/experiences", func(c *fiber.Ctx) error {
		return c.Render("experiences/index", fiber.Map{
			"Title":       "Cultural Experiences | Kwangdi Pahuna Ghar",
			"Description": "Discover authentic Nepali cultural experiences during your stay",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Cultural experiences
	app.Get("/experiences/trekking", func(c *fiber.Ctx) error {
		return c.Render("experiences/trekking", fiber.Map{
			"Title":       "Guided Treks | Kwangdi Pahuna Ghar",
			"Description": "Explore the beautiful trails of Gulmi district with our expert guides",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/experiences/cultural", func(c *fiber.Ctx) error {
		return c.Render("experiences/cultural", fiber.Map{
			"Title":       "Cultural Tours | Kwangdi Pahuna Ghar",
			"Description": "Immerse yourself in the rich cultural heritage of rural Nepal",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/experiences/cooking", func(c *fiber.Ctx) error {
		return c.Render("experiences/cooking", fiber.Map{
			"Title":       "Nepali Cooking Classes | Kwangdi Pahuna Ghar",
			"Description": "Learn how to prepare authentic Nepali dishes with our experienced cooks",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/experiences/farming", func(c *fiber.Ctx) error {
		return c.Render("experiences/farming", fiber.Map{
			"Title":       "Farming Experience | Kwangdi Pahuna Ghar",
			"Description": "Experience traditional Nepali farming methods and organic gardening",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Cultural dances and performances
	app.Get("/experiences/panche-baja", func(c *fiber.Ctx) error {
		return c.Render("experiences/panche-baja", fiber.Map{
			"Title":       "Panche Baja | Kwangdi Pahuna Ghar",
			"Description": "Experience traditional Nepali folk music with Panche Baja performances",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/experiences/sorathi", func(c *fiber.Ctx) error {
		return c.Render("experiences/sorathi", fiber.Map{
			"Title":       "Sorathi Dance | Kwangdi Pahuna Ghar",
			"Description": "Witness the enchanting Sorathi folk dance cultural performance",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/experiences/gatu-nach", func(c *fiber.Ctx) error {
		return c.Render("experiences/gatu-nach", fiber.Map{
			"Title":       "Gatu Nach | Kwangdi Pahuna Ghar",
			"Description": "Enjoy the traditional Gatu Nach dance of the Gurung community",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/experiences/kwangdi-club", func(c *fiber.Ctx) error {
		return c.Render("experiences/kwangdi-club", fiber.Map{
			"Title":       "Kwangdi Club Dance | Kwangdi Pahuna Ghar",
			"Description": "Experience fusion dance performances by talented local youth",
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupDiningRoutes configures dining-related routes
func setupDiningRoutes(app *fiber.App) {
	// Dining main page
	app.Get("/dining", func(c *fiber.Ctx) error {
		return c.Render("dining/index", fiber.Map{
			"Title":       "Dining | Kwangdi Pahuna Ghar",
			"Description": "Discover authentic Nepali cuisine prepared with fresh local ingredients",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Special dining experiences
	app.Get("/dining/menu", func(c *fiber.Ctx) error {
		return c.Render("dining/menu", fiber.Map{
			"Title":       "Menu | Kwangdi Pahuna Ghar",
			"Description": "Our full menu featuring traditional Nepali dishes and local specialties",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/dining/special", func(c *fiber.Ctx) error {
		return c.Render("dining/special", fiber.Map{
			"Title":       "Special Dining | Kwangdi Pahuna Ghar",
			"Description": "Special dining experiences and celebrations at our guesthouse",
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupBookingRoutes configures booking-related routes
// setupBookingRoutes configures booking-related routes
func setupBookingRoutes(app *fiber.App) {
	app.Get("/booking", func(c *fiber.Ctx) error {

		today := time.Now()
		tomorrow := today.AddDate(0, 0, 1)

		// Format dates for HTML date input (YYYY-MM-DD)
		todayStr := today.Format("2006-01-02")
		tomorrowStr := tomorrow.Format("2006-01-02")

		// Format dates for display
		todayDisplay := today.Format("Jan 2, 2006")
		tomorrowDisplay := tomorrow.Format("Jan 2, 2006")

		fmt.Println("Attempting to render booking/index template")
		err := c.Render("booking/form", fiber.Map{
			"Title":           "Book Your Stay | Kwangdi Pahuna Ghar",
			"Description":     "Reserve your room at our traditional Nepali guesthouse",
			"CurrentYear":     today.Year(),
			"TodayDate":       todayStr,
			"TomorrowDate":    tomorrowStr,
			"TodayDisplay":    todayDisplay,
			"TomorrowDisplay": tomorrowDisplay,
		})

		if err != nil {
			fmt.Println("ERROR rendering template:", err)
			return c.Status(500).SendString("Template error: " + err.Error())
		}

		return nil
	})

	// Availability check - updated to handle form data
	app.Get("/booking/availability", func(c *fiber.Ctx) error {
		// Get form data
		checkIn := c.Query("check_in")
		checkOut := c.Query("check_out")
		guests := c.Query("guests")

		// Format dates for display
		checkInDate, _ := time.Parse("2006-01-02", checkIn)
		checkOutDate, _ := time.Parse("2006-01-02", checkOut)
		checkInDisplay := checkInDate.Format("Jan 2, 2006")
		checkOutDisplay := checkOutDate.Format("Jan 2, 2006")

		// Calculate nights
		nights := checkOutDate.Sub(checkInDate).Hours() / 24

		return c.Render("booking/availability", fiber.Map{
			"Title":         "Available Rooms | Kwangdi Pahuna Ghar",
			"Description":   "Available rooms for your selected dates",
			"CurrentYear":   time.Now().Year(),
			"Check_in":      checkInDisplay,
			"Check_out":     checkOutDisplay,
			"Check_in_raw":  checkIn,
			"Check_out_raw": checkOut,
			"Guests":        guests,
			"Nights":        int(nights),
		})
	})

	// Booking confirmation
	app.Get("/booking/confirmation/:id", func(c *fiber.Ctx) error {
		bookingID := c.Params("id")
		return c.Render("booking/confirmation", fiber.Map{
			"Title":       "Booking Confirmation | Kwangdi Pahuna Ghar",
			"BookingID":   bookingID,
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupGalleryRoutes configures photo gallery routes
func setupGalleryRoutes(app *fiber.App) {
	// Gallery main page
	app.Get("/gallery", func(c *fiber.Ctx) error {
		return c.Render("gallery/index", fiber.Map{
			"Title":       "Photo Gallery | Kwangdi Pahuna Ghar",
			"Description": "Browse photos of our guesthouse, rooms, surroundings and experiences",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Gallery categories
	app.Get("/gallery/accommodations", func(c *fiber.Ctx) error {
		return c.Render("gallery/accommodations", fiber.Map{
			"Title":       "Accommodation Photos | Kwangdi Pahuna Ghar",
			"Description": "Photos of our rooms and accommodations",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/gallery/surroundings", func(c *fiber.Ctx) error {
		return c.Render("gallery/surroundings", fiber.Map{
			"Title":       "Surroundings Photos | Kwangdi Pahuna Ghar",
			"Description": "Photos of the beautiful Shantipur valley and surroundings",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/gallery/cultural", func(c *fiber.Ctx) error {
		return c.Render("gallery/cultural", fiber.Map{
			"Title":       "Cultural Photos | Kwangdi Pahuna Ghar",
			"Description": "Photos of our cultural performances and experiences",
			"CurrentYear": time.Now().Year(),
		})
	})

	app.Get("/gallery/dining", func(c *fiber.Ctx) error {
		return c.Render("gallery/dining", fiber.Map{
			"Title":       "Dining Photos | Kwangdi Pahuna Ghar",
			"Description": "Photos of our traditional Nepali cuisine and dining experiences",
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupBlogRoutes configures blog-related routes
func setupBlogRoutes(app *fiber.App) {
	// Blog main page
	app.Get("/blog", func(c *fiber.Ctx) error {
		return c.Render("blog/index", fiber.Map{
			"Title":       "Blog | Kwangdi Pahuna Ghar",
			"Description": "News, stories and insights from our guesthouse and the Shantipur valley",
			"CurrentYear": time.Now().Year(),
		})
	})

	// Blog categories
	app.Get("/blog/category/:category", func(c *fiber.Ctx) error {
		category := c.Params("category")
		return c.Render("blog/category", fiber.Map{
			"Title":       fmt.Sprintf("%s | Blog | Kwangdi Pahuna Ghar", category),
			"Category":    category,
			"CurrentYear": time.Now().Year(),
		})
	})

	// Individual blog post
	app.Get("/blog/:slug", func(c *fiber.Ctx) error {
		slug := c.Params("slug")
		// In a real app, you would fetch the blog post from a database
		return c.Render("blog/post", fiber.Map{
			"Title":       "Blog Post | Kwangdi Pahuna Ghar",
			"Slug":        slug,
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupAdminRoutes configures admin panel routes
func setupAdminRoutes(app *fiber.App) {
	// Admin routes should be protected with authentication middleware
	admin := app.Group("/admin", func(c *fiber.Ctx) error {
		// This is where you would check for admin authentication
		// For now, we'll just pass through
		return c.Next()
	})

	admin.Get("/", func(c *fiber.Ctx) error {
		return c.Render("admin/dashboard", fiber.Map{
			"Title":       "Admin Dashboard | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	admin.Get("/bookings", func(c *fiber.Ctx) error {
		return c.Render("admin/bookings", fiber.Map{
			"Title":       "Manage Bookings | Admin | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	admin.Get("/rooms", func(c *fiber.Ctx) error {
		return c.Render("admin/rooms", fiber.Map{
			"Title":       "Manage Rooms | Admin | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	admin.Get("/blog", func(c *fiber.Ctx) error {
		return c.Render("admin/blog", fiber.Map{
			"Title":       "Manage Blog | Admin | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})

	admin.Get("/users", func(c *fiber.Ctx) error {
		return c.Render("admin/users", fiber.Map{
			"Title":       "Manage Users | Admin | Kwangdi Pahuna Ghar",
			"CurrentYear": time.Now().Year(),
		})
	})
}

// setupAPIRoutes configures API endpoints (separate from web routes)
func setupAPIRoutes(app *fiber.App) {
	api := app.Group("/api")

	// API version group
	v1 := api.Group("/v1")

	// Room availability API
	v1.Get("/rooms/available", func(c *fiber.Ctx) error {
		// In a real app, you would check room availability from a database
		return c.JSON(fiber.Map{
			"success": true,
			"data": []fiber.Map{
				{"id": "1", "type": "standard", "available": true, "price": 2500},
				{"id": "2", "type": "deluxe", "available": true, "price": 4000},
				{"id": "3", "type": "family", "available": false, "price": 6000},
			},
		})
	})

	// Room booking API
	v1.Post("/bookings", func(c *fiber.Ctx) error {
		// In a real app, you would save the booking to a database
		return c.Status(201).JSON(fiber.Map{
			"success":    true,
			"message":    "Booking created successfully",
			"booking_id": "BK12345",
		})
	})

	// Contact form submission API
	v1.Post("/contact", func(c *fiber.Ctx) error {
		// In a real app, you would process the contact form submission
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Thank you for contacting us. We will respond shortly.",
		})
	})

	// Newsletter subscription API
	v1.Post("/subscribe", func(c *fiber.Ctx) error {
		// In a real app, you would save the subscriber to a database
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Successfully subscribed to our newsletter",
		})
	})

	// Blog posts API
	v1.Get("/blog/posts", func(c *fiber.Ctx) error {
		// In a real app, you would fetch blog posts from a database
		return c.JSON(fiber.Map{
			"success": true,
			"data": []fiber.Map{
				{"id": "1", "title": "Welcome to Kwangdi Pahuna Ghar", "slug": "welcome"},
				{"id": "2", "title": "Exploring Shantipur Valley", "slug": "exploring-shantipur"},
				{"id": "3", "title": "Traditional Nepali Cuisine", "slug": "nepali-cuisine"},
			},
		})
	})

	// Testimonials API
	v1.Get("/testimonials", func(c *fiber.Ctx) error {
		// In a real app, you would fetch testimonials from a database
		return c.JSON(fiber.Map{
			"success": true,
			"data": []fiber.Map{
				{"id": "1", "name": "Maria & John", "country": "United Kingdom", "rating": 5, "comment": "Our stay at Kwangdi Pahuna Ghar was the highlight of our Nepal trip."},
				{"id": "2", "name": "David & Lisa", "country": "Australia", "rating": 5, "comment": "If you want to experience authentic rural Nepal, this is the place to stay."},
			},
		})
	})

	// Admin API endpoints (should be protected with authentication)
	admin := v1.Group("/admin", func(c *fiber.Ctx) error {
		// This is where you would check for admin API authentication
		// For now, we'll just pass through
		return c.Next()
	})

	admin.Get("/bookings", func(c *fiber.Ctx) error {
		// In a real app, you would fetch bookings from a database
		return c.JSON(fiber.Map{
			"success": true,
			"data": []fiber.Map{
				{"id": "BK12345", "guest": "John Smith", "room": "1", "check_in": "2023-10-15", "check_out": "2023-10-18"},
				{"id": "BK12346", "guest": "Jane Doe", "room": "2", "check_in": "2023-10-20", "check_out": "2023-10-25"},
			},
		})
	})
}
