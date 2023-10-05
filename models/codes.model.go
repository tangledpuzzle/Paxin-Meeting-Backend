package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Codes struct {
    ID        uint       `gorm:"primary_key"`
    Code      string	 `gorm:"not null"`
    Balance   string 	 `gorm:"not null"`
    UserId    uuid.UUID  `gorm:"not null"`
	Activated bool		 `gorm:"not null"`
    CreatedAt time.Time  `gorm:"not null"`
    Used 	  uint64  	 `gorm:"null"`
    UpdatedAt time.Time  `gorm:"not null"`
    DeletedAt *time.Time `gorm:"index"`
}