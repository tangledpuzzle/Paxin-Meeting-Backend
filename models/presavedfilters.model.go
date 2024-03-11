package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Presavedfilters struct {
	ID        uint64            `gorm:"primaryKey"`
	UserID    uuid.UUID         `gorm:"type:uuid;not null"`
	Name      string            `gorm:"not null"`
	CreatedAt time.Time         `gorm:"autoCreateTime"`
	Meta      map[string]string `gorm:"type:jsonb"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`
	DeletedAt *time.Time        `gorm:"index"`
}
