package main



import (

    "os"
    "os/signal"
    "github.com/gofiber/fiber/v2"
    "fmt"
    "syscall"
    "strings"
    "github.com/IamMaheshGurung/privateOnsenBooking/utils"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "go.uber.org/zap"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/IamMaheshGurung/privateOnsenBooking/database"
)



func main() {


    _, err := database.ConnectDB()
    if err != nil {
        fmt.Println("Error connecting to database", err)
        return
    }




    database.SeedRooms()

    logger, err  := zap.NewProduction()
    if err != nil {
        fmt.Println("Error initializing logger", err)
    }


    app := fiber.New(fiber.Config{
        AppName: "private onsen booking",
    })


    app.Use(recover.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins:     "*",
        AllowHeaders:     "Origin, Content-Type, Accept",
        AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
    }))

   
    app.Get("/happyworld", func(c *fiber.Ctx) error {
        slots := utils.GetTimeSlots()
        slotStr := strings.Join(slots, ", ") // Joins all time slots with comma and space
        return c.SendString("Available slots are: " + slotStr)
    })
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-c
        logger.Info("Shutting down server...gracefully")
        if err := app.Shutdown(); err != nil {
            logger.Fatal("Error shutting down server", zap.Error(err))
        }
        logger.Info("Server shut down gracefully")
    }()

    logger.Info("Starting server on port 3000")
    if err := app.Listen(":3000"); err != nil {
        logger.Fatal("Error starting server", zap.Error(err))
    }

    logger.Info("Server started on port 3000")

}





