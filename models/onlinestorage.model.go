package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type OnlineStorage struct {
	ID        uint64           `gorm:"primaryKey"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null"`
	Year      int              `gorm:"not null"`
	Data      []byte           `gorm:"type:json"`
	CreatedAt time.Time        `gorm:"autoCreateTime"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime"`
	DeletedAt *time.Time       `gorm:"index"`
}

type MonthData struct {
	Month string        `gorm:"-"`
    Hours []TimeEntry   `gorm:"type:json"`
	PostsCount int           `gorm:"-" json:"posts_count"`
}
