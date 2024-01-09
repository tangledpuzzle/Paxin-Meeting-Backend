package models

import "time"

// CityTranslation represents a translation of a city.
// @swagger:model
type City struct {
	ID           uint              `gorm:"primary_key"`
	Hex          string            `gorm:"not null"`
	UpdatedAt    time.Time         `gorm:"not null"`
	DeletedAt    *time.Time        `gorm:"index"`
	Translations []CityTranslation `gorm:"foreignkey:CityID"`
}

// City represents a city with translations.
// @swagger:model
type CityTranslation struct {
	ID       uint   `json:"id" gorm:"primary_key"`
	CityID   uint   `json:"cityId"` // link id
	Language string `json:"language"`
	Name     string `json:"name"`
}
