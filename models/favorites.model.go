package models

import uuid "github.com/satori/go.uuid"

type Favorite struct {
	ID     uint64    `gorm:"primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid"`
	BlogID uint64
	User   User `gorm:"foreignKey:UserID"`
	Blog   Blog `gorm:"foreignKey:BlogID"`
}
