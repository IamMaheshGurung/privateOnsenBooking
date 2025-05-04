package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/IamMaheshGurung/privateOnsenBooking/database"
	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"github.com/IamMaheshGurung/privateOnsenBooking/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"go.uber.org/zap"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		fmt.Println("Error connecting to database", err)
		return
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.Room{}, &models.Guest{}, &models.RoomBooking{}); err != nil {
		fmt.Println("Error auto-migrating database:", err)
		return
	}

	database.SeedRooms(db)

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("Error initializing logger", err)
	}

	// Configure the template engine with correct directory
	engine := html.New("./templates", ".html")

	// Add template functions
	engine.AddFunc("dict", func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	})

	// Enable reloading for development
	engine.Reload(true)

	// Explicitly load all templates including partials
	if err := engine.Load(); err != nil {
		fmt.Printf("Error loading templates: %v\n", err)
		os.Exit(1)
	}

	app := fiber.New(fiber.Config{
		AppName: "private onsen booking",
		Views:   engine,
	})

	app.Static("/static", "./static")

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Add route for partials
	app.Get("/partials/:partial", func(c *fiber.Ctx) error {
		partial := c.Params("partial")
		// For security reasons, sanitize the partial parameter
		partial = strings.ReplaceAll(partial, ".", "")
		partial = strings.ReplaceAll(partial, "/", "")
		partial = strings.ReplaceAll(partial, "\\", "")

		return c.Render("partials/"+partial, fiber.Map{})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		slots := utils.GetTimeSlots()
		return c.Render("base", fiber.Map{
			"Title": "Kwangdi Hotel - Traditional Japanese Ryokan",
			"Slots": slots,
		})
	})

	app.Get("/rooms/preview", func(c *fiber.Ctx) error {
		fmt.Println("Room preview handler called")

		// Get rooms from database
		var rooms []models.Room

		if err := db.Limit(3).Find(&rooms).Error; err != nil {
			fmt.Println("Error fetching rooms:", err)

			// Return simple HTML for error state
			return c.SendString(`
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    <div class="col-span-3 py-12 text-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto text-leaf-light opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                        <h3 class="mt-4 text-xl font-medium text-leaf">Error Loading Rooms</h3>
                        <p class="mt-2 text-leaf-dark">We encountered a problem loading room data.</p>
                    </div>
                </div>
            `)
		}

		fmt.Printf("Found %d rooms\n", len(rooms))

		if len(rooms) == 0 {
			// Return HTML for empty state
			return c.SendString(`
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    <div class="col-span-3 py-12 text-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto text-leaf-light opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                        <h3 class="mt-4 text-xl font-medium text-leaf">No Rooms Available</h3>
                        <p class="mt-2 text-leaf-dark">No rooms match your current selection criteria.</p>
                    </div>
                </div>
            `)
		}

		// If debug parameter is specified, return simple HTML
		if c.Query("debug") == "true" {
			html := `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">`

			for _, room := range rooms {
				html += fmt.Sprintf(`
					<div class="bg-white rounded-lg shadow-md overflow-hidden">
							<img src="%s" alt="%s" class="w-full h-48 object-cover">
							<div class="p-4">
								<h3 class="text-xl font-bold text-leaf">%s</h3>
								<p class="text-sm">%d guests</p>
								<div class="mt-4 flex justify-between items-center">
									<span class="text-lg font-bold text-leaf">¥%d</span>
									<a href="/booking?room_id=%d" class="bg-leaf text-white px-4 py-2 rounded-md text-sm">Book Now</a>
								</div>
							</div>
						</div>
                `, room.ImageURL, room.Type, room.RoomNo, room.Capacity, int(room.PricePerNight), room.ID)
			}

			html += `</div>`
			return c.SendString(html)
		}

		// Generate inline HTML rather than using template
		html := `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">`

		for _, room := range rooms {
			// Split amenities
			var amenities []string
			if room.Amenities != "" {
				amenities = strings.Split(room.Amenities, ",")
			}

			// Build amenities HTML
			amenitiesHTML := ""
			for _, amenity := range amenities {
				amenitiesHTML += fmt.Sprintf(`
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-cream-light text-leaf-dark">
                        %s
                    </span>
                `, strings.TrimSpace(amenity))
			}

			html += fmt.Sprintf(`
                <div class="bg-white overflow-hidden rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300">
                    <div class="relative">
                        <img src="%s" alt="%s Room" class="w-full h-56 object-cover">
                        <div class="absolute top-0 right-0 mt-4 mr-4">
                            <span class="px-3 py-1 bg-cream-dark bg-opacity-90 text-leaf rounded-full text-xs font-medium">
                                %s Room
                            </span>
                        </div>
                    </div>
                    
                    <div class="p-6">
                        <!-- Room Title and Capacity -->
                        <div class="flex justify-between items-start">
                            <h3 class="text-xl font-semibold text-leaf">%s</h3>
                            <div class="flex items-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-leaf-light mr-1" viewBox="0 0 20 20" fill="currentColor">
                                    <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z" />
                                </svg>
                                <span class="text-sm text-leaf-dark font-medium">
                                    %d guests
                                </span>
                            </div>
                        </div>
                        
                        <!-- Room Description -->
                        <p class="mt-2 text-sm text-leaf-dark line-clamp-3">%s</p>
                        
                        <!-- Divider -->
                        <div class="my-4 border-t border-cream-dark"></div>
                        
                        <!-- Amenities -->
                        <div class="mb-4">
                            <h4 class="text-xs uppercase tracking-wider text-leaf-light font-semibold mb-2">Amenities</h4>
                            <div class="flex flex-wrap gap-2">
                                %s
                            </div>
                        </div>
                        
                        <!-- Price and Book Button -->
                        <div class="flex justify-between items-center mt-4">
                            <div class="flex items-baseline">
                                <span class="text-xl font-bold text-leaf">¥%d</span>
                                <span class="text-sm text-leaf-dark ml-1">/night</span>
                            </div>
                            <a href="/booking?room_id=%d" 
                               class="bg-leaf hover:bg-leaf-dark text-white px-4 py-2 rounded-md text-sm font-medium transition-colors duration-300 flex items-center">
                                Book Now
                            </a>
                        </div>
                    </div>
                </div>
            `, room.ImageURL, room.Type, room.Type, room.RoomNo, room.Capacity, room.Description, amenitiesHTML, int(room.PricePerNight), room.ID)
		}

		html += `</div>`
		return c.SendString(html)
	})

	app.Get("/api/rooms/check", func(c *fiber.Ctx) error {
		var count int64
		db.Model(&models.Room{}).Count(&count)

		return c.JSON(fiber.Map{
			"roomCount": count,
			"status":    "success",
		})
	})

	app.Post("/booking", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		date := c.FormValue("date")
		timeSlot := c.FormValue("slot")

		return c.Render("success", fiber.Map{
			"Name": name,
			"Date": date,
			"Time": timeSlot,
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
