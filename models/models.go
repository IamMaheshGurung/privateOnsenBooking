package models



import(
    "time"
)


type Guest struct {
    ID       uint      `json:"id" gorm:"primaryKey"`
    Name    string    `json:"name" gorm:"not null"`
    Email   string    `json:"email" gorm:"not null;unique"`
    Phone string   `json:"phone" gorm:"not null"`
    CreatedAt time.Time 
}



type Room struct {
    ID       uint      `json:"id" gorm:"primaryKey"`
    RoomNo   string    `json:"room_no"`
    CreatedAt time.Time 
}



type RoomBooking struct {
    ID       uint      `json:"id" gorm:"primaryKey"`
    GuestID  uint      `json:"guest_id"`
    Guest   Guest   `json:"guest" gorm:"foreignKey:GuestID"`
    RoomID   uint      `json:"room_id"`
    Room    Room    `json:"room" gorm:"foreignKey:RoomID"`
    CheckIn time.Time `json:"check_in" gorm:"not null"`
    CheckOut time.Time `json:"check_out" gorm:"not null"`
    CreatedAt time.Time
}


//for Onsen Bookiing later that Night

type OnsenBooking struct {
    ID      uint `json:"id" gorm:"primaryKey"`
    GuestID uint `json:"guest_id"`
    RoomId uint `json:"room_id" gorm:"not null"`
    TimeSlot string `json:"time_slot" gorm:"not null"`
    CreatedAt  time.Time
}


