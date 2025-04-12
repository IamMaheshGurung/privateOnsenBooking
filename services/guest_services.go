package services

import (
    "fmt"
    "time"
    "github.com/IamMaheshGurung/privateOnsenBooking/models"
    "github.com/IamMaheshGurung/privateOnsenBooking/database"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type GuestService struct {
    db *gorm.DB
    logger *zap.Logger
}

func NewGuestService(db *gorm.DB, logger *zap.Logger) *GuestService {
    return &GuestService{
        db: db,
        logger: logger,
    }
}

func (gs *GuestService) GuestDetail (guestID uint, name string, email string, checkIn time.Time, checkOut time.Time) {
    guest := models.Guest{
        ID:       guestID,
        Name:     name,
        Email:    email,
        CheckIn:  checkIn,
        CheckOut: checkOut,
    }

    if err := gs.db.Create(&guest).Error; err != nil {
        gs.logger.Error("failed to create guest", zap.Error(err))
    }
}


func (gs *GuestService) GetGuestByID(guestID uint) (models.Guest, error) {
    var guest models.Guest
    if err := gs.db.First(&guest, guestID).Error; err != nil {
        gs.logger.Error("failed to get guest by ID", zap.Error(err))
        return models.Guest{}, fmt.Errorf("failed to get guest by ID: %w", err)
    }
    return guest, nil
}


func (gs *GuestService) GetAllGuests(checkIn time.Time) ([]models.Guest, error) {
    var guests []models.Guest
    if err := gs.db.
        Where("check_in = ?", checkIn).
        Find(&guests).Error; err != nil {

        gs.logger.Error("failed to get all guests", zap.Error(err))
        return nil, fmt.Errorf("failed to get all guests: %w", err)
    }
    return guests, nil
}



func (gs *GuestService) UpdateGuest(guestID uint, name string, email string, checkIn time.Time, checkOut time.Time) error {
    var guest models.Guest
    if err := gs.db.First(&guest, guestID).Error; err != nil {
        gs.logger.Error("failed to get guest by ID", zap.Error(err))
        return fmt.Errorf("failed to get guest by ID: %w", err)
    }

    guest.Name = name
    guest.Email = email
    guest.CheckIn = checkIn
    guest.CheckOut = checkOut

    if err := gs.db.Save(&guest).Error; err != nil {
        gs.logger.Error("failed to update guest", zap.Error(err))
        return fmt.Errorf("failed to update guest: %w", err)
    }
    return nil
}       


