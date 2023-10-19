package models

import "time"

type CityTranslation struct {
	ID       uint `gorm:"primary_key"`
	CityID   uint // Ссылка на идентификатор гильдии
	Language string
	Name     string
}

type City struct {
	ID           uint              `gorm:"primary_key"`
	Hex          string            `gorm:"not null"`
	UpdatedAt    time.Time         `gorm:"not null"`
	DeletedAt    *time.Time        `gorm:"index"`
	Translations []CityTranslation `gorm:"foreignkey:CityID"`
}
