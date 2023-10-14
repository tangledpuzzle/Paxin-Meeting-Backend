package models

import (
	"time"
)

type GuildTranslation struct {
	ID       uint `gorm:"primary_key"`
	GuildID  uint // Ссылка на идентификатор гильдии
	Language string
	Name     string
}

type Guilds struct {
	ID           uint               `gorm:"primary_key"`
	Hex          string             `gorm:"not null"`
	UpdatedAt    time.Time          `gorm:"not null"`
	DeletedAt    *time.Time         `gorm:"index"`
	Translations []GuildTranslation `gorm:"foreignkey:GuildID"`
}
