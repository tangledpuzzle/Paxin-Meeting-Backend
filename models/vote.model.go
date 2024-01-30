package models

import (
	uuid "github.com/satori/go.uuid"
)

type Vote struct {
	ID     uint64 `gorm:"primaryKey"`
	IsUP   bool
	UserID uuid.UUID `gorm:"foreignKey:ID"`
	BlogID uint64    `gorm:"foreignKey:ID"`
}
