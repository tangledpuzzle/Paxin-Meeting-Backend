package models

import (
	"time"
)


type ProfileService struct {
    ID        uint64       `gorm:"primaryKey"`
    ProfileID    uint64     `gorm:"not null"`
    CreatedAt time.Time  `gorm:"not null"`
    UpdatedAt time.Time  `gorm:"not null"`
    DeletedAt *time.Time `gorm:"index"`
}
