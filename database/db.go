package database

import (
	"github.com/IamMaheshGurung/privateOnsenBooking/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func SeedRooms(db *gorm.DB) error {
	rooms := []models.Room{
		{
			RoomNo:        "Sakura",
			Type:          "Traditional",
			Capacity:      2,
			PricePerNight: 15000,
			Description:   "Traditional Japanese style room with tatami flooring and views of the cherry blossom garden.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set",
			ImageURL:      "/static/images/rooms/room1.jpg",
		},
		{
			RoomNo:        "Fuji",
			Type:          "Premium",
			Capacity:      2,
			PricePerNight: 25000,
			Description:   "Premium room with a private outdoor bath and mountain views.",
			Amenities:     "Wi-Fi,Private Bathroom,Private Outdoor Bath,Air Conditioning,Yukata,Tea Set,Mini Fridge,TV",
			ImageURL:      "/static/images/rooms/room3.jpg",
		},
		{
			RoomNo:        "Koi",
			Type:          "Deluxe",
			Capacity:      3,
			PricePerNight: 18000,
			Description:   "Spacious room overlooking our koi pond garden with both Western and Japanese-style seating.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set,Mini Fridge",
			ImageURL:      "/static/images/rooms/room1.jpg",
		},
		{
			RoomNo:        "Ajisai",
			Type:          "Traditional",
			Capacity:      2,
			PricePerNight: 14000,
			Description:   "Cozy traditional room with a view of our hydrangea garden.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set",
			ImageURL:      "/static/images/rooms/room2.jpg",
		},
		{
			RoomNo:        "Rhindo",
			Type:          "Family",
			Capacity:      4,
			PricePerNight: 30000,
			Description:   "Spacious family room with separate sleeping areas and garden access.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set,Mini Fridge,TV,Extra Futons",
			ImageURL:      "/static/images/rooms/room3.jpg",
		},
		{
			RoomNo:        "Yuki",
			Type:          "Premium",
			Capacity:      2,
			PricePerNight: 28000,
			Description:   "Premium corner room with panoramic views and a private veranda.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set,Mini Fridge,TV,Veranda",
			ImageURL:      "/static/images/rooms/room1.jpg",
		},
		{
			RoomNo:        "Kiku",
			Type:          "Traditional",
			Capacity:      2,
			PricePerNight: 16000,
			Description:   "Traditional room with authentic decor and chrysanthemum garden views.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set",
			ImageURL:      "/static/images/rooms/room3.jpg",
		},
		{
			RoomNo:        "Matsu",
			Type:          "Deluxe",
			Capacity:      3,
			PricePerNight: 20000,
			Description:   "Deluxe room with a pine tree garden view and upgraded amenities.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set,Mini Fridge,TV",
			ImageURL:      "/static/images/rooms/room2.jpg",
		},
		{
			RoomNo:        "Tsubaki",
			Type:          "Traditional",
			Capacity:      2,
			PricePerNight: 15000,
			Description:   "Traditional room with camellia flower garden views and morning sunlight.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set",
			ImageURL:      "/static/images/rooms/room2.jpg",
		},
		{
			RoomNo:        "Ume",
			Type:          "Family",
			Capacity:      5,
			PricePerNight: 32000,
			Description:   "Our largest family room with plum blossom garden views and sitting area.",
			Amenities:     "Wi-Fi,Private Bathroom,Air Conditioning,Yukata,Tea Set,Mini Fridge,TV,Extra Futons",
			ImageURL:      "/static/images/rooms/room1.jpg",
		},
	}

	// Use a transaction to ensure all-or-nothing insertion
	return db.Transaction(func(tx *gorm.DB) error {
		for _, room := range rooms {
			// Check if room already exists by room number
			var existingRoom models.Room
			result := tx.Where("room_no = ?", room.RoomNo).First(&existingRoom)

			if result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					// Room doesn't exist, create it
					if err := tx.Create(&room).Error; err != nil {
						return err
					}
				} else {
					// Some other error occurred
					return result.Error
				}
			} else {
				// Room exists, update its fields
				existingRoom.Type = room.Type
				existingRoom.Capacity = room.Capacity
				existingRoom.PricePerNight = room.PricePerNight
				existingRoom.Description = room.Description
				existingRoom.Amenities = room.Amenities
				existingRoom.ImageURL = room.ImageURL

				if err := tx.Save(&existingRoom).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
