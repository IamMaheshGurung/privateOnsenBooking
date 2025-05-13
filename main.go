package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/IamMaheshGurung/privateOnsenBooking/config"
	"github.com/IamMaheshGurung/privateOnsenBooking/controllers"
	"github.com/IamMaheshGurung/privateOnsenBooking/database"
	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"github.com/IamMaheshGurung/privateOnsenBooking/routes"
	"github.com/IamMaheshGurung/privateOnsenBooking/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"go.uber.org/zap"
)

func init() {
	// Check if the template file exists
	templatePath := "./templates/booking/form.html"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Printf("ERROR: Template file does not exist at path: %s\n", templatePath)
	} else {
		fmt.Printf("SUCCESS: Template file exists at path: %s\n", templatePath)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	config := config.GetConfig()

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		logger.Error("Error connecting to database", zap.Error(err))
		return
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.Room{}, &models.Guest{}, &models.RoomBooking{}); err != nil {
		logger.Error("Error auto-migrating database:", zap.Error(err))
		return
	}

	// Check if we need to seed the database
	var roomCount int64
	db.Model(&models.Room{}).Count(&roomCount)
	if roomCount == 0 {
		logger.Info("No rooms found in database. Seeding initial data...")
		if err := database.SeedRooms(db); err != nil {
			logger.Error("Error seeding database:", zap.Error(err))
			// Continue anyway, as this is not critical
		}
	}

	// Initialize template engine
	engine := html.New("./templates", ".html")
	engine.Reload(true)
	engine.Debug(true)
	// Test template loading
	if err := engine.Load(); err != nil {
		fmt.Printf("ERROR LOADING TEMPLATES: %v\n", err)
		// Exit early if templates can't be loaded
		os.Exit(1)
	}
	if err := engine.AddFile("booking/form", "./templates/booking/form.html"); err != nil {
		log.Fatalf("Error adding booking form template: %v", err)
	}

	// Add template functions
	engine.AddFunc("dict", func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict requires an even number of arguments")
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

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:     "private onsen booking",
		Views:       engine,
		ViewsLayout: "base",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			logger.Error("Request error",
				zap.String("path", c.Path()),
				zap.Error(err))

			// Return JSON error for API requests
			if strings.HasPrefix(c.Path(), "/api") {
				return c.Status(code).JSON(fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			}

			// Return HTML error page for web requests
			return c.Status(code).Render("error", fiber.Map{
				"Title": "Error - Kwangdi Hotel",
				"Error": err.Error(),
				"Code":  code,
			})
		},
	})

	// Setup static file serving
	app.Static("/static", "./static")

	econfig := services.EmailConfig{
		SMTPServer:   config.SMTPHost,
		SMTPPort:     config.SMTPPort,
		SMTPUsername: config.SMTPUser,
		SMTPPassword: config.SMTPPass,
		FromEmail:    config.SMTPFrom,
		FromName:     config.SMTPFromName,
		TemplatesDir: config.SMTPTemplates,
		Environment:  config.Environment,
	}

	// Initialize services
	emailService := services.NewEmailService(logger, econfig) // Configure this with your email settings
	roomBookingService := services.NewRoomBookingService(db, logger, emailService)
	guestService := services.NewGuestService(db, logger)

	// Initialize controllers
	roomController := controllers.NewRoomController(roomBookingService, logger)
	bookingController := controllers.NewBookingController(roomBookingService, guestService, emailService, logger)
	guestController := controllers.NewGuestController(guestService, logger)

	// Setup routes
	routes.SetupRoutes(app, roomController, bookingController, guestController)
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory", zap.Error(err))
		return
	}
	fmt.Printf("Checking if index.html exists: %v\n",
		fileExists(filepath.Join(cwd, "templates", "index.html")))

	// Setup routes

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start server
	logger.Info("Server starting", zap.String("port", port))
	log.Fatal(app.Listen(":" + port))
}
