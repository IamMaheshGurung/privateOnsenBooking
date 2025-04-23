package database


import(
    "gorm.io/gorm"
    "github.com/IamMaheshGurung/privateOnsenBooking/models"
    "gorm.io/driver/postgres"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
    var err error
    dsn := "host=localhost user=postgres password=Gurung67 dbname=privateonsen sslmode=disable"
    DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // Migrate the schema
    err = DB.AutoMigrate(&models.Guest{}, &models.Room{}, &models.RoomBooking{})
    if err != nil {
        return nil, err
    }
    return DB, nil
}

func SeedRooms(db *gorm.DB) {
    roomNames := []string{"Sakura", "Fuji", "Koi", "Ajisai", "Rhindo", "Yuki", "Kiku", "Matsu", "Tsubaki", "Ume"}
    
    for _, name := range roomNames {
        // Search for the room by RoomNo, and create it if it doesn't exist
        if err := db.FirstOrCreate(&models.Room{}, models.Room{RoomNo: name}).Error; err != nil {
            // Handle error (e.g., log it)
            panic("failed to seed rooms: " + err.Error())
        }
    }
}
