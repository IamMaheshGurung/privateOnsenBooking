package main



import (

    "os"
    "os/signal"
    "github.com/gofiber/fiber/v2"
    "fmt"
    "github.com/gofiber/template/html/v2"
    "syscall"
    "github.com/IamMaheshGurung/privateOnsenBooking/utils"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "go.uber.org/zap"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/IamMaheshGurung/privateOnsenBooking/database"
)



func main() {


    db, err := database.ConnectDB()
    if err != nil {
        fmt.Println("Error connecting to database", err)
        return
    }




    database.SeedRooms(db)

    logger, err  := zap.NewProduction()
    if err != nil {
        fmt.Println("Error initializing logger", err)
    }
    
    engine := html.New("./templates", ".html")
    

    app := fiber.New(fiber.Config{
        AppName: "private onsen booking",
        Views: engine,

    })
    

    app.Static("/static", "./static")

    app.Use(recover.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins:     "*",
        AllowHeaders:     "Origin, Content-Type, Accept",
        AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
    }))

   
    app.Get("/", func(c *fiber.Ctx) error {
        slots := utils.GetTimeSlots()
        return c.Render("index", fiber.Map{"slots": slots})
    })


    app.Post("/booking", func(c *fiber.Ctx) error {
        name := c.FormValue("name")
        date := c.FormValue("date")
        timeSlot := c.FormValue("slot")
       
        return c.Render("success", fiber.Map{
            "name": name,
            "date": date,
            "time": timeSlot,
        })

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





