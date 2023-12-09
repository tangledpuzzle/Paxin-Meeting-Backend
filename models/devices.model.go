package models

import (
	"time"
)

type DevicesIOS struct {
	ID        uint       `gorm:"primary_key"`
	Device    string     `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:"index"`
}
