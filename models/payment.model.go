package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Payments struct {
	ID        uint64     `gorm:"primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	Amount    float64    `gorm:"not null"`
	PaymentId string     `gorm:"not null"`
	Status    string     `gorm:"not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
