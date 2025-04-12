package database


import(
    "gorm.io/gorm"
    "github.com/IamMaheshGurung/privateOnsenBooking/models"
    "gorm.io/driver/postgres"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
    var err error
    dsn := "host=localhost user=postgres password=Gurung67 dbname=privateOnsen sslmode=disable"
    DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // Migrate the schema
    err = DB.AutoMigrate(&models.Guest{}, &models.Room{}, &models.Booking{})
    if err != nil {
        return nil, err
    }
    return DB, nil
}

func SeedRooms() {
    
    roomNames := [] string {"Sakura", "Fuji", "Koi", "Ajisai", "Rhindo", "Yuki", "Kiku", "Matsu", "Tsubaki", "Ume"}
    for _, name := range roomNames {
        DB.FirstOrCreate(&models.Room{}, models.Room{RoomNo: name})
    }

}
