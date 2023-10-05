package models

import (
	"time"
)

type Guilds struct { 
	ID         uint      `gorm:"primary_key"`
	Name   	   string    `gorm:"not null;unique"`
    Hex        string	 `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:"index"`
}
